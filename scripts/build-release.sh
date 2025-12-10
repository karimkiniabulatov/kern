#!/bin/bash
set -e

echo "Building kern - System Monitoring Tool"
# УДАЛИТЬ строку: VERSION="1.2.3" - используем версию из main.go

# Определяем корневую директорию проекта
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
echo "Project root: $PROJECT_ROOT"

# Получаем версию из main.go динамически
VERSION=$(grep 'const version' "$PROJECT_ROOT/cmd/kern/main.go" | awk -F'"' '{print $2}')
echo "Building version: $VERSION"

# Переходим в корневую директорию проекта
cd "$PROJECT_ROOT"

# Проверяем существование main.go
MAIN_GO="./cmd/kern/main.go"
if [ ! -f "$MAIN_GO" ]; then
    echo "Error: main.go not found at $MAIN_GO"
    exit 1
fi

echo "Found main.go at: $MAIN_GO"

# Функция сборки для целевой платформы
build_target() {
    local os=$1
    local arch=$2
    local output_dir=$3
    local output_name=$4
    
    echo "Building for $os/$arch..."
    
    # Создаем целевую директорию
    mkdir -p "$output_dir"
    
    if [ "$os" = "windows" ]; then
        GOOS=$os GOARCH=$arch go build -o "$output_dir/kern.exe" -ldflags="-s -w" ./cmd/kern
        GOOS=$os GOARCH=$arch go build -o "$output_dir/kern-service.exe" -ldflags="-s -w -H=windowsgui" ./cmd/kern
    else
        GOOS=$os GOARCH=$arch go build -o "$output_dir/kern" -ldflags="-s -w" ./cmd/kern
    fi
    
    # Копируем конфиги если есть
    if [ -d "i18n" ]; then
        mkdir -p "$output_dir/config"
        cp i18n/active.*.json "$output_dir/config/" 2>/dev/null || echo "No language files found, continuing..."
    fi
    
    # Создаем README
    create_readme "$os" "$output_dir" "$VERSION"
}

# Функция создания README для платформы
create_readme() {
    local os=$1
    local dir=$2
    local version=$3
    
    cat > "$dir/README.md" << EOF
# kern v$version for $os

## Installation

### Linux
\`\`\`bash
sudo cp kern /usr/local/bin/
sudo chmod +x /usr/local/bin/kern
\`\`\`

### macOS
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

# Service mode (Linux/macOS)
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
rm -rf dist
mkdir -p dist/{linux-amd64,linux-arm64,macos-amd64,macos-arm64,windows-amd64}
mkdir -p build

# Сборка для всех платформ
echo ""
echo "=== Building for All Platforms ==="

build_target linux amd64 dist/linux-amd64 kern
build_target linux arm64 dist/linux-arm64 kern
build_target darwin amd64 dist/macos-amd64 kern
build_target darwin arm64 dist/macos-arm64 kern
build_target windows amd64 dist/windows-amd64 kern

echo "All builds complete!"

# Создаем архивы для всех платформ
echo ""
echo "Creating distribution archives..."
cd dist

# Архивируем все платформы
tar -czf kern-$VERSION-linux-amd64.tar.gz linux-amd64/*
tar -czf kern-$VERSION-linux-arm64.tar.gz linux-arm64/*
tar -czf kern-$VERSION-macos-amd64.tar.gz macos-amd64/*
tar -czf kern-$VERSION-macos-arm64.tar.gz macos-arm64/*
zip -r kern-$VERSION-windows-amd64.zip windows-amd64/*

# Создаем общий README
cat > README.md << EOF
# kern v$VERSION - System Monitoring Tool

## Available Binaries

### Desktop Platforms
- kern-linux-amd64.tar.gz - Linux 64-bit
- kern-linux-arm64.tar.gz - Linux ARM64  
- kern-macos-amd64.tar.gz - macOS Intel
- kern-macos-arm64.tar.gz - macOS Apple Silicon
- kern-windows-amd64.zip - Windows 64-bit

## Quick Start

### Linux/macOS
\`\`\`bash
tar -xzf kern-linux-amd64.tar.gz
cd kern-linux-amd64
chmod +x kern
./kern
\`\`\`

### Windows
1. Extract kern-windows-amd64.zip
2. Run \`install.bat\` as Administrator
3. Or run \`kern.exe\` directly

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

# Возвращаемся в корень проекта
cd "$PROJECT_ROOT"

echo ""
echo "Build complete!"
echo "Files are in: dist/"

echo ""
echo "Distribution archives created:"
echo "  dist/kern-$VERSION-linux-amd64.tar.gz"
echo "  dist/kern-$VERSION-linux-arm64.tar.gz" 
echo "  dist/kern-$VERSION-macos-amd64.tar.gz"
echo "  dist/kern-$VERSION-macos-arm64.tar.gz"
echo "  dist/kern-$VERSION-windows-amd64.zip"

echo ""
echo "Release ready for distribution!"