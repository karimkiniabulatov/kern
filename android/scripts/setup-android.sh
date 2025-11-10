#!/bin/bash
# Script to setup Android build environment

echo "Setting up Android build environment..."

# Install Android SDK
wget https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip
unzip commandlinetools-linux-8512546_latest.zip
mkdir -p android-sdk
mv cmdline-tools android-sdk/

export ANDROID_HOME=$(pwd)/android-sdk
export PATH=$ANDROID_HOME/cmdline-tools/bin:$PATH

# Accept licenses
yes | sdkmanager --licenses

# Install build tools
sdkmanager "build-tools;33.0.0" "platforms;android-33"

echo "Android build environment setup complete!"