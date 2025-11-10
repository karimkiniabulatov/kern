#!/bin/bash
set -e

echo "🔨 Building kern binaries for release..."

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

# 📱 Сборка для Android
echo "Building for Android..."

# Проверяем наличие Android NDK
if [ -z "$ANDROID_NDK_HOME" ]; then
    echo "⚠️ ANDROID_NDK_HOME not set. Installing Android NDK..."
    
    # Скачиваем и устанавливаем Android NDK
    wget https://dl.google.com/android/repository/android-ndk-r25b-linux.zip
    unzip android-ndk-r25b-linux.zip
    export ANDROID_NDK_HOME=$(pwd)/android-ndk-r25b
    export PATH=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin:$PATH
fi

# Компиляция для Android
echo "Compiling for Android ARM64..."
GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang go build -ldflags="$LDFLAGS" -o "dist/kern-android-arm64" ./cmd/kern

echo "Compiling for Android ARM..."
GOOS=android GOARCH=arm CGO_ENABLED=1 CC=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi21-clang go build -ldflags="$LDFLAGS" -o "dist/kern-android-arm" ./cmd/kern

# 🎯 Создаем APK файл
echo "Creating APK package..."

# Создаем структуру APK
mkdir -p dist/android-apk/{lib/arm64-v8a,lib/armeabi-v7a,assets}

# Копируем бинарники в соответствующие директории
cp dist/kern-android-arm64 dist/android-apk/lib/arm64-v8a/libkern.so
cp dist/kern-android-arm dist/android-apk/lib/armeabi-v7a/libkern.so

# Создаем базовый AndroidManifest.xml
cat > dist/android-apk/AndroidManifest.xml << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.kern.monitor"
    android:versionCode="1"
    android:versionName="1.2.0">

    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
    <uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
    
    <application
        android:allowBackup="true"
        android:icon="@mipmap/ic_launcher"
        android:label="kern Monitor"
        android:theme="@style/AppTheme">
        
        <activity
            android:name=".MainActivity"
            android:label="kern Monitor"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>
EOF

# Устанавливаем Android SDK tools для создания APK
if ! command -v aapt &> /dev/null; then
    echo "Installing Android SDK tools..."
    wget https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip
    unzip commandlinetools-linux-8512546_latest.zip
    export ANDROID_HOME=$(pwd)/android-sdk
    yes | $ANDROID_HOME/cmdline-tools/bin/sdkmanager --sdk_root=$ANDROID_HOME "build-tools;33.0.0"
fi

# Создаем APK используя aapt и apksigner
cd dist/android-apk

# Создаем базовый APK
$aapt package -f -m -F kern-unsigned.apk -M AndroidManifest.xml -I $ANDROID_HOME/platforms/android-33/android.jar

# Добавляем нативные библиотеки в APK
cd lib
for arch in *; do
    cd $arch
    zip -r ../../kern-unsigned.apk *.so
    cd ..
done
cd ..

# Выравниваем APK
$zipalign -f -p 4 kern-unsigned.apk kern-aligned.apk

# Подписываем APK (используем debug ключ для примера)
if [ ! -f debug.keystore ]; then
    keytool -genkey -v -keystore debug.keystore -alias androiddebugkey -keyalg RSA -keysize 2048 -validity 10000 -storepass android -keypass android -dname "CN=Android Debug,O=Android,C=US"
fi

$apksigner sign --ks debug.keystore --ks-pass pass:android --ks-key-alias androiddebugkey --key-pass pass:android --out kern-android.apk kern-aligned.apk

# Возвращаемся в корневую директорию
cd ../..

# Перемещаем готовый APK в dist
mv dist/android-apk/kern-android.apk dist/kern-android.apk

# 🎯 Создаем чексуммы
echo "Creating checksums..."
cd dist
sha256sum * > sha256sums.txt
cd ..

echo "✅ Build complete! Binaries are in dist/ directory:"
ls -la dist/