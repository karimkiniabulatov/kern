#!/bin/bash
set -e

echo "🚀 Starting kern release process..."

# Проверяем, что мы в правильной директории
if [ ! -f "cmd/kern/main.go" ]; then
    echo "❌ Error: Run this script from project root"
    exit 1
fi

# Получаем версию
VERSION=$(grep 'const version' cmd/kern/main.go | awk -F'"' '{print $2}')
echo "Releasing version: $VERSION"

# 1. Сборка бинарников (включая APK)
echo "📦 Step 1: Building binaries and APK..."
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
echo "5. Upload all files from dist/ directory"
echo "6. Click 'Publish release'"
echo ""
echo "📁 Files to upload:"
ls -la dist/
echo ""
echo "📱 Android APK: dist/kern-android.apk"