#!/bin/bash
set -e

echo "🚀 Starting kern release process..."

# Проверяем, что мы в правильной директории
if [ ! -f "cmd/kern/main.go" ]; then
    echo "❌ Error: Run this script from project root"
    exit 1
fi

# Получаем версию динамически
VERSION=$(grep 'const version' cmd/kern/main.go | awk -F'"' '{print $2}')
echo "Releasing version: $VERSION"

# 1. Сборка бинарников
echo "📦 Step 1: Building binaries..."
./scripts/build-release.sh

# 2. Создаем тег
echo "🏷️ Step 2: Creating git tag..."
git tag -a "v$VERSION" -m "Release v$VERSION"
git push origin "v$VERSION"

# 3. Показываем инструкцию
echo ""
echo "✅ Release preparation complete!"
echo ""
echo "📋 Next steps:"
echo "1. Go to: https://github.com/karimkiniabulatov/kern/releases"
echo "2. Click 'Draft new release'"
echo "3. Select tag 'v$VERSION'"
echo "4. Copy content from RELEASE_NOTES.md"
echo "5. Upload all files from dist/ directory:"
echo ""
ls -la dist/*.tar.gz dist/*.zip
echo ""
echo "📁 Files to upload:"
echo "  kern-$VERSION-linux-amd64.tar.gz"
echo "  kern-$VERSION-linux-arm64.tar.gz"
echo "  kern-$VERSION-macos-amd64.tar.gz"
echo "  kern-$VERSION-macos-arm64.tar.gz"
echo "  kern-$VERSION-windows-amd64.zip"