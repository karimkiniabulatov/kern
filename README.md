# kern - Advanced System Monitoring Tool

A comprehensive, real-time system monitoring tool written in Go with beautiful TUI interface, support for 50+ languages, and multi-protocol remote monitoring.

![kern demo](https://img.shields.io/badge/version-1.2.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-GPLv3-green)

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

### Installation

```bash
# Quick install
go install github.com/karimkiniabulatov/kern@latest

# Or clone and build
git clone https://github.com/karimkiniabulatov/kern
cd kern
chmod +x ./scripts/install.sh
./scripts/install.sh

-----------------------------------------------------------------------------

### Troubleshooting

If you encounter permission issues:

```bash
# Make all scripts executable
chmod +x scripts/*.sh

# Or run the fix permissions script
./scripts/fix-permissions.sh

# If scripts still can't run, try:
bash scripts/install.sh
bash scripts/test.sh

-----------------------------------------------------------------------------

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

# Russian interface
kern -l ru

# Download and use French language
kern --download-lang fr
kern -l fr

# List all supported languages
kern --list-languages

# Start API server on default port 28126
kern -r

# Start API server on custom port
kern -r 26001

# Show logo during monitoring
kern --logo

# Show version and logo
kern -v

Advanced Usage

# GPU and AI training monitoring
kern --gpu --ai

# Mining monitoring with detailed info
kern --mining --detailed

# Combined monitoring for AI workloads
kern --cpu --gpu --ai --refresh=1

# Remote monitoring with all modules on custom port
kern --all -r 26001

🌐 Remote Monitoring
kern supports multiple remote monitoring protocols and access methods:

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

GET /api/cpu - CPU information and usage

GET /api/mem - Memory and swap usage

GET /api/disk - Disk usage and filesystems

GET /api/net - Network interfaces and traffic

GET /api/gpu - GPU information and metrics

GET /api/ai - AI training processes and VRAM

GET /api/mining - Mining activity and efficiency

GET /api/system - System information and version

GET /health - Health check endpoint

GET / - API information and endpoints

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

# Use downloaded language
kern -l fr

# Download multiple languages
kern --download-lang es
kern --download-lang de
kern --download-lang ja

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
│   └── test_api.sh
├── man/                            # Manual page
│   └── kern.1
├── .github/
│   └── workflows/                  # CI/CD
├── go.mod
├── go.sum
└── README.md

🛠 Installation Details
Dependencies
Go 1.21+ (automatically installed if missing)

System tools: df, lscpu, free, ip (installed automatically)

GPU monitoring: nvidia-smi (NVIDIA) or rocm-smi (AMD)

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

🧪 Testing

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

# Test remote monitoring
kern -r 28126 &
curl http://localhost:28126/api/cpu

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

curl http://localhost:28126/api/gpu

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

curl http://localhost:28126/api/ai

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

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Adding New Languages

1. Add translation file to `i18n/` directory
2. Update supported languages list in `internal/i18n/translations.go`
3. Test with: `kern -l your_language_code`

### Adding New Monitoring Modules

1. Create module in `internal/` directory
2. Implement `Summary()` function returning structured data
3. Add TUI rendering in `internal/ui/tui.go`
4. Update API endpoints in `cmd/kern/main.go`
5. Add translations for new module

## 📄 License

This project is licensed under the GNU GPLv3 License - see the LICENSE file for details.

## 🆘 Support

- **Documentation**: `man kern`
- **API Documentation**: See above sections
- **Issues**: [GitHub Issues](https://github.com/karimkiniabulatov/kern/issues)
- **Questions**: Check existing issues or create new one

## 🎯 Technical Highlights

### New TUI Architecture
- **Double Buffering**: Eliminates screen flickering completely
- **Event-Driven**: Responsive to keyboard and resize events
- **Color Management**: Professional color scheme with semantic meaning
- **Cross-Platform**: Consistent experience across Linux, macOS, Windows

### Performance Optimizations
- **Concurrent Data Collection**: All system metrics gathered in parallel
- **Efficient Rendering**: Only changed content is updated
- **Memory Efficient**: Minimal allocations during updates
- **Network Optimized**: Efficient remote data transfer

### Security Features
- **API Security**: Can be protected with reverse proxy and TLS
- **SSH Integration**: Secure remote access via SSH tunnels
- **Access Control**: Firewall and network segmentation support
- **Minimal Footprint**: No persistent data storage required

## 🔄 Version History

### v1.2.0 (Current)
- Added GPU monitoring (NVIDIA/AMD support)
- Added AI training monitoring
- Added cryptocurrency mining monitoring
- Enhanced remote monitoring with multiple protocols
- Improved TUI with better histograms
- Expanded API endpoints

### v1.1.0
- Initial TUI implementation
- Multi-language support
- Basic remote API
- Core monitoring modules

### v1.0.0
- Initial release
- Basic system monitoring
- Command-line interface

---

**Enjoy monitoring your system with kern! 🎯**

*Monitor locally, access globally - kern makes system monitoring accessible everywhere.*