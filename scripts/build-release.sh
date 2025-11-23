#!/bin/bash
set -e

echo "Building kern - System Monitoring Tool"
VERSION="1.2.0"

# Определяем корневую директорию проекта
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
echo "Project root: $PROJECT_ROOT"

# Переходим в корневую директорию проекта
cd "$PROJECT_ROOT"

# Проверяем существование main.go
MAIN_GO="./cmd/kern/main.go"
if [ ! -f "$MAIN_GO" ]; then
    echo "Error: main.go not found at $MAIN_GO"
    echo "Project structure should be:"
    echo "  kern/"
    echo "  ├── cmd/"
    echo "  │   └── kern/"
    echo "  │       └── main.go"
    echo "  ├── scripts/"
    echo "  │   └── build-release.sh"
    echo "  └── ..."
    exit 1
fi

echo "Found main.go at: $MAIN_GO"

# Функция для настройки Android NDK
setup_android_ndk() {
    echo "Setting up Android NDK..."
    
    # Проверяем, не установлен ли уже NDK
    if [ -n "$ANDROID_NDK_HOME" ] && [ -d "$ANDROID_NDK_HOME" ]; then
        echo "Using existing Android NDK: $ANDROID_NDK_HOME"
        return 0
    fi
    
    NDK_VERSION="r25b"
    NDK_DIR="android-ndk-${NDK_VERSION}"
    
    # Если NDK уже скачан в текущей директории
    if [ -d "$NDK_DIR" ]; then
        export ANDROID_NDK_HOME="$(pwd)/$NDK_DIR"
        echo "Using local Android NDK: $ANDROID_NDK_HOME"
        return 0
    fi
    
    # Определяем URL для скачивания в зависимости от ОС
    case "$(uname -s)" in
        Linux*)
            NDK_URL="https://dl.google.com/android/repository/android-ndk-${NDK_VERSION}-linux.zip"
            NDK_FILE="android-ndk-${NDK_VERSION}-linux.zip"
            ;;
        Darwin*)
            NDK_URL="https://dl.google.com/android/repository/android-ndk-${NDK_VERSION}-darwin.zip"
            NDK_FILE="android-ndk-${NDK_VERSION}-darwin.zip"
            ;;
        *)
            echo "Unsupported operating system: $(uname -s)"
            echo "Please download Android NDK manually from:"
            echo "https://developer.android.com/ndk/downloads"
            echo "Then set ANDROID_NDK_HOME environment variable"
            return 1
            ;;
    esac
    
    # Скачиваем и распаковываем NDK
    echo "Downloading Android NDK ${NDK_VERSION}..."
    if command -v wget &> /dev/null; then
        wget -q "$NDK_URL" -O "$NDK_FILE"
    elif command -v curl &> /dev/null; then
        curl -L -o "$NDK_FILE" "$NDK_URL"
    else
        echo "Neither wget nor curl available. Please install one of them."
        return 1
    fi
    
    if [ ! -f "$NDK_FILE" ]; then
        echo "Failed to download Android NDK"
        return 1
    fi
    
    echo "Extracting Android NDK..."
    unzip -q "$NDK_FILE"
    rm "$NDK_FILE"
    
    export ANDROID_NDK_HOME="$(pwd)/$NDK_DIR"
    echo "Android NDK setup complete: $ANDROID_NDK_HOME"
}

# Функция сборки для целевой платформы
build_target() {
    local os=$1
    local arch=$2
    local output_dir=$3
    local output_name=$4
    
    echo "Building for $os/$arch..."
    
    if [ "$os" = "windows" ]; then
        GOOS=$os GOARCH=$arch go build -o "$output_dir/kern.exe" -ldflags="-s -w" ./cmd/kern
        GOOS=$os GOARCH=$arch go build -o "$output_dir/kern-service.exe" -ldflags="-s -w -H=windowsgui" ./cmd/kern
    else
        GOOS=$os GOARCH=$arch go build -o "$output_dir/kern" -ldflags="-s -w" ./cmd/kern
    fi
    
    # Копируем конфиги
    mkdir -p "$output_dir/config"
    cp internal/i18n/active.*.json "$output_dir/config/" 2>/dev/null || echo "No language files found, continuing..."
    
    # Создаем README
    create_readme "$os" "$output_dir"
}

# Функция создания README для платформы
create_readme() {
    local os=$1
    local dir=$2
    
    cat > "$dir/README.md" << EOF
# kern v$VERSION for $os

## Installation

### Linux
\`\`\`bash
sudo cp kern /usr/local/bin/
sudo chmod +x /usr/local/bin/kern
\`\`\`

### MacOS
\`\`\`bash
sudo cp kern /usr/local/bin/
# Or use Homebrew (coming soon)
\`\`\`

### Windows
Run \`install.bat\` as Administrator or manually add to PATH.

## Usage

\`\`\`bash
# Interactive monitoring
kern --all

# API server
kern --remote

# Service mode (Linux/Mac)
kern --daemon

# Remote monitoring
kern --api http://server:28126
\`\`\`

## Platform Notes

$([ "$os" = "windows" ] && echo "For best TUI experience, use Windows Terminal or PowerShell." || echo "Full TUI support available.")
EOF

    # Для Windows создаем установочный скрипт
    if [ "$os" = "windows" ]; then
        cat > "$dir/install.bat" << 'EOF'
@echo off
setlocal enabledelayedexpansion

echo Installing kern for Windows...

:: Check admin rights
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Please run as Administrator
    pause
    exit /b 1
)

:: Create installation directory
set INSTALL_DIR="%PROGRAMFILES%\kern"
if not exist %INSTALL_DIR% mkdir %INSTALL_DIR%

:: Copy files
copy "kern.exe" %INSTALL_DIR%
copy "kern-service.exe" %INSTALL_DIR%
if exist config xcopy config %INSTALL_DIR%\config /E /I

:: Add to PATH
for /f "tokens=2*" %%a in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path 2^>nul') do set OLD_PATH=%%b
echo %OLD_PATH% | find /i "%INSTALL_DIR%" > nul
if errorlevel 1 (
    setx Path "%OLD_PATH%;%INSTALL_DIR%" /M
    echo Added to system PATH
)

:: Create start menu shortcut
set START_MENU="%APPDATA%\Microsoft\Windows\Start Menu\Programs\kern"
if not exist %START_MENU% mkdir %START_MENU%

(
echo Set WshShell = CreateObject("WScript.Shell")
echo Set shortcut = WshShell.CreateShortcut("%START_MENU%\kern.lnk")
echo shortcut.TargetPath = "%INSTALL_DIR%\kern.exe"
echo shortcut.WorkingDirectory = "%INSTALL_DIR%"
echo shortcut.Description = "kern System Monitor"
echo shortcut.Save
) > "%TEMP%\create_shortcut.vbs"

cscript //nologo "%TEMP%\create_shortcut.vbs"
del "%TEMP%\create_shortcut.vbs"

echo.
echo Installation complete!
echo.
echo Usage:
echo   kern --help
echo   kern --remote
echo   kern-service --daemon
echo.
pause
EOF
    fi
}

# Создаем директории для сборки
echo "Creating build directories..."
mkdir -p dist/{linux,linux-arm64,macos,windows}
mkdir -p build

# Сборка для десктопных платформ
echo ""
echo "=== Building for Desktop Platforms ==="

build_target linux amd64 dist/linux kern
build_target linux arm64 dist/linux-arm64 kern
build_target darwin amd64 dist/macos kern
build_target windows amd64 dist/windows kern

echo "Desktop builds complete!"

# Android сборка
echo ""
echo "=== Building for Android ==="

# Настраиваем Android NDK
setup_android_ndk

if [ $? -ne 0 ] || [ -z "$ANDROID_NDK_HOME" ]; then
    echo "Android NDK not available, skipping Android builds"
    echo "You can build Android later with:"
    echo "export ANDROID_NDK_HOME=/path/to/android-ndk"
    echo "cd scripts && ./build-release.sh"
    ANDROID_BUILD_SUCCESS=false
else
    # Настраиваем пути к компиляторам
    NDK_BIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin"
    
    # Проверяем существование компиляторов
    if [ ! -f "$NDK_BIN/aarch64-linux-android21-clang" ]; then
        echo "Android compiler not found at $NDK_BIN/aarch64-linux-android21-clang"
        echo "Please check your Android NDK installation"
        ANDROID_BUILD_SUCCESS=false
    else
        echo "Using Android NDK from: $ANDROID_NDK_HOME"
        ANDROID_BUILD_SUCCESS=true
        
        # ARM64
        echo "Building for Android ARM64..."
        export CGO_ENABLED=1
        export GOOS=android
        export GOARCH=arm64
        export CC="$NDK_BIN/aarch64-linux-android21-clang"
        go build -ldflags="-s -w" -o dist/kern-android-arm64 ./cmd/kern
        
        # ARM
        echo "Building for Android ARM..."
        export GOARCH=arm
        export GOARM=7
        export CC="$NDK_BIN/armv7a-linux-androideabi21-clang"
        go build -ldflags="-s -w" -o dist/kern-android-arm ./cmd/kern
        
        # AMD64
        echo "Building for Android x86_64..."
        export GOARCH=amd64
        export CC="$NDK_BIN/x86_64-linux-android21-clang"
        go build -ldflags="-s -w" -o dist/kern-android-amd64 ./cmd/kern
        
        # 386
        echo "Building for Android x86..."
        export GOARCH=386
        export CC="$NDK_BIN/i686-linux-android21-clang"
        go build -ldflags="-s -w" -o dist/kern-android-386 ./cmd/kern
        
        echo "Android builds complete!"
        
        # Создаем структуру пакета для Android
        echo "Creating Android package structure..."
        mkdir -p dist/android-package/lib/arm64-v8a
        mkdir -p dist/android-package/lib/armeabi-v7a
        mkdir -p dist/android-package/lib/x86_64
        mkdir -p dist/android-package/lib/x86
        
        # Копируем бинарники в структуру APK
        cp dist/kern-android-arm64 dist/android-package/lib/arm64-v8a/libkern.so
        cp dist/kern-android-arm dist/android-package/lib/armeabi-v7a/libkern.so
        cp dist/kern-android-amd64 dist/android-package/lib/x86_64/libkern.so
        cp dist/kern-android-386 dist/android-package/lib/x86/libkern.so
        
        # Создаем базовую структуру APK
        mkdir -p dist/android-package/META-INF
        mkdir -p dist/android-package/res
        mkdir -p dist/android-package/assets
        
        # Создаем манифест для Android
        cat > dist/android-package/AndroidManifest.xml << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.kern.monitor"
    android:versionCode="1"
    android:versionName="1.2.0">

    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
    <uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
    <uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />

    <application
        android:allowBackup="true"
        android:icon="@mipmap/ic_launcher"
        android:label="@string/app_name"
        android:theme="@style/AppTheme">
        
        <activity
            android:name=".MainActivity"
            android:label="@string/app_name">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>
EOF

        # Создаем простой wrapper для Android
        cat > dist/android-package/assets/kern-wrapper.sh << 'EOF'
#!/system/bin/sh
# Wrapper script for kern on Android

export PATH=/system/bin:/system/xbin:$PATH
cd /data/local/tmp

# Check if kern binary exists
if [ -f "./kern" ]; then
    chmod +x ./kern
    ./kern "$@"
else
    echo "kern binary not found in /data/local/tmp/"
    echo "Please copy the appropriate kern binary to /data/local/tmp/"
    echo "Available binaries:"
    echo "  kern-android-arm64 (for ARM64 devices)"
    echo "  kern-android-arm (for ARM devices)"
    echo "  kern-android-amd64 (for x86_64 devices)"
    echo "  kern-android-386 (for x86 devices)"
fi
EOF

        chmod +x dist/android-package/assets/kern-wrapper.sh
    fi
fi

# Создаем архивы для десктопных платформ
echo ""
echo "Creating distribution archives..."
cd dist

# Архивируем десктопные платформы
tar -czf kern-$VERSION-linux-amd64.tar.gz linux/*
tar -czf kern-$VERSION-linux-arm64.tar.gz linux-arm64/*
tar -czf kern-$VERSION-macos-amd64.tar.gz macos/*
zip -r kern-$VERSION-windows-amd64.zip windows/*

# Архивируем Android бинарники если они есть
if [ "$ANDROID_BUILD_SUCCESS" = true ]; then
    tar -czf kern-$VERSION-android-binaries.tar.gz \
        kern-android-* \
        android-package/
fi

# Создаем общий README
cat > README.md << EOF
# kern v$VERSION - System Monitoring Tool

## Available Binaries

### Desktop Platforms
- kern-linux-amd64 - Linux 64-bit
- kern-linux-arm64 - Linux ARM64  
- kern-windows-amd64.exe - Windows 64-bit
- kern-darwin-amd64 - macOS Intel
- kern-darwin-arm64 - macOS Apple Silicon

### Android Platforms
- kern-android-arm64 - Android ARM64 (most devices)
- kern-android-arm - Android ARM
- kern-android-amd64 - Android x86_64
- kern-android-386 - Android x86

## Quick Start

### Linux/macOS
\`\`\`bash
chmod +x kern-linux-amd64
./kern-linux-amd64
\`\`\`

### Windows
\`\`\`bash
kern-windows-amd64.exe
\`\`\`

### Android
See ANDROID-README.md for detailed instructions.

## Features
- CPU, Memory, Disk, Network monitoring
- GPU monitoring (NVIDIA/AMD)
- AI training process detection  
- Mining software monitoring
- Remote API server
- Cross-platform support

## Documentation
Full documentation available at: https://github.com/karimkiniabulatov/kern
EOF

# Создаем README для Android если билд был успешным
if [ "$ANDROID_BUILD_SUCCESS" = true ]; then
    cat > ANDROID-README.md << 'EOF'
# kern for Android

## Installation

### Method 1: Standalone binary (Recommended)
1. Copy the appropriate binary to your Android device:
   - kern-android-arm64 for most modern devices (ARM64)
   - kern-android-arm for older ARM devices
   - kern-android-amd64 for x86_64 devices
   - kern-android-386 for x86 devices

2. Push to device using ADB:
adb push kern-android-arm64 /data/local/tmp/kern
adb shell chmod +x /data/local/tmp/kern
adb shell /data/local/tmp/kern --cpu --mem

3. Or copy via file manager and run in terminal.

### Method 2: Using Termux
1. Install Termux from Google Play Store or F-Droid
2. Copy the binary to Termux home directory
3. Make executable and run:
cp kern-android-arm64 ~/kern
chmod +x ~/kern
./kern --cpu --mem --net

## Usage Examples on Android
# Basic system info
./kern

# CPU and memory only
./kern --cpu --mem

# Network monitoring
./kern --net

# Start API server (requires network permissions)
./kern --remote

## Limitations on Android
- Some system metrics may require root access
- GPU monitoring works only on supported devices with proper drivers
- Network monitoring may be limited by Android permissions
- Disk monitoring shows internal storage only

## Required Permissions
- INTERNET (for remote monitoring)
- ACCESS_NETWORK_STATE (for network stats)
- READ_EXTERNAL_STORAGE (for config files)
- WRITE_EXTERNAL_STORAGE (for logs)

## Troubleshooting
- If you get "permission denied", ensure binary is executable: chmod +x kern
- For full system access, root may be required
- Some features may not work on all Android versions
EOF
fi

# Возвращаемся в корень проекта
cd "$PROJECT_ROOT"

echo ""
echo "Build complete!"
echo "Files are in: dist/"

if [ "$ANDROID_BUILD_SUCCESS" = true ]; then
    echo ""
    echo "Android builds successful:"
    echo "  kern-android-arm64    (ARM64 devices)"
    echo "  kern-android-arm      (ARM devices)" 
    echo "  kern-android-amd64    (x86_64 devices)"
    echo "  kern-android-386      (x86 devices)"
    echo ""
    echo "See dist/ANDROID-README.md for Android installation instructions"
else
    echo ""
    echo "Android builds were skipped"
    echo "To build for Android, ensure ANDROID_NDK_HOME is set and run again"
fi

echo ""
echo "Distribution archives created:"
echo "  dist/kern-$VERSION-linux-amd64.tar.gz"
echo "  dist/kern-$VERSION-linux-arm64.tar.gz" 
echo "  dist/kern-$VERSION-macos-amd64.tar.gz"
echo "  dist/kern-$VERSION-windows-amd64.zip"
[ "$ANDROID_BUILD_SUCCESS" = true ] && echo "  dist/kern-$VERSION-android-binaries.tar.gz"

echo ""
echo "Release ready for distribution!"