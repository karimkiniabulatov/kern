#!/bin/bash

set -e

echo "Building kern - System Monitoring Tool"

# Определяем корневую директорию проекта
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
echo "Project root: $PROJECT_ROOT"

# Проверяем существование main.go
MAIN_GO="$PROJECT_ROOT/cmd/kern/main.go"
if [ ! -f "$MAIN_GO" ]; then
    echo "❌ Error: main.go not found at $MAIN_GO"
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

echo "✅ Found main.go at: $MAIN_GO"

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
            echo "❌ Unsupported operating system: $(uname -s)"
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
        echo "❌ Neither wget nor curl available. Please install one of them."
        return 1
    fi
    
    if [ ! -f "$NDK_FILE" ]; then
        echo "❌ Failed to download Android NDK"
        return 1
    fi
    
    echo "Extracting Android NDK..."
    unzip -q "$NDK_FILE"
    rm "$NDK_FILE"
    
    export ANDROID_NDK_HOME="$(pwd)/$NDK_DIR"
    echo "✅ Android NDK setup complete: $ANDROID_NDK_HOME"
}

# Переходим в корневую директорию проекта
cd "$PROJECT_ROOT"

# Создаем директории для сборки
mkdir -p build
cd build

# Сборка для десктопных платформ
echo "=== Building for Desktop Platforms ==="

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o kern-linux-amd64 "$PROJECT_ROOT/cmd/kern"
GOOS=linux GOARCH=arm64 go build -o kern-linux-arm64 "$PROJECT_ROOT/cmd/kern"

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o kern-windows-amd64.exe "$PROJECT_ROOT/cmd/kern"

echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o kern-darwin-amd64 "$PROJECT_ROOT/cmd/kern"
GOOS=darwin GOARCH=arm64 go build -o kern-darwin-arm64 "$PROJECT_ROOT/cmd/kern"

echo "✅ Desktop builds complete!"

# Android сборка
echo ""
echo "=== Building for Android ==="

# Настраиваем Android NDK
setup_android_ndk

if [ $? -ne 0 ] || [ -z "$ANDROID_NDK_HOME" ]; then
    echo "⚠️  Android NDK not available, skipping Android builds"
    echo "You can build Android later with:"
    echo "export ANDROID_NDK_HOME=/path/to/android-ndk"
    echo "cd scripts && ./build-release.sh"
    ANDROID_BUILD_SUCCESS=false
else
    # Настраиваем пути к компиляторам
    NDK_BIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin"
    
    # Проверяем существование компиляторов
    if [ ! -f "$NDK_BIN/aarch64-linux-android21-clang" ]; then
        echo "❌ Android compiler not found at $NDK_BIN/aarch64-linux-android21-clang"
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
        go build -ldflags="-s -w" -o kern-android-arm64 "$PROJECT_ROOT/cmd/kern"
        
        # ARM
        echo "Building for Android ARM..."
        export GOARCH=arm
        export GOARM=7
        export CC="$NDK_BIN/armv7a-linux-androideabi21-clang"
        go build -ldflags="-s -w" -o kern-android-arm "$PROJECT_ROOT/cmd/kern"
        
        # AMD64
        echo "Building for Android x86_64..."
        export GOARCH=amd64
        export CC="$NDK_BIN/x86_64-linux-android21-clang"
        go build -ldflags="-s -w" -o kern-android-amd64 "$PROJECT_ROOT/cmd/kern"
        
        # 386
        echo "Building for Android x86..."
        export GOARCH=386
        export CC="$NDK_BIN/i686-linux-android21-clang"
        go build -ldflags="-s -w" -o kern-android-386 "$PROJECT_ROOT/cmd/kern"
        
        echo "✅ Android builds complete!"
        
        # Создаем структуру пакета для Android
        echo "Creating Android package structure..."
        mkdir -p android-package/lib/arm64-v8a
        mkdir -p android-package/lib/armeabi-v7a
        mkdir -p android-package/lib/x86_64
        mkdir -p android-package/lib/x86
        
        # Копируем бинарники в структуру APK
        cp kern-android-arm64 android-package/lib/arm64-v8a/libkern.so
        cp kern-android-arm android-package/lib/armeabi-v7a/libkern.so
        cp kern-android-amd64 android-package/lib/x86_64/libkern.so
        cp kern-android-386 android-package/lib/x86/libkern.so
        
        # Создаем базовую структуру APK
        mkdir -p android-package/META-INF
        mkdir -p android-package/res
        mkdir -p android-package/assets
        
        # Создаем манифест для Android
        cat > android-package/AndroidManifest.xml << 'EOF'
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
        cat > android-package/assets/kern-wrapper.sh << 'EOF'
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

        chmod +x android-package/assets/kern-wrapper.sh
    fi
fi

echo "Creating distribution files..."
mkdir -p dist
cp kern-* dist/

# Создаем общий README в папке dist
cat > dist/README.md << 'EOF'
# kern v1.2.0 - System Monitoring Tool

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
chmod +x kern-linux-amd64
./kern-linux-amd64

### Windows
kern-windows-amd64.exe

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

# Копируем Android package если он был создан
if [ "$ANDROID_BUILD_SUCCESS" = true ]; then
    cp -r android-package dist/
    
    # Создаем README для Android в папке dist
    cat > dist/ANDROID-README.md << 'EOF'
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

# Возвращаемся в корень проекта для создания архива
cd "$PROJECT_ROOT"

# Создаем архив с релизом
tar -czf kern-v1.2.0-release.tar.gz -C build/dist .

echo ""
echo "🎉 Build complete!"
echo "📁 Files are in: build/dist/"
echo "📦 Archive: kern-v1.2.0-release.tar.gz"

if [ "$ANDROID_BUILD_SUCCESS" = true ]; then
    echo ""
    echo "📱 Android builds successful:"
    echo "  ✅ kern-android-arm64    (ARM64 devices)"
    echo "  ✅ kern-android-arm      (ARM devices)" 
    echo "  ✅ kern-android-amd64    (x86_64 devices)"
    echo "  ✅ kern-android-386      (x86 devices)"
    echo ""
    echo "See build/dist/ANDROID-README.md for Android installation instructions"
else
    echo ""
    echo "⚠️  Android builds were skipped"
    echo "To build for Android, ensure ANDROID_NDK_HOME is set and run again"
fi

echo ""
echo "🚀 Release ready for distribution!"