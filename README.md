# kern - System Monitoring Tool

A comprehensive, real-time system monitoring tool written in Go with beautiful interface and support for 50+ languages.

![kern demo](https://img.shields.io/badge/version-1.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-GPLv3-green)

## ✨ Features

- **📊 Real-time Monitoring**: Live updates with configurable refresh intervals
- **🎨 Clean Interface**: Simple, line-based output without visual clutter
- **🌐 Multi-language**: Support for 50+ languages with auto-download capability
- **💾 Disk Monitoring**: Filesystem usage with type detection (SSD/HDD/NVMe/RAID)
- **⚡ CPU Monitoring**: Core/thread information with detailed core usage
- **🧠 Memory Monitoring**: RAM and swap usage with visual graphs
- **🌐 Network Monitoring**: Interface status, speeds, and activity indicators
- **🔌 REST API**: HTTP API for remote monitoring on port 28126
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
-----------
Basic Usage
-----------
kern                    # Full system monitoring (default)
kern --cpu              # CPU information only
kern --disk             # Disk information only  
kern --mem              # Memory information only
kern --net              # Network information only

# Monitor everything with custom refresh
kern --refresh=3

# Monitor specific components
kern --cpu --mem --disk

# Russian interface
kern -l ru

# Download and use French language
kern --download-lang fr
kern -l fr

# List all supported languages
kern --list-languages

# Start API server on default port 28126
kern -r 28126

# Show version and logo
kern -v
---------------------------------------------------------
🎨 Interface Design
The clean line-based interface provides stable, flicker-free monitoring:
---------------------------------------------------------
CPU Information
  Model: AMD Ryzen 7 5800X
  Cores: 8 Cores, 16 Threads
  Usage: ███████████████████████ 95.5%
  Frequency: 3800 MHz
  Load Average: 1.25, 1.10, 0.95

Memory Information
  RAM: 15.2G / 32.0G ███████████████████████ 47.5%
  Available: 16.8G
  Swap: 2.1G / 8.0G ███████████████████████ 26.2%

Press 'q' to quit | Auto-refresh every 2 seconds
----------------------------------------------------------
Color Scheme
Headers: Bright yellow

Borders: Blue spaces for clean separation

Parameters: Cyan labels

Graphs: Color-coded by module (orange for CPU, green for memory, etc.)

Values: Appropriate colors for different metrics

🌍 Language Support
kern supports 50+ languages with automatic download capability:

Major Language Groups
European: English, Russian, Spanish, French, German, Italian, Portuguese, Dutch, Polish, Swedish, Danish, Norwegian, Finnish

Eastern European: Czech, Hungarian, Romanian, Bulgarian, Croatian, Slovak, Slovenian, Estonian, Latvian, Lithuanian, Ukrainian, Serbian, Bosnian, Macedonian, Albanian, Greek

Asian: Japanese, Korean, Chinese, Arabic, Hindi, Indonesian, Vietnamese, Thai, Turkish

Other: Hebrew, Persian, Urdu, Bengali, Tamil, Telugu, and more
-------------------
Language Management
-------------------
# List all supported languages
kern --list-languages

# Download a language pack
kern --download-lang fr

# Use downloaded language
kern -l fr
-------------------
Language packs are automatically downloaded from GitHub when requested.

🚀 Remote Monitoring
kern supports multiple remote monitoring methods:
--------------------
--------------------
API-Based Monitoring
--------------------
# Start API server on remote machine
kern -r 8080

# Monitor from client machine
kern --api http://remote-server:8080
--------------------
API Endpoints
GET /api/cpu - CPU information

GET /api/mem - Memory information

GET /api/disk - Disk information

GET /api/net - Network information

GET /api/system - System information

GET /health - Health check
--------------------------------------------
--------------------------------------------
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
----------------------------------------------------------------
🛠 Installation Details
Dependencies
Go 1.21+ (automatically installed if missing)

System tools: df, lscpu, free, ip (installed automatically)
-------------------
Manual Installation
-------------------
# One-line install
curl -sSL https://raw.githubusercontent.com/karimkiniabulatov/kern/main/scripts/install.sh | bash

# Or run manually
./scripts/install.sh

# Build from source
git clone https://github.com/karimkiniabulatov/kern
cd kern
go build -o kern ./cmd/kern
./kern
-------------------
📋 Usage Tips
Interactive Controls
Press 'q': Quit the application

Stable output: Line-based design prevents screen flickering
--------------------
--------------------
Command Line Options
--------------------
Basic Monitoring:
  -a, --all           Show all system information (default)
  -c, --cpu           Show CPU information only
  -d, --disk          Show disk information only
  -m, --mem           Show memory information only
  -n, --net           Show network information only
  -v, --version       Show version information
      --refresh=2     Set refresh interval in seconds
  -l, --lang string   Set interface language
      --detailed      Show detailed CPU core information

Remote Features:
  -r, --remote int    Start remote API server on port
  --ssh host          Monitor remote server via SSH
  --api url           Monitor remote server via API

Language Support:
  --download-lang code Download language pack
  --list-languages    List all supported languages
------------
🧪 Testing🔌
------------
# Run test suite
./scripts/test.sh

# Test API functionality  
./scripts/test_api.sh

# Test language support
kern --list-languages
kern --download-lang es
kern -l es --refresh=1

# Quick verification
go build -o kern ./cmd/kern && ./kern --cpu --refresh=1
-------------------------------------------------------
-------------------------------------------------------
🔌 API Usage
kern provides a REST API for remote monitoring on port 28126:

Start API Server
-------------------
kern -r 28126
--------------------------------------------------------
API Endpoints
GET /api/cpu - CPU information

GET /api/mem - Memory information

GET /api/disk - Disk information

GET /api/net - Network information

GET /api/system - System information

GET /health - Health check

GET / - API information
--------------------------------------------------------
Example API Usage

# Get CPU information
curl http://localhost:28126/api/cpu

# Get memory information
curl http://localhost:28126/api/mem

# Get all system information
curl http://localhost:28126/api/system

API Response Examples
---------------------
# Get CPU information
curl http://localhost:28126/api/cpu

# Response example:
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
-------------------------
🤝 Contributing
Fork the repository

Create a feature branch: git checkout -b feature/amazing-feature

Commit changes: git commit -m 'Add amazing feature'

Push to branch: git push origin feature/amazing-feature

Open a Pull Request

Adding New Languages
Add translation file to i18n/ directory

Update supported languages list in internal/i18n/translations.go

Test with: kern -l your_language_code

📄 License
This project is licensed under the GNU GPLv3 License - see the LICENSE file for details.

🆘 Support
Documentation: man kern

API Documentation: See above section

Issues: GitHub Issues

Questions: Check existing issues or create new one

Enjoy monitoring your system with kern! 🎯