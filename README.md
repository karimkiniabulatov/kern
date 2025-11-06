# kern - System Monitoring Tool

A comprehensive, real-time system monitoring tool written in Go with beautiful ANSI-colored interface and support for 50+ languages.

![kern demo](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-GPLv3-green)

## ✨ Features

- **📊 Real-time Monitoring**: Live updates with configurable refresh intervals
- **🎨 Beautiful Interface**: ANSI-colored histograms and stable layout
- **🌐 Multi-language**: Support for 50+ languages with auto-detection
- **💾 Disk Monitoring**: Filesystem usage with type detection (SSD/HDD/NVMe/RAID)
- **⚡ CPU Monitoring**: Core/thread information with grouped visualization
- **🧠 Memory Monitoring**: RAM and swap usage with visual graphs
- **🌐 Network Monitoring**: Interface status, speeds, and activity indicators
- **🔌 REST API**: HTTP API for remote monitoring and integration
- **⌨️ Interactive**: Press 'q' to quit gracefully
- **🔧 Cross-platform**: Works on Linux, Android (Termux), macOS

## 🚀 Quick Start

### Installation

```bash
# Quick install
go install github.com/karimkiniabulatov/kern@latest

# Or clone and build
git clone https://github.com/karimkiniabulatov/kern
cd kern
chmod +x ./scripts/install.sh
./scripts/install.sh

Basic Usage

kern                    # Full system monitoring (default)
kern --cpu              # CPU information only
kern --disk             # Disk information only  
kern --mem              # Memory information only
kern --net              # Network information only

# Monitor everything with custom refresh
kern --refresh=3

# Monitor specific components
kern --cpu --mem --disk

# Russian interface with fast updates
kern -l ru --refresh=1

# Custom language (50+ supported)
kern -l es              # Spanish
kern -l ja              # Japanese
kern -l zh              # Chinese

# Remote API
kern -r 8080            # Start API server on port 8080

# Show version
kern -v

🌐 API Usage
kern provides a REST API for remote monitoring when started with the -r option:

Start API Server

kern -r 8080

API Endpoints
GET /api/cpu - CPU information

GET /api/mem - Memory information

GET /api/disk - Disk information

GET /api/net - Network information

GET /api/system - System information

GET /health - Health check

GET / - API information

Example API Usage

# Get CPU information
curl http://localhost:8080/api/cpu

# Get memory information
curl http://localhost:8080/api/mem

# Get all system information
curl http://localhost:8080/api/system

API Response Example

{
  "model": "AMD Ryzen 7 5800X",
  "vendor": "AuthenticAMD", 
  "architecture": "x86_64",
  "cores": 8,
  "threads": 16,
  "usage": 23.5,
  "frequency": "3800 MHz",
  "load1": 1.2,
  "load5": 1.5,
  "load15": 1.3
}

🎨 Interface Features
Stable Layout
Fixed positioning: Blocks remain in consistent order (Disk → Memory → Network → CPU)

No screen jumping: Data updates in-place without scrolling or layout shifts

Consistent viewing: Focus on changing metrics without visual disruptions

Color Scheme
Disk: 🟠 Orange (used) / ⬜ Gray (free)

CPU: 🟠 Orange (active) / 🔵 Light blue (inactive)

Memory: 🟢 Green (used) / ⬜ Gray (free)

Network: 🟣 Purple (active) / ⬜ Gray (inactive)

Smart Grouping
CPU threads: Logical processors grouped under physical cores

Device types: Automatic detection (SSD/HDD/NVMe/RAID/Virtual)

Network interfaces: Active interfaces with speed calculations

🌍 Supported Languages
kern supports over 50 languages including:

Major Languages: English, Russian, Chinese, Spanish, French, German, Japanese, Korean, Italian, Portuguese, Arabic, Hindi

European: Polish, Dutch, Swedish, Danish, Norwegian, Finnish, Czech, Hungarian, Romanian, Bulgarian, Croatian, Slovak, Slovenian, Estonian, Latvian, Lithuanian, Ukrainian, Serbian, Bosnian, Macedonian, Albanian, Greek

Asian: Indonesian, Vietnamese, Thai, Turkish, Bengali, Tamil, Telugu, Malayalam, Kannada, Gujarati, Marathi, Punjabi, Nepali

Other: Hebrew, Persian, Urdu, Latin, Tibetan, Bashkir, Kazakh, Moldovan, Georgian, Armenian, Belarusian, Mongolian

🛠 Installation Details
Dependencies
Go 1.21+ (automatically installed if missing)

System tools: df, lscpu, free, ip (installed automatically)

Manual Installation

# One-line install
curl -sSL https://raw.githubusercontent.com/karimkiniabulatov/kern/main/scripts/install.sh | bash

# Or run manually
./scripts/install.sh

# Build from source
git clone https://github.com/karimkiniabulatov/kern
cd kern
go build -o kern ./cmd/kern
./kern

📋 Usage Tips
Interactive Controls
Press 'q': Quit the application

Stable output: No terminal flickering or layout changes

Command Line Options
text
  -a, --all           Show all system information (default)
  -c, --cpu           Show CPU information only
  -d, --disk          Show disk information only
  -m, --mem           Show memory information only
  -n, --net           Show network information only
  -v, --version       Show version information
      --refresh=2     Set refresh interval in seconds
  -l, --lang string   Set interface language
  -r, --remote int    Start remote API server on port
      --detailed      Show detailed CPU core information
🧪 Testing
bash
# Run test suite
./scripts/test.sh

# Test API functionality  
./scripts/test_api.sh

# Quick verification
go build -o kern ./cmd/kern && ./kern --cpu --refresh=1

📁 Project Structure

kern/
├── cmd/
│   └── kern/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration handling
│   ├── cpu/
│   │   └── cpu.go                  # CPU monitoring
│   ├── disk/
│   │   └── disk.go                 # Disk monitoring
│   ├── mem/
│   │   └── mem.go                  # Memory monitoring
│   ├── net/
│   │   └── net.go                  # Network monitoring
│   ├── ui/
│   │   └── renderer.go             # Terminal rendering
│   └── i18n/
│       └── translations.go         # Translation system
├── i18n/                           # Language files (50+)
│   ├── active.en.json
│   ├── active.ru.json
│   └── ...
├── scripts/                        # Install/test scripts
│   ├── install.sh
│   ├── test.sh
│   └── test_api.sh
├── man/                            # Manual page
│   └── kern.1
├── .github/
│   └── workflows/                  # CI/CD
├── go.mod
├── go.sum
└── README.md

🔌 API Documentation
CPU Endpoint
GET /api/cpu
Returns detailed CPU information including model, usage, load averages, and core statistics.

Memory Endpoint
GET /api/mem
Returns memory usage information for both RAM and swap space.

Disk Endpoint
GET /api/disk
Returns filesystem usage information for all mounted disks.

Network Endpoint
GET /api/net
Returns network interface statistics including traffic and speeds.

System Endpoint
GET /api/system
Returns general system information and kern version.

🤝 Contributing
Fork the repository

Create a feature branch: git checkout -b feature/amazing-feature

Commit changes: git commit -m 'Add amazing feature'

Push to branch: git push origin feature/amazing-feature

Open a Pull Request

Adding New Languages
Add translation file to i18n/ directory

Update language list in man/kern.1 and README

Test with: kern -l your_language_code

📄 License
This project is licensed under the GNU GPLv3 License - see the LICENSE file for details.

🆘 Support
Documentation: man kern

API Documentation: See above section

Issues: GitHub Issues

Questions: Check existing issues or create new one

Enjoy monitoring your system with kern! 🎯