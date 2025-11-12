#!/bin/bash
# Script to setup Android build environment for kern

set -e

echo "Setting up Android build environment for kern..."

# Check if running on Linux (required for Android builds)
if [[ "$(uname)" != "Linux" ]]; then
    echo "Warning: Android NDK builds are best performed on Linux"
    echo "You can still build for other platforms"
fi

# Install required packages
echo "Installing required packages..."
if command -v apt-get &> /dev/null; then
    sudo apt-get update
    sudo apt-get install -y wget unzip git curl
elif command -v yum &> /dev/null; then
    sudo yum install -y wget unzip git curl
elif command -v brew &> /dev/null; then
    brew install wget unzip git curl
fi

# Download and setup Android NDK
echo "Setting up Android NDK..."
NDK_VERSION="r25b"
NDK_DIR="android-ndk-$NDK_VERSION"

if [ ! -d "$NDK_DIR" ]; then
    echo "Downloading Android NDK $NDK_VERSION..."
    
    if [[ "$(uname)" == "Linux" ]]; then
        wget -q https://dl.google.com/android/repository/android-ndk-${NDK_VERSION}-linux.zip
        unzip -q android-ndk-${NDK_VERSION}-linux.zip
        rm android-ndk-${NDK_VERSION}-linux.zip
    elif [[ "$(uname)" == "Darwin" ]]; then
        wget -q https://dl.google.com/android/repository/android-ndk-${NDK_VERSION}-darwin.zip
        unzip -q android-ndk-${NDK_VERSION}-darwin.zip
        rm android-ndk-${NDK_VERSION}-darwin.zip
    else
        echo "Unsupported OS for automatic NDK download"
        echo "Please download Android NDK manually from:"
        echo "https://developer.android.com/ndk/downloads"
        exit 1
    fi
fi

export ANDROID_NDK_HOME="$(pwd)/$NDK_DIR"
export PATH="$ANDROID_NDK_HOME:$PATH"

echo "Android NDK setup complete!"
echo "ANDROID_NDK_HOME set to: $ANDROID_NDK_HOME"

# Verify NDK installation
if [ -f "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang" ]; then
    echo "✅ Android NDK compilers are available"
else
    echo "⚠️  Some Android compilers not found, but setup completed"
    echo "You may need to adjust paths for your specific platform"
fi

echo ""
echo "To build kern for Android, run:"
echo "export ANDROID_NDK_HOME=\"$ANDROID_NDK_HOME\""
echo "cd scripts && ./build-release.sh"
echo ""
echo "For Android APK creation, additional steps are required:"
echo "- Use Android Studio for APK packaging"
echo "- Or build manually with gradle"