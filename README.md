# kern - Advanced System Monitoring Tool 🔍

![kern Logo](https://img.shields.io/badge/kern-v1.2.3-blue)
![Go Version](https://img.shields.io/badge/Go-1.21+-green)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%20%7C%20macOS-lightgrey)

**kern** is a comprehensive, real-time system monitoring tool with a professional Terminal User Interface (TUI) and REST API. Designed for servers, desktops, and development environments.

## ✨ Features

### 📊 Monitoring Capabilities
- **CPU**: Model, cores, threads, usage %, frequency, load averages
- **Memory**: RAM and swap usage with visual graphs, detailed memory module information
- **Disk**: Filesystem usage, mount points, storage analysis, disk type (SSD/HDD/NVMe)
- **Network**: Interface status, IP/MAC addresses, traffic statistics, connection type (Ethernet/Wi-Fi/Bluetooth/etc.)
- **GPU**: NVIDIA/AMD GPU monitoring (temperature, memory, utilization, power, clock speeds)
- **AI Training**: Framework detection, VRAM usage, training metrics
- **Cryptocurrency Mining**: Algorithm detection, hashrate, efficiency

### 🎯 Advanced Features
- **Smart Defaults**: Remembers your last used monitoring modules
- **Multi-language Support**: 50+ languages with automatic download capability
- **Double Buffered TUI**: Smooth, flicker-free display
- **REST API**: Full HTTP/HTTPS API with CORS support
- **Remote Monitoring**: Monitor multiple servers via API or SSH
- **Daemon Mode**: Background service with auto-start capability
- **Cross-platform**: Native support for Linux, Windows, and macOS

## 🚀 Quick Start

### Installation

#### Linux/macOS
```bash
# Download the appropriate binary for your platform
# Linux AMD64
wget https://github.com/karimkiniabulatov/kern/releases/latest/download/kern-{version}-linux-amd64.tar.gz
tar -xzf kern-{version}-linux-amd64.tar.gz
cd kern-linux-amd64
sudo cp kern /usr/local/bin/
sudo chmod +x /usr/local/bin/kern

# Linux ARM64
wget https://github.com/karimkiniabulatov/kern/releases/latest/download/kern-{version}-linux-arm64.tar.gz

# macOS Intel
wget https://github.com/karimkiniabulatov/kern/releases/latest/download/kern-{version}-macos-amd64.tar.gz

# macOS Apple Silicon
wget https://github.com/karimkiniabulatov/kern/releases/latest/download/kern-{version}-macos-arm64.tar.gz

-----------------------------------------------------------------------------

Windows

Windows
Download kern-{version}-windows-amd64.zip from releases

Extract and run install.bat as Administrator

Or run kern.exe directly

-----------------------------------------------------------------------------

Basic Usage

```text
# Show all system information (smart defaults)
kern

# Show specific modules
kern --cpu --mem
kern --gpu --ai
kern --mining

# Custom refresh rate
kern --refresh=5

# Show detailed CPU core information
kern --detailed

# Show detailed memory module information
kern --detailed-mem

# Show detailed network interface information
kern --detailed-net

# Show with logo
kern --logo

```
-----------------------------------------------------------------------------

📡 Remote Monitoring

```text

# Start on default port (28126)
kern --remote

# Start on custom port
kern --remote-port 8080

```

Monitor Remote Servers

```text

# Via HTTP/HTTPS API
kern --api http://192.168.1.100:28126
kern --api https://monitor.example.com:28126

# Via SSH (requires kern on remote host)
kern --ssh user@hostname

```

API Endpoints

```text

GET /api/cpu - CPU information

GET /api/mem - Memory information

GET /api/disk - Disk information

GET /api/net - Network information

GET /api/gpu - GPU information

GET /api/ai - AI training information

GET /api/mining - Mining information

GET /api/system - System information

GET /health - Health check

```

-----------------------------------------------------------------------------

🔧 Service Management

```text

# Start as daemon service
kern --start-service

# Check service status
kern --service-status

# Stop service
kern --stop-service

# Restart service
kern --restart-service

# Enable auto-start on boot
kern --enable-service

# Disable auto-start
kern --disable-service

# Ensure daemon is running
kern --ensure-running

```
-----------------------------------------------------------------------------

🌐 Language Support

```text

# List supported languages
kern --list-languages

# Download language pack
kern --download-lang ru

# Use specific language
kern --lang ru

```
-----------------------------------------------------------------------------

🏗️ Architecture

```text

kern/
├── cmd/kern/
│   └── main.go                 # CLI entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── cpu/                    # CPU monitoring
│   ├── mem/                    # Memory monitoring  
│   ├── disk/                   # Disk monitoring
│   ├── net/                    # Network monitoring
│   ├── gpu/                    # GPU monitoring
│   ├── ai/                     # AI training monitoring
│   ├── mining/                 # Mining monitoring
│   ├── ui/                     # Terminal UI
│   ├── service/                # Daemon service management
│   └── i18n/                   # Internationalization
├── scripts/
│   └── build-release.sh        # Release building
└── Makefile                    # Build automation

```
-----------------------------------------------------------------------------

📋 Supported Platforms

🐧 Linux
AMD64 (x86_64) - Full support

ARM64 (aarch64) - Full support

ARMv7 (armhf) - Full support

🪟 Windows
AMD64 - Full support

Native service management

Color support in terminal

🍎 macOS
Intel (x86_64) - Full support

Apple Silicon (ARM64) - Full support

🐳 Docker
Multi-architecture images

Lightweight containers

-----------------------------------------------------------------------------

🔨 Development

Build from Source

```text

# Clone repository
git clone https://github.com/karimkiniabulatov/kern
cd kern

# Build for current platform
make build

# Build for all platforms
make release

# Development mode
make dev

```

Testing

```text

# Run all tests
make test

# Test specific module
go test ./internal/cpu/...

```
🤝 Contributing
Fork the repository

Create a feature branch

Make your changes

Add tests

Submit a pull request

-----------------------------------------------------------------------------

📄 License

This project is licensed under the GNU GPLv3 License - see the LICENSE file for details.

-----------------------------------------------------------------------------

💖 Support the Project
kern is developed with ❤️ as an open-source project. Your support helps:

🚀 Accelerate new feature development

🐛 Faster bug fixes and stability improvements

🌍 Support for more platforms and architectures

📚 Better documentation and tutorials

🔗 Cryptocurrency Donations

Bitcoin (BTC): 
```text 
1GymM3w4fmbWj6K6dHydgGWYwMckCpHVAn
```

Ethereum (ERC20): 
```text
0x78509af08ce7f9c4a34d87612d47aceef9b534c
```

🙌 Other Ways to Support
⭐ Star the repository on GitHub

🐛 Report bugs and issues

💡 Suggest new features

🔧 Contribute code or documentation

📢 Share with your friends and colleagues

🙏 Thank You!
Thank you to all the contributors and users who make kern better every day!

🆘 Support
📖 Documentation: man kern

🐛 Issues: GitHub Issues

🌐 API Documentation: See above sections

⭐ If you find kern useful, please consider giving it a star on GitHub!