kern - Advanced System Monitoring Tool

![version](https://img.shields.io/badge/version-1.2.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![license](https://img.shields.io/badge/license-GPLv3-green)
![platform](https://img.shields.io/badge/platform-Linux%2520%257C%2520macOS%2520%257C%2520Windows%2520%257C%2520Android-lightgrey)

A comprehensive, real-time system monitoring tool written in Go with beautiful TUI interface, support for 50+ languages, and multi-protocol remote monitoring. Runs on desktop systems and Android via terminal emulators.

## ✨ Features

- **📊 Real-time Monitoring**: Live updates with configurable refresh intervals
- **🎨 Professional TUI**: Terminal User Interface with double buffering (no flickering)
- **🌐 Multi-language**: Support for 50+ languages with auto-download capability
- **💾 Disk Monitoring**: Filesystem usage with type detection (SSD/HDD/NVMe/RAID)
- **⚡ CPU Monitoring**: Core/thread information with detailed core usage
- **🧠 Memory Monitoring**: RAM and swap usage with visual graphs
- **🌐 Network Monitoring**: Interface status, speeds, and activity indicators
- **🎮 GPU Monitoring**: NVIDIA/AMD GPU temperature, usage, memory, and power stats
- **🤖 AI Training Monitoring**: Framework detection, VRAM usage, training metrics
- **⛏️ Mining Monitoring**: Cryptocurrency mining detection, hashrate, efficiency
- **🔌 Multi-protocol API**: HTTP, HTTPS, SSH, Telnet support for remote monitoring
- **🌍 Global Access**: Monitor local and remote systems across networks
- **🚀 Auto-daemon**: Background API server starts automatically on installation
- **⌨️ Interactive**: Press 'q' or ESC to quit gracefully
- **🔧 Cross-platform**: Works on Linux, macOS, Windows, Android (terminal emulators)
- **📱 Resize Support**: Automatic redraw on terminal resize

-------------------------------------------------------------------------------------------

## 🚀 Quick Start

### Installation

**Quick install (recommended):**

curl -sSL https://raw.githubusercontent.com/karimkiniabulatov/kern/main/scripts/install.sh | bash

Or clone and build:

git clone https://github.com/karimkiniabulatov/kern
chmod +x ./scripts/install.sh
cd kern
./scripts/install.sh

Immediate Access

After installation, the API server starts automatically in the background:

Test API immediately (no manual start required!):

curl http://localhost:28126/api/cpu
curl http://localhost:28126/api/mem
curl http://localhost:28126/health

Basic Usage

Interactive TUI with all system information:

kern

Monitor specific components:

kern --cpu --mem
kern --gpu --ai
kern --mining

Custom refresh rate:

kern --refresh=5

System language interface:

kern --download-lang zh
kern -l zh

Show version and logo:

kern -v
-------------------------------------------------------------------------------------------

🎯 Core Monitoring Modules

⚡ CPU Monitoring

kern --cpu

📊 Model, cores, threads information

📈 Real-time usage percentage with graphs

📊 Load averages and frequency

🔍 Detailed per-core usage with --detailed flag


🧠 Memory Monitoring

kern --mem

💾 RAM and swap usage with visual indicators

🔄 Available memory tracking

📊 Usage percentages and trends


💾 Disk Monitoring

kern --disk

💽 Filesystem type detection (SSD/HDD/NVMe/RAID)

📁 Mount points and usage percentages

🔧 Smart filtering of system filesystems


🌐 Network Monitoring

kern --net

📡 Interface status and configuration

🚀 Real-time traffic speeds

📊 Activity percentage indicators

🌐 IP and MAC address display


🎮 GPU Monitoring

kern --gpu

🟢 NVIDIA: Full support via nvidia-smi

🔴 AMD: Basic support via rocm-smi

🌡️ Temperature, utilization, VRAM usage

⚡ Power consumption and clock speeds


🤖 AI Training Monitoring

kern --ai

🔍 Detects TensorFlow, PyTorch, HuggingFace processes

💾 VRAM usage tracking for training sessions

📊 Batch size, throughput, loss/accuracy metrics

⏱️ Training time and epoch progress


⛏️ Mining Monitoring

kern --mining

⚙️ Algorithm detection (RandomX, Ethash, SHA-256)

📈 Hashrate and share statistics

💡 Power efficiency calculations

🌡️ Temperature and revenue estimation

-------------------------------------------------------------------------------------------

🌐 Remote Monitoring & API

🤖 Automatic Background API

kern starts an API server automatically on port 28126:

No manual start required - it's always running!:

curl http://localhost:28126/api/cpu
curl http://localhost:28126/api/gpu
curl http://localhost:28126/health

-------------------------------------------------------------------------------------------

🛠️ Service Management

Check daemon status:

kern --service-status

Restart daemon if needed:

kern --restart-service

Enable auto-start on boot:

kern --enable-service

🌍 Remote Access

Monitor remote servers via API:

kern --api http://192.168.1.100:28126
kern --api https://monitoring.example.com:28126

SSH-based monitoring:

kern --api http://server1:28126 --api http://server2:28126

-------------------------------------------------------------------------------------------

📡 API Endpoints
All endpoints return JSON data:

📊 GET /api/cpu - CPU information - {"model": "AMD Ryzen 7", "usage": 23.5}

🧠 GET /api/mem - Memory usage - {"total": "32G", "used": "15.2G"}

💾 GET /api/disk - Disk information - {"filesystem": "/dev/sda1", "usage": 45.2}

🌐 GET /api/net - Network stats - {"interface": "eth0", "speed": "125 KB/s"}

🎮 GET /api/gpu - GPU metrics - {"model": "RTX 4080", "temp": 68.2}

🤖 GET /api/ai - AI training info - {"framework": "PyTorch", "vram": "12.4G"}

⛏️ GET /api/mining - Mining activity - {"algorithm": "Ethash", "hashrate": "85.2 MH/s"}

🔧 GET /api/system - System info - {"version": "1.2.0", "time": "2024-01-15T10:30:00Z"}

💚 GET /health - Health check - {"status": "ok"}

📋 GET / - API information - List of all endpoints

-------------------------------------------------------------------------------------------

🎨 TUI Interface
The professional Terminal User Interface provides:

✨ Flicker-free updates with double buffering

🎨 Color-coded information for quick comprehension

📱 Responsive design that adapts to terminal size

💾 Smart module memory - remembers your preferences

⌨️ Keyboard controls with intuitive navigation

Example TUI Output:

CPU Information
Model: AMD Ryzen 7 5800X
Cores: 8 cores, 16 threads
Usage: ██████████████████ 78.5%
Frequency: 4500 MHz
Load Average: 1.2, 1.5, 1.3

Memory Information
RAM: 15.2G / 32.0G ███████████████████████ 47.5%
Available: 16.8G
Swap: 2.1G / 8.0G ███████████████████████ 26.2%

Press 'q' to quit | Auto-refresh every 2 seconds

📱 Android Support

kern runs on Android devices through terminal emulators like Termux, providing full system monitoring capabilities for remote servers and local network devices.

Install in Termux (recommended terminal emulator):

pkg update && pkg install wget
wget https://github.com/karimkiniabulatov/kern/releases/download/v1.2.0/kern-android-arm64
chmod +x kern-android-arm64
./kern-android-arm64 --api http://your-server:28126

Features on Android:

📡 Monitor remote servers via SSH/HTTP/HTTPS

🔌 Connect to kern API servers on your network

🎨 Full TUI interface with all monitoring modules

⚡ Lightweight and optimized for mobile terminals

👨‍💼 Perfect for sysadmins on the go

Available Android binaries:

📱 kern-android-arm64 - Modern ARM64 devices

📱 kern-android-arm - Older ARM devices

📱 kern-android-amd64 - x86_64 devices

📱 kern-android-386 - x86 devices

-------------------------------------------------------------------------------------------

🌍 Language Support
kern supports 50+ languages with automatic download:

🌎 Available Languages
🇪🇺 European: English, Russian, Spanish, French, German, Italian, Portuguese, Dutch, Polish, Swedish

🇪🇺 Eastern European: Czech, Hungarian, Romanian, Bulgarian, Croatian, Ukrainian, Serbian

🌏 Asian: Japanese, Korean, Chinese, Arabic, Hindi, Indonesian, Vietnamese, Thai, Turkish

🌍 Other: Hebrew, Persian, Urdu, Bengali, Tamil, and more

🔤 Language Management

List all supported languages:

kern --list-languages

Download language pack:

kern --download-lang fr

Download multiple languages:

kern --download-lang es
kern --download-lang de
kern --download-lang ja

-------------------------------------------------------------------------------------------

🛠️ Installation Details

📋 System Requirements

💻 OS: Linux, macOS, Windows, Android (via Termux or other terminal emulators)

⚙️ Architecture: AMD64, ARM64, ARM, x86, x86_64

🔧 Dependencies: Go 1.21+ (auto-installed), basic system tools

📱 Android: Requires terminal emulator (Termux recommended)

🖥️ Platform-Specific Installation
🐧 Linux
Follow the Quick Start instructions above. Works on most distributions including Ubuntu, Debian, CentOS, Arch.

🍎 macOS
Install Go via Homebrew first:

brew install go

Then follow the Quick Start instructions.

🪟 Windows
kern is fully compatible with Windows via Windows Subsystem for Linux (WSL):

Install WSL 2 (if not installed):

wsl --install

Then follow Linux instructions inside WSL environment.

💡 Note: Native Windows support without WSL is experimental and may not provide all features.

📱 Android
Use Termux terminal emulator and follow Android Support instructions above.

-------------------------------------------------------------------------------------------

🔧 Optional Dependencies

🟢 NVIDIA GPU: nvidia-smi (included with drivers)

🔴 AMD GPU: rocm-smi (included with ROCm)

🤖 AI Monitoring: Python with ML frameworks

⛏️ Mining Monitoring: Mining software detection

🔨 Manual Installation

git clone https://github.com/karimkiniabulatov/kern
cd kern
go build -o kern ./cmd/kern
sudo mv kern /usr/local/bin/

Start service (auto-starts on first use):

kern --enable-service

📋 Usage Examples

👨‍💼 System Administration
Complete system overview:

kern --all

High-frequency monitoring for troubleshooting:

kern --cpu --mem --net --refresh=1

Disk space monitoring:

kern --disk --refresh=10

🤖 AI/ML Workloads

GPU and AI training monitoring:

kern --gpu --ai --refresh=2

Detailed performance analysis:

kern --cpu --gpu --ai --mem --refresh=1

⛏️ Mining Operations

Mining performance monitoring:

kern --mining --gpu --refresh=5

Efficiency tracking:

kern --mining --detailed

🌐 Remote Infrastructure

Start local API server (already running by default):

kern --service-status

Monitor remote data center:

kern --api http://dc1-server:28126 --api http://dc2-server:28126

SSH-based remote monitoring:

kern --ssh admin@prod-server --cpu --mem

-------------------------------------------------------------------------------------------

🔧 Service Management


👨‍🔧 Daemon Control

Check status (usually running):

kern --service-status

Restart if needed:

kern --restart-service

Stop daemon (not recommended):

kern --stop-service

Start daemon:

kern --start-service

🔄 Auto-start Management

Enable auto-start on boot (recommended):

kern --enable-service

Disable auto-start:

kern --disable-service

📜 Using Service Script

Alternative service management:

sudo kern-service status
sudo kern-service restart
sudo kern-service logs

-------------------------------------------------------------------------------------------

🧪 Testing & Verification

🔍 Basic Testing
Run test suite:

./scripts/test.sh

Test API functionality:

./scripts/test_api.sh

Verify installation:

kern --service-status
curl http://localhost:28126/health

⚙️ Module Testing

Test individual modules:

kern --cpu --refresh=1
kern --gpu --ai
kern --mining --detailed

Language testing:

kern --download-lang es
kern -l es --cpu --mem

-------------------------------------------------------------------------------------------

🐛 Troubleshooting

❗ Common Issues

API not responding:

Check if daemon is running:

kern --service-status

Restart if needed:

kern --restart-service

Check logs:

sudo kern-service logs

GPU monitoring not working:

Verify NVIDIA drivers:

nvidia-smi

Check AMD ROCm:

rocm-smi

Language not available:

Download language pack:

kern --download-lang LANG_CODE

List available languages:

kern --list-languages

Build issues:

Clean and rebuild:

go clean -modcache
go mod tidy
go build -o kern ./cmd/kern

-------------------------------------------------------------------------------------------

📁 Project Structure

kern/
├── 📄 cmd/kern/main.go          # CLI entry point
├── 📂 internal/
│   ├── ⚙️ config/               # Configuration handling
│   ├── ⚡ cpu/                  # CPU monitoring
│   ├── 💾 disk/                 # Disk monitoring
│   ├── 🧠 mem/                  # Memory monitoring
│   ├── 🌐 net/                  # Network monitoring
│   ├── 🎮 gpu/                  # GPU monitoring
│   ├── 🤖 ai/                   # AI training monitoring
│   ├── ⛏️ mining/               # Mining monitoring
│   ├── 🔧 service/              # Daemon management
│   ├── 🎨 ui/                   # Terminal interface
│   └── 🌐 i18n/                 # Translation system
├── 📚 i18n/                     # Language packs (50+)
├── 🔧 scripts/                  # Management scripts
├── 📖 man/                      # Documentation
└── 📄 README.md

-------------------------------------------------------------------------------------------

🤝 Contributing

We welcome contributions! Please see our contributing guidelines for details.

-------------------------------------------------------------------------------------------

🌐 Adding New Languages

📝 Add translation file to i18n/ directory

🔄 Update supported languages in internal/i18n/translations.go

🧪 Test with: kern -l your_language_code

-------------------------------------------------------------------------------------------

⚙️ Adding Monitoring Modules

📁 Create module in internal/ directory

🔧 Implement Summary() function

🎨 Add TUI rendering in internal/ui/tui.go

🌐 Update API endpoints in cmd/kern/main.go

🌍 Add translations for new module

-------------------------------------------------------------------------------------------

📄 License

This project is licensed under the GNU GPLv3 License - see the LICENSE file for details.

-------------------------------------------------------------------------------------------

💖 Support the Project

kern is developed with ❤️ as an open-source project. If you find it useful and want to support its development, consider making a donation. Your support helps:

🚀 Accelerate new feature development

📱 Improve Android and mobile support

🐛 Faster bug fixes and stability improvements

🌍 Support for more platforms and architectures

📚 Better documentation and tutorials

🔗 Cryptocurrency Donations
Bitcoin (BTC):
1GymM3w4fmbWj6K6dHydgGWYwMckCpHVAn

Ethereum (ERC20):
0x78509af08ce7f9c4a34dd87612d47aceef9b534c

🙌 Other Ways to Support
⭐ Star the repository on GitHub

🐛 Report bugs and issues

💡 Suggest new features

🔧 Contribute code or documentation

📢 Share with your friends and colleagues

🙏 Thank You!
Thank you to all the contributors and users who make kern better every day! Your support, whether through code, donations, or simply using the tool, is greatly appreciated.

-------------------------------------------------------------------------------------------

🆘 Support
📖 Documentation: man kern

🐛 Issues: GitHub Issues

🌐 API Documentation: See above sections

⭐ If you find kern useful, please consider giving it a star on GitHub!

-------------------------------------------------------------------------------------------

Enjoy monitoring your system with kern! 🎯