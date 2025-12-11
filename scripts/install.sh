#!/bin/bash
set -e

# Set executable permissions for all scripts
echo "Setting executable permissions for scripts..."
chmod +x scripts/*.sh 2>/dev/null || echo "Warning: Could not set script permissions"

show_logo() {
    echo -e "\033[1;36m"
    cat << "EOF"
 ██╗  ██╗███████╗██████╗ ███╗   ██╗
 ██║ ██╔╝██╔════╝██╔══██╗████╗  ██║
 █████╔╝ █████╗  ██████╔╝██╔██╗ ██║
 ██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║
 ██║  ██╗███████╗██║  ██║██║ ╚████║
 ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
EOF
    echo "kern v1.2.3 - System Monitoring Tool"
    echo -e "\033[0m"
}

show_logo

echo "Installing kern system monitor..."

# Проверка Go версии
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_GO_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]; then
    echo "Go version $GO_VERSION is less than required $REQUIRED_GO_VERSION"
    exit 1
fi

# Check required system tools
for tool in df lscpu free ip; do
    if ! command -v $tool &> /dev/null; then
        echo "Installing required tool: $tool"
        if command -v apt &> /dev/null; then
            sudo apt update
            sudo apt install -y procps net-tools util-linux iproute2
            break
        elif command -v yum &> /dev/null; then
            sudo yum install -y procps-ng net-tools util-linux iproute
            break
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y procps-ng net-tools util-linux iproute
            break
        fi
    fi
done

# Check for GPU monitoring tools
echo "Checking GPU monitoring capabilities..."
if command -v nvidia-smi &> /dev/null; then
    echo "NVIDIA GPU monitoring available"
else
    echo "nvidia-smi not found (NVIDIA GPU monitoring disabled)"
fi

if command -v rocm-smi &> /dev/null; then
    echo "AMD GPU monitoring available"
else
    echo "rocm-smi not found (AMD GPU monitoring disabled)"
fi

# Update Go dependencies and install
echo "Building kern..."
cd "$(dirname "$0")/.."

# Clean up any previous builds
rm -f kern kern-test

# Download all dependencies
echo "Downloading Go dependencies..."
go mod download

# Verify dependencies
echo "Verifying dependencies..."
go mod verify

# Update go.sum
echo "Updating go.sum..."
go mod tidy

# Check for common code issues and fix them
echo "Checking for common code issues..."
if ! grep -q '"strings"' cmd/kern/main.go; then
    echo "Fixing missing strings import in main.go..."
    # Apply the strings import fix
    sed -i '7a\\t"strings"' cmd/kern/main.go
    echo "Fixed strings import"
fi

# Build locally for testing
echo "Building local binary..."
if ! go build -o kern ./cmd/kern; then
    echo "Build failed. Trying alternative build method..."
    # Try building with more verbose output
    go build -x -o kern ./cmd/kern 2>&1 | tail -20
    echo "If build continues to fail, check Go dependencies"
    exit 1
fi

# Install globally
echo "Installing globally..."
go install ./cmd/kern

# Install service management script
echo "Installing service management script..."
if [ -f "scripts/kern-service.sh" ]; then
    sudo cp scripts/kern-service.sh /usr/local/bin/kern-service
    sudo chmod +x /usr/local/bin/kern-service
    echo "Service management script installed"
else
    echo "Service management script not found at scripts/kern-service.sh"
fi

# Create log directory and files
echo "Creating log directory..."
sudo mkdir -p /var/log/kern
sudo touch /var/log/kern.log
sudo chmod 666 /var/log/kern.log
echo "Log directory created at /var/log/kern"

# Install man page with proper permissions
echo "Installing man page..."
if command -v install &> /dev/null; then
    sudo install -g 0 -o 0 -m 0644 man/kern.1 /usr/local/share/man/man1/kern.1 2>/dev/null && \
    sudo mandb >/dev/null 2>&1 && echo "Man page installed successfully" || echo "Man page installation failed - continuing..."
else
    sudo mkdir -p /usr/local/share/man/man1
    sudo cp man/kern.1 /usr/local/share/man/man1/ 2>/dev/null && \
    sudo mandb >/dev/null 2>&1 && echo "Man page installed successfully" || echo "Man page installation failed - continuing..."
fi

# Install i18n files
echo "Installing language files..."
if [ -d "i18n" ]; then
    # Copy to system directory
    sudo mkdir -p /usr/local/share/kern/i18n
    sudo cp i18n/active.*.json /usr/local/share/kern/i18n/ 2>/dev/null && \
    echo "System language files installed successfully" || echo "System language files installation failed - continuing..."
    
    # Copy to user config directory
    mkdir -p ~/.config/kern/i18n
    cp i18n/active.*.json ~/.config/kern/i18n/ 2>/dev/null && \
    echo "User language files installed successfully" || echo "User language files installation failed - continuing..."
else
    echo "No i18n directory found - skipping language files installation"
fi

# Ensure script permissions are set
echo "Ensuring script permissions..."
chmod +x scripts/*.sh

# Check where kern was installed
KERN_PATH="$(go env GOPATH)/bin/kern"
if [ -f "$KERN_PATH" ]; then
    echo "kern v1.2.0 installed successfully at: $KERN_PATH"
    
    # Check if the path is in PATH
    if [[ ":$PATH:" != *":$(go env GOPATH)/bin:"* ]]; then
        echo "Adding GOPATH/bin to PATH..."
        echo "export PATH=\$PATH:$(go env GOPATH)/bin" >> ~/.bashrc
        echo "Please run: source ~/.bashrc"
        echo "Or use: $(go env GOPATH)/bin/kern"
    else
        echo "You can now run: kern"
    fi
else
    echo "Installation failed. You can run locally with: ./kern"
fi

# Configure kern daemon
echo "Configuring kern daemon..."

# Create config directory
mkdir -p ~/.config/kern

# Enable and start kern daemon
echo "Enabling and starting kern daemon..."
if command -v kern &> /dev/null; then
    kern --enable-service 2>/dev/null || echo "Could not enable service automatically"
else
    "$(go env GOPATH)/bin/kern" --enable-service 2>/dev/null || echo "Could not enable service automatically"
fi

# Check daemon status
echo "Checking daemon status..."
if command -v kern &> /dev/null; then
    if kern --service-status 2>/dev/null | grep -q "running.*true"; then
        echo "kern daemon is running and accessible via API"
        echo "API URL: http://localhost:28126"
    else
        echo "Daemon might not be running. Starting manually..."
        kern --start-service 2>/dev/null || echo "Could not start service automatically"
        sleep 2
    fi
else
    echo "kern command not found in PATH, skipping daemon status check"
fi

# Test API connectivity
echo "Testing API connectivity..."
if curl -s http://localhost:28126/health > /dev/null 2>&1; then
    echo "API server is responding correctly"
else
    echo "API server is not responding. Check logs with: kern-service logs"
    echo "You can start the daemon manually with: kern --start-service"
fi

# Re-install service management script to ensure it's available
echo "Ensuring service management script is installed..."
if [ -f "scripts/kern-service.sh" ]; then
    sudo cp scripts/kern-service.sh /usr/local/bin/kern-service
    sudo chmod +x /usr/local/bin/kern-service
    echo "Service management script updated"
fi

echo ""
echo "Installation complete!"
echo ""
echo "Quick start:"
echo "   kern-service status    # Check daemon status"
echo "   curl http://localhost:28126/api/cpu  # Test API"
echo "   kern-service logs      # View logs"
echo ""
echo "API endpoints:"
echo "   http://localhost:28126/api/cpu"
echo "   http://localhost:28126/api/mem" 
echo "   http://localhost:28126/api/disk"
echo "   http://localhost:28126/api/gpu"
echo "   http://localhost:28126/api/ai"
echo "   http://localhost:28126/api/mining"
echo ""
echo "The API server is now running in the background and will"
echo "automatically start on system boot."

# Display service management information
echo ""
echo "Service management commands:"
echo "  kern-service start    - Start daemon"
echo "  kern-service stop     - Stop daemon"
echo "  kern-service status   - Check status"
echo "  kern-service enable   - Enable auto-start"
echo "  kern-service disable  - Disable auto-start"
echo "  kern-service logs     - View logs"
echo ""
echo "Direct kern commands:"
echo "  kern --daemon         - Start as daemon service"
echo "  kern --start-service  - Start the kern daemon"
echo "  kern --stop-service   - Stop the kern daemon"
echo "  kern --status         - Show daemon status"
echo "  kern --enable-service - Enable auto-start on boot"
echo "  kern --disable-service - Disable auto-start on boot"
echo "  kern --ensure-running - Ensure daemon is running"

echo ""
echo "Usage examples:"
echo "  kern                       # Show all system information"
echo "  kern --cpu --mem           # Show only CPU and memory"
echo "  kern --gpu --ai           # Show GPU and AI training info"
echo "  kern --mining             # Show mining information"
echo "  kern --refresh=5 -l ru    # 5 sec refresh with Russian interface"
echo "  kern --detailed           # Show detailed CPU core information"
echo "  kern -r                    # Start API server on port 28126"
echo "  kern -r 26001             # Start API server on custom port"
echo "  kern --logo               # Show logo during monitoring"
echo "  kern --help               # Show help"
echo ""
echo "To view manual: man kern"