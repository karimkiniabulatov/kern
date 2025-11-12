#!/bin/bash

set -e

echo "Building kern - System Monitoring Tool"

# Создаем директории для сборки
mkdir -p build
cd build

# Сборка для различных платформ
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o kern-linux-amd64 ../cmd/kern
GOOS=linux GOARCH=arm64 go build -o kern-linux-arm64 ../cmd/kern

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o kern-windows-amd64.exe ../cmd/kern

echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o kern-darwin-amd64 ../cmd/kern
GOOS=darwin GOARCH=arm64 go build -o kern-darwin-arm64 ../cmd/kern

# Android сборка
echo "Building for Android..."
export CGO_ENABLED=1
export GOOS=android

# ARM64
export GOARCH=arm64
export CC=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang
go build -ldflags="-s -w" -o kern-android-arm64 ../cmd/kern

# ARM
export GOARCH=arm
export GOARM=7
export CC=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi21-clang
go build -ldflags="-s -w" -o kern-android-arm ../cmd/kern

# AMD64
export GOARCH=amd64
export CC=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang
go build -ldflags="-s -w" -o kern-android-amd64 ../cmd/kern

# 386
export GOARCH=386
export CC=$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang
go build -ldflags="-s -w" -o kern-android-386 ../cmd/kern

echo "Creating Android APK package..."
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

echo "Creating distribution archive..."
mkdir -p dist
cp kern-* dist/
cp -r android-package dist/

# Создаем архив с релизом
tar -czf kern-v1.2.0-release.tar.gz -C dist .

echo "Build complete! Files are in build/dist/"
echo "Android package structure created in build/dist/android-package/"