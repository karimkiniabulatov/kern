#!/bin/bash
# Script to setup Android build environment for kern

set -e

echo "Setting up Android build environment for kern..."

# Check if running on Linux (required for Android builds)
if [[ "$(uname)" != "Linux" ]]; then
    echo "Warning: Android builds are best performed on Linux"
    echo "Some features may not work properly on other systems"
fi

# Install required packages
echo "Installing required packages..."
if command -v apt-get &> /dev/null; then
    sudo apt-get update
    sudo apt-get install -y wget unzip openjdk-11-jdk git
elif command -v yum &> /dev/null; then
    sudo yum install -y wget unzip java-11-openjdk-devel git
elif command -v brew &> /dev/null; then
    brew install wget unzip openjdk@11 git
fi

# Set environment variables
export ANDROID_HOME=${ANDROID_HOME:-$(pwd)/android-sdk}
export PATH=$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$PATH

# Create Android SDK directory
mkdir -p $ANDROID_HOME

# Download and install Android Command Line Tools
if [ ! -d "$ANDROID_HOME/cmdline-tools" ]; then
    echo "Downloading Android Command Line Tools..."
    wget -q https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip
    unzip -q commandlinetools-linux-8512546_latest.zip
    mkdir -p $ANDROID_HOME/cmdline-tools
    mv cmdline-tools $ANDROID_HOME/cmdline-tools/latest
    rm commandlinetools-linux-8512546_latest.zip
fi

# Accept licenses
echo "Accepting Android SDK licenses..."
yes | sdkmanager --licenses > /dev/null 2>&1 || true

# Install Android SDK components
echo "Installing Android SDK components..."
sdkmanager "platform-tools" "platforms;android-33" "build-tools;33.0.0"

# Download Android NDK for native builds
echo "Setting up Android NDK..."
if [ ! -d "$ANDROID_HOME/ndk" ]; then
    wget -q https://dl.google.com/android/repository/android-ndk-r25b-linux.zip
    unzip -q android-ndk-r25b-linux.zip
    mv android-ndk-r25b $ANDROID_HOME/ndk
    rm android-ndk-r25b-linux.zip
fi

export ANDROID_NDK_HOME=$ANDROID_HOME/ndk
export PATH=$ANDROID_NDK_HOME:$PATH

# Verify installation
echo "Verifying Android environment..."
$ANDROID_HOME/platform-tools/adb version
echo "ANDROID_HOME: $ANDROID_HOME"
echo "ANDROID_NDK_HOME: $ANDROID_NDK_HOME"

echo "Android build environment setup complete!"
echo ""
echo "To build kern for Android, run:"
echo "cd scripts && ./build-release.sh"
echo ""
echo "For Android APK creation, you'll need additional tools:"
echo "- Android Studio for final APK packaging"
echo "- Or use gradle build system"