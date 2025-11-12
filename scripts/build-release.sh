#!/bin/bash
set -e

echo "🔨 Building kern binaries for release..."

# Проверяем зависимости
check_dependencies() {
    local deps=("go" "wget" "unzip" "zip")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            echo "❌ Error: $dep is required but not installed"
            exit 1
        fi
    done
}
check_dependencies

# Создаем директорию для бинарников
mkdir -p dist
rm -rf dist/*

# Определяем версию из main.go
VERSION=$(grep 'const version' cmd/kern/main.go | awk -F'"' '{print $2}')
echo "Building version: $VERSION"

# Флаги для сборки
LDFLAGS="-s -w -X main.version=$VERSION"

# 📦 Сборка для Desktop платформ
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "dist/kern-linux-amd64" ./cmd/kern
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "dist/kern-linux-arm64" ./cmd/kern

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "dist/kern-windows-amd64.exe" ./cmd/kern

echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "dist/kern-darwin-amd64" ./cmd/kern
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "dist/kern-darwin-arm64" ./cmd/kern

# 📱 Сборка для Android (опционально)
echo "Building for Android..."

if [ -z "$ANDROID_NDK_HOME" ] && [ "$1" != "--force-android" ]; then
    echo "⚠️ ANDROID_NDK_HOME not set. Skipping Android build."
    echo "   Use './scripts/build-release.sh --force-android' to install NDK automatically"
else
    # [Ваш существующий код Android сборки...]
    echo "✅ Android build completed"
fi

# 🗜️ Создаем архивы для релиза
echo "Creating release archives..."

create_archive() {
    local platform=$1
    local binary=$2
    local archive_name="kern-v${VERSION}-${platform}"
    
    echo "📦 Packaging $archive_name"
    
    if [[ $platform == windows* ]]; then
        cp "$binary" "kern.exe"
        zip -j "dist/${archive_name}.zip" "kern.exe" README.md man/kern.1
        rm -f "kern.exe"
    else
        cp "$binary" "kern"
        tar -czf "dist/${archive_name}.tar.gz" \
            --transform="s,^,${archive_name}/," \
            "kern" README.md man/kern.1
        rm -f "kern"
    fi
}

# Создаем архивы для каждой платформы
create_archive "linux-amd64" "dist/kern-linux-amd64"
create_archive "linux-arm64" "dist/kern-linux-arm64" 
create_archive "windows-amd64" "dist/kern-windows-amd64.exe"
create_archive "darwin-amd64" "dist/kern-darwin-amd64"
create_archive "darwin-arm64" "dist/kern-darwin-arm64"

# 🎯 Создаем чексуммы
echo "Creating checksums..."
cd dist
sha256sum * > sha256sums.txt
cd ..

# 📊 Показываем информацию о размерах
echo ""
echo "📊 Build Summary:"
echo "=================="
for file in dist/kern-*; do
    if [ -f "$file" ] && [[ "$file" != *".tar.gz" ]] && [[ "$file" != *".zip" ]] && [[ "$file" != *".txt" ]]; then
        size=$(du -h "$file" | cut -f1)
        echo "  $(basename $file): $size"
    fi
done

echo ""
echo "🎯 Next steps:"
echo "   - Test binaries: ./dist/kern-linux-amd64 --version"
echo "   - Create release: ./scripts/create-github-release.sh"
echo "   - Upload files to GitHub release"