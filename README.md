# kern - System Monitoring Tool

A comprehensive, real-time system monitoring tool written in Go with beautiful ANSI-colored interface and support for 50+ languages.

![kern demo](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-GPLv3-green)

## ✨ Features

- **📊 Real-time Monitoring**: Live updates with configurable refresh intervals
- **🎨 Beautiful Interface**: ANSI-colored histograms and fixed-position blocks
- **🌐 Multi-language**: Support for 50+ languages with auto-detection
- **💾 Disk Monitoring**: Filesystem usage with type detection (SSD/HDD/NVMe/RAID)
- **⚡ CPU Monitoring**: Core/thread information with grouped visualization
- **🧠 Memory Monitoring**: RAM and swap usage with visual graphs
- **🌐 Network Monitoring**: Interface status, speeds, and activity indicators
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
./scripts/install.sh

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

# Remote API (future feature)
kern -r 26001           # Start API server

# Show version
kern -v

🎨 Interface Features
Fixed Layout
Stable positioning: Blocks remain in fixed order (Disk → Memory → Network → CPU)

No screen jumping: Data updates in-place without scrolling reset

Consistent viewing: Focus on changing metrics without layout shifts

Color Scheme
Disk: 🟠 Pale carrot (used) / 🟢 Green (free)

CPU: 🟠 Orange (active) / 🔵 Light blue (inactive)

Memory: 🟡 Yellow-green (used) / 💚 Pale green (free)

Network: 🟣 Purple (active) / 🩷 Light purple (inactive)

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


# One-line install
curl -sSL https://raw.githubusercontent.com/karimkiniabulatov/kern/main/scripts/install.sh | bash

# Or run manually
./scripts/install.sh

git clone https://github.com/karimkiniabulatov/kern
cd kern
go build -o kern ./cmd/kern
./kern

📋 Usage Tips
Interactive Controls
Press 'q': Quit the application

Fixed blocks: Scroll to view historical data while new data updates in-place

Stable output: No terminal flickering or layout changes

Command Line Options

  -a, --all           Show all system information (default)
  -c, --cpu           Show CPU information only
  -d, --disk          Show disk information only
  -m, --mem           Show memory information only  
  -n, --net           Show network information only
  -v, --version       Show version information
      --refresh=2     Set refresh interval in seconds
  -l, --lang string   Set interface language
  -r, --remote int    Start remote API server on port
  
  🧪 Testing
  
  # Run test suite
./scripts/test.sh

# Quick verification
go build -o kern ./cmd/kern && ./kern --cpu --refresh=1


📁 Project Structure

kern/
├── cmd/kern/main.go          # CLI entry point
├── internal/
│   ├── config/              # Configuration handling
│   ├── cpu/                 # CPU monitoring
│   ├── disk/                # Disk monitoring  
│   ├── mem/                 # Memory monitoring
│   ├── net/                 # Network monitoring
│   └── ui/                  # Terminal rendering
├── i18n/                    # Language files (50+)
├── scripts/                 # Install/test scripts
├── man/                     # Manual page
└── .github/workflows/       # CI/CD

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

Issues: GitHub Issues

Questions: Check existing issues or create new one

Enjoy monitoring your system with kern! 🎯