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
    echo "kern v1.2.0 - System Monitoring Tool"
    echo -e "\033[0m"
}

show_logo

echo "Installing kern system monitor..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing Go..."
    if command -v apt &> /dev/null; then
        sudo apt update
        sudo apt install -y golang-go
    elif command -v yum &> /dev/null; then
        sudo yum install -y golang
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y golang
    elif command -v pacman &> /dev/null; then
        sudo pacman -S go
    else
        echo "Please install Go manually from https://golang.org/dl/"
        exit 1
    fi
fi

# Check required system tools
for tool in df lscpu free ip; do
    if ! command -v $tool &> /dev/null; then
        echo "Installing required tool: $tool"
        if command -v apt &> /dev/null; then
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
    echo "✓ NVIDIA GPU monitoring available"
else
    echo "⚠ nvidia-smi not found (NVIDIA GPU monitoring disabled)"
fi

if command -v rocm-smi &> /dev/null; then
    echo "✓ AMD GPU monitoring available"
else
    echo "⚠ rocm-smi not found (AMD GPU monitoring disabled)"
fi

# Update Go dependencies and install
echo "Building kern..."
cd "$(dirname "$0")/.."
go mod tidy

# Build locally for testing
echo "Building local binary..."
go build -o kern ./cmd/kern

# Install globally
echo "Installing globally..."
go install ./cmd/kern

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