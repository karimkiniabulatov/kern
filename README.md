# kern - Advanced System Monitoring Tool

A comprehensive, real-time system monitoring tool written in Go with beautiful TUI interface, support for 50+ languages, and multi-protocol remote monitoring.

![kern demo](https://img.shields.io/badge/version-1.2.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-GPLv3-green)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

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
- **⌨️ Interactive**: Press 'q' or ESC to quit gracefully
- **🔧 Cross-platform**: Works on Linux, macOS, Windows
- **📱 Resize Support**: Automatic redraw on terminal resize

## 🚀 Quick Start

### Prerequisites

**System Requirements:**
- Linux, macOS, or Windows
- Go 1.21+ (automatically installed if missing)
- Basic system tools: `df`, `lscpu`, `free`, `ip`

**Optional Dependencies:**
- NVIDIA GPUs: `nvidia-smi` (included with NVIDIA drivers)
- AMD GPUs: `rocm-smi` (included with ROCm)
- AI Monitoring: Python with TensorFlow/PyTorch
- Mining Monitoring: Mining software (xmrig, ethminer, etc.)

### Installation

```bash
# Quick install (recommended)
curl -sSL https://raw.githubusercontent.com/karimkiniabulatov/kern/main/scripts/install.sh | bash

# Or clone and build manually
git clone https://github.com/karimkiniabulatov/kern
cd kern
chmod +x ./scripts/install.sh
./scripts/install.sh

# Docker (coming soon)
# docker run -it karimkini/kern:latest

--------------------------------------------------------------------------------------------------

Language Setup

# List all supported languages
kern --list-languages

# Download a language pack (e.g., French)
kern --download-lang fr

# Use the language (must be downloaded first)
kern -l fr

# Check if language is available on your system
kern -l es  # Will show warning if Spanish not downloaded

Basic Usage

kern                    # Show all system information (default)
kern --cpu              # CPU information only
kern --disk             # Disk information only  
kern --mem              # Memory information only
kern --net              # Network information only
kern --gpu              # GPU information only (NVIDIA/AMD)
kern --ai               # AI training information only
kern --mining           # Mining information only

# Monitor everything
kern --all

# Custom refresh rate
kern --refresh=3

# Russian interface (must be downloaded first)
kern -l ru

# Show logo during monitoring
kern --logo

# Show version
kern -v

--------------------------------------------------------------------------------------------------

🎯 Advanced Usage
Smart Module Selection
kern remembers your last used modules:

# First run: shows CPU, Memory, Disk, Network by default
kern

# If you run with specific modules:
kern --gpu --ai

# Next time you run without arguments, it will show GPU and AI
kern  # Shows GPU and AI (your last used modules)

# To reset to defaults, use --all
kern --all

Combined Monitoring

# GPU and AI training monitoring
kern --gpu --ai

# Mining monitoring with detailed info
kern --mining --detailed

# Combined monitoring for AI workloads
kern --cpu --gpu --ai --refresh=1

# System administrator view
kern --cpu --mem --disk --net --refresh=5

--------------------------------------------------------------------------------------------------

🌐 Remote Monitoring
kern supports multiple remote monitoring protocols:

API-Based Monitoring

# Start API server on remote machine (default port 28126)
kern -r

# Start API server on custom port
kern -r 26001

# Monitor remote server via HTTP
kern --api http://192.168.1.100:28126

# Monitor remote server via HTTPS
kern --api https://monitoring.example.com:28126

# Monitor via SSH (requires kern on remote)
kern --ssh user@remote-host

# Multiple remote endpoints
kern --api http://server1:28126 --api http://server2:28126

Access Protocols
HTTP: http://host:28126/api/cpu

HTTPS: https://host:28126/api/cpu (with TLS termination)

SSH: Direct execution or tunneling

Telnet: Via SSH compatibility layer

Network Access
Local Network: http://192.168.1.100:28126

Global Internet: http://your-domain.com:28126 (with port forwarding)

VPN Networks: Access through corporate or private VPNs

Cloud Instances: Monitor AWS, GCP, Azure VMs

API Endpoints
All endpoints return JSON data:

Endpoint	Description	Authentication
GET /api/cpu	CPU information and usage	None
GET /api/mem	Memory and swap usage	None
GET /api/disk	Disk usage and filesystems	None
GET /api/net	Network interfaces and traffic	None
GET /api/gpu	GPU information and metrics	None
GET /api/ai	AI training processes and VRAM	None
GET /api/mining	Mining activity and efficiency	None
GET /api/system	System information and version	None
GET /health	Health check endpoint	None
GET /	API information and endpoints	None

Example API Usage

# Start server on remote machine
kern -r 28126

# Monitor from client (local network)
kern --api http://192.168.1.100:28126

# Monitor from client (internet)
kern --api https://monitoring.example.com:28126

# Direct API calls with curl
curl http://remote-server:28126/api/cpu
curl http://remote-server:28126/api/gpu
curl http://remote-server:28126/health

# SSH tunnel for secure access
ssh -L 28126:localhost:28126 user@remote-host
kern --api http://localhost:28126

SSH Monitoring

# Direct SSH execution
kern --ssh user@remote-host

# SSH with custom command
ssh user@remote-host 'kern --all --refresh=2'

# SSH tunnel + API
ssh -L 28126:localhost:28126 user@remote-host &
kern --api http://localhost:28126

--------------------------------------------------------------------------------------------------

🎨 TUI Interface
The Terminal User Interface provides:

Flicker-free display with double buffering

Professional color scheme with semantic coloring

Responsive design that adapts to terminal size

Smooth updates without screen artifacts

Keyboard controls with intuitive navigation

Example Output:

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

GPU Information
Model: NVIDIA GeForce RTX 4080
Temperature: ██████████████████ 68.2°C
Utilization: ██████████████████ 92.5%
Memory: 12450 MB / 16384 MB
Power: 285 W / 320 W

Press 'q' to quit | Auto-refresh every 2 seconds

--------------------------------------------------------------------------------------------------

🎮 GPU Monitoring
Supported GPUs
NVIDIA: All cards with nvidia-smi support

AMD: Cards with ROCm and rocm-smi support

Intel: Basic detection (experimental)

GPU Metrics
Temperature and fan speed

GPU and memory utilization

VRAM usage and capacity

Power consumption and limits

Core and memory clock speeds

Performance state

Requirements

# For NVIDIA GPUs
sudo apt install nvidia-driver-535 nvidia-utils-535

# For AMD GPUs (ROCm)
sudo apt install rocm-smi-lib

--------------------------------------------------------------------------------------------------

🤖 AI Training Monitoring
Detected Frameworks
TensorFlow and Keras

PyTorch and TorchVision

HuggingFace Transformers

Jupyter notebooks

Custom training scripts

AI Metrics
Active training processes

VRAM usage and allocation

Model name and batch size

Training throughput (samples/sec)

Loss and accuracy metrics

Epoch progress and training time

Example AI Output:

AI Training
Framework: PyTorch
Processes: 2
VRAM: 12450 MB / 16384 MB
Model: resnet-50
Batch Size: 32 | Throughput: 45.7 samples/sec
Epoch: 12 | Loss: 0.234 | Accuracy: 89.2%
Training Time: 2h 15m

--------------------------------------------------------------------------------------------------

⛏️ Mining Monitoring
Supported Algorithms
RandomX: Monero (XMR)

Ethash: Ethereum (ETH) and variants

SHA-256: Bitcoin (BTC)

KAWPOW: Ravencoin (RVN)

BeamHash: Beam (BEAM)

Mining Metrics
Hashrate and algorithm detection

Valid and invalid shares

GPU temperature and power

Mining efficiency (hash/watt)

Uptime and pool information

24-hour revenue estimation

Example Mining Output:

Mining Information
Algorithm: Ethash (Ethereum)
Hashrate: 85.2 MH/s
Shares: 1452/1460 (99.5%)
Temperature: ██████████ 68.5°C
Power: 285 W
Efficiency: 0.30 MH/s/W
Uptime: 3d 12h
24h Revenue: ~$4.25
Pool: ethermine.org

--------------------------------------------------------------------------------------------------

🌍 Language Support
kern supports 50+ languages with automatic download capability:

Major Language Groups
European: English, Russian, Spanish, French, German, Italian, Portuguese, Dutch, Polish, Swedish

Eastern European: Czech, Hungarian, Romanian, Bulgarian, Croatian, Ukrainian, Serbian

Asian: Japanese, Korean, Chinese, Arabic, Hindi, Indonesian, Vietnamese, Thai, Turkish

Other: Hebrew, Persian, Urdu, Bengali, Tamil, and more

Language Management

# List all supported languages
kern --list-languages

# Download a language pack
kern --download-lang fr

# Use downloaded language (must be downloaded first)
kern -l fr

# Download multiple languages
kern --download-lang es
kern --download-lang de  
kern --download-lang ja

# Check current language configuration
cat ~/.config/kern/kern.json | grep language

Important Note About Languages
Language packs must be downloaded before use. If you try to use a language that isn't downloaded, kern will:

Show a warning message

Fall back to English

Suggest the download command

Example:

kern -l fr
# Output: Language 'fr' is not supported. Using English.
#         Use 'kern --download-lang fr' to download language pack.

--------------------------------------------------------------------------------------------------

📁 Project Structure

kern/
├── cmd/
│   └── kern/
│       └── main.go                 # CLI entry point with TUI
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
│   ├── gpu/                        # GPU monitoring
│   │   └── gpu.go
│   ├── ai/                         # AI training monitoring
│   │   └── ai.go
│   ├── mining/                     # Mining monitoring
│   │   └── mining.go
│   ├── ui/
│   │   ├── tui.go                  # Terminal User Interface
│   │   └── renderer.go             # Legacy renderer
│   └── i18n/
│       └── translations.go         # Translation system
├── i18n/                           # Language files (50+)
│   ├── active.en.json
│   ├── active.ru.json
│   └── ...
├── scripts/                        # Install/test scripts
│   ├── install.sh
│   ├── test.sh
│   ├── test_api.sh
│   └── update-deps.sh
├── man/                            # Manual page
│   └── kern.1
├── .github/
│   └── workflows/                  # CI/CD
├── go.mod
├── go.sum
└── README.md

🛠 Installation Details
Dependencies
Required:

Go 1.21+ (automatically installed if missing)

System tools: df, lscpu, free, ip (installed automatically)

Optional:

GPU monitoring: nvidia-smi (NVIDIA) or rocm-smi (AMD)

AI monitoring: Python with ML frameworks

Mining monitoring: Mining software

Manual Installation

# Clone repository
git clone https://github.com/karimkiniabulatov/kern
cd kern

# Make scripts executable
chmod +x scripts/*.sh

# Run installation script
./scripts/install.sh

# Or build manually
go build -o kern ./cmd/kern
sudo cp kern /usr/local/bin/

Docker Installation

# Build from Dockerfile
docker build -t kern .

# Run container
docker run -it --rm kern

# With host networking for full system access
docker run -it --rm --net=host --pid=host kern

--------------------------------------------------------------------------------------------------

📋 Usage Tips
Interactive Controls
q, Q, ESC: Quit the application

Ctrl+C: Quit the application

Terminal Resize: Automatic redraw

Flicker-free: Double buffering for smooth updates

Command Line Options
Basic Monitoring:

-a, --all - Show all system information (default)

-c, --cpu - Show CPU information only

-d, --disk - Show disk information only

-m, --mem - Show memory information only

-n, --net - Show network information only

-g, --gpu - Show GPU information only

--ai - Show AI training information only

--mining - Show mining information only

-v, --version - Show version information

--refresh=2 - Set refresh interval in seconds

-l, --lang string - Set interface language

--detailed - Show detailed CPU core information

Remote Features:

-r, --remote int - Start remote API server on port

--api url - Monitor remote server via API

--ssh host - Monitor remote server via SSH

Language Support:

--download-lang code - Download language pack

--list-languages - List all supported languages

--------------------------------------------------------------------------------------------------

🧪 Testing

# Run complete test suite
./scripts/test.sh

# Test API functionality  
./scripts/test_api.sh

# Test specific modules
kern --cpu --refresh=1
kern --gpu --ai
kern --mining --detailed

# Test language support
kern --list-languages
kern --download-lang es
kern -l es --refresh=1

# Quick verification
go build -o kern ./cmd/kern && ./kern --cpu --refresh=1

# Test remote monitoring
kern -r 28126 &
curl http://localhost:28126/api/cpu
kill %1

--------------------------------------------------------------------------------------------------

🔌 API Usage Examples

Start API Server

kern -r 28126

API Response Examples

CPU Information:

curl http://localhost:28126/api/cpu

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

GPU Information:

{
  "model": "NVIDIA GeForce RTX 4080",
  "driver_version": "535.104.05",
  "gpu_temp": 68.2,
  "memory_total": "16384 MB",
  "memory_used": "12450 MB", 
  "memory_free": "3934 MB",
  "utilization": 92.5,
  "power_draw": "285 W",
  "power_limit": "320 W",
  "fan_speed": 65.0,
  "clock_core": "2500 MHz",
  "clock_memory": "2250 MHz"
}

AI Training Information:

{
  "framework": "PyTorch",
  "process_count": 2,
  "vram_usage": "12450 MB",
  "vram_total": "16384 MB", 
  "model_name": "resnet-50",
  "batch_size": 32,
  "throughput": 45.7,
  "epoch": 12,
  "loss": 0.234,
  "accuracy": 0.892,
  "training_time": "2h 15m"
}

--------------------------------------------------------------------------------------------------

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

Adding New Monitoring Modules
Create module in internal/ directory

Implement Summary() function returning structured data

Add TUI rendering in internal/ui/tui.go

Update API endpoints in cmd/kern/main.go

Add translations for new module

--------------------------------------------------------------------------------------------------

📄 License

This project is licensed under the GNU GPLv3 License - see the LICENSE file for details.

--------------------------------------------------------------------------------------------------

🆘 Support
Documentation: man kern

API Documentation: See above sections

Issues: GitHub Issues

Questions: Check existing issues or create new one

Troubleshooting

Common Issues:

# Permission issues with scripts
chmod +x scripts/*.sh

# Language not working
kern --download-lang LANG_CODE
kern -l LANG_CODE

# GPU monitoring not working
nvidia-smi  # Check if NVIDIA drivers are installed
rocm-smi    # Check if ROCm is installed

# Build issues
./scripts/update-deps.sh
go clean -modcache
go mod tidy

--------------------------------------------------------------------------------------------------

🎯 Technical Highlights
New TUI Architecture
Double Buffering: Eliminates screen flickering completely

Event-Driven: Responsive to keyboard and resize events

Color Management: Professional color scheme with semantic meaning

Cross-Platform: Consistent experience across Linux, macOS, Windows

Performance Optimizations
Concurrent Data Collection: All system metrics gathered in parallel

Efficient Rendering: Only changed content is updated

Memory Efficient: Minimal allocations during updates

Network Optimized: Efficient remote data transfer

Security Features
API Security: Can be protected with reverse proxy and TLS

SSH Integration: Secure remote access via SSH tunnels

Access Control: Firewall and network segmentation support

Minimal Footprint: No persistent data storage required

--------------------------------------------------------------------------------------------------

🔄 Version History
v1.2.0 (Current)
Added GPU monitoring (NVIDIA/AMD support)

Added AI training monitoring

Added cryptocurrency mining monitoring

Enhanced remote monitoring with multiple protocols

Improved TUI with better histograms

Expanded API endpoints

Smart module preferences

v1.1.0
Initial TUI implementation

Multi-language support

Basic remote API

Core monitoring modules

v1.0.0
Initial release

Basic system monitoring

Command-line interface

--------------------------------------------------------------------------------------------------

Enjoy monitoring your system with kern! 🎯