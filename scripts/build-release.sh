# Копируем Android package если он был создан
if [ "$ANDROID_BUILD_SUCCESS" = true ]; then
    cp -r android-package dist/
    
    # Создаем README для Android
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

# Создаем общий README
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

# Создаем архив с релизом
cd ..
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