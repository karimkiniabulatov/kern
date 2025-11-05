markdown
# kern - System Monitoring Tool

A comprehensive system monitoring tool written in Go that provides real-time information about your system's resources.

## Features

- **Disk Monitoring**: Filesystem usage and statistics
- **CPU Monitoring**: Load averages, core information, frequency
- **Memory Monitoring**: RAM and swap usage
- **Network Monitoring**: Interface status and traffic
- **Real-time Updates**: Configurable refresh intervals
- **ANSI Colors**: Colorful terminal output with histograms
- **Cross-platform**: Works on Linux, Android (Termux), and other Unix-like systems
- **Remote API**: Optional remote monitoring capability

## Installation

### Quick Install
```bash
go install github.com/karimkiniabulatov/kern@latest
Automated Installation
bash
curl -sSL https://raw.githubusercontent.com/karimkiniabulatov/kern/main/scripts/install.sh | bash
Manual Installation
bash
git clone https://github.com/karimkiniabulatov/kern
cd kern
./scripts/install.sh
Usage
Basic Monitoring
bash
kern                    # Show all information
kern -d                 # Show only disk information
kern --cpu              # Show only CPU information  
kern -m                 # Show only memory information
kern --net              # Show only network information
Advanced Options
bash
kern --refresh=5        # Update every 5 seconds
kern -l ru              # Russian interface (when translations available)
kern -r 26001           # Start remote API on port 26001
kern -v                 # Show version
Examples
bash
# Monitor everything with 3-second updates
kern --refresh=3

# Monitor only CPU and memory
kern --cpu --mem

# Monitor disk usage with Russian interface
kern -d -l ru
Remote API
kern includes a remote monitoring API:

bash
# Start remote server
kern -r 26001

# Connect from another machine (future feature)
curl http://localhost:26001/api/stats
Platform Support
✅ Linux

✅ Android (via Termux)

✅ macOS (basic support)

⚠️ Windows (limited functionality)

Dependencies
Go 1.21 or higher

Standard Unix tools: df, lscpu, free, ip

The install script will automatically install these dependencies on most systems.

Testing
Run the test suite to verify everything works:

bash
./scripts/test.sh
Building from Source
bash
git clone https://github.com/karimkiniabulatov/kern
cd kern
go build -o kern ./cmd/kern
./kern
Contributing
Fork the repository

Create a feature branch

Make your changes

Add tests if applicable

Submit a pull request

License
GNU GPLv3 - See LICENSE file for details.

Support
For issues and questions:

Open an issue on GitHub

Check the man page: man kern