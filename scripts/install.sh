#!/bin/bash

set -e

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

# Update Go dependencies and install
echo "Building kern..."
cd "$(dirname "$0")/.."  # Перейти в корень проекта
go mod tidy

# Сначала собираем локально для проверки
echo "Building local binary..."
go build -o kern ./cmd/kern

# Затем устанавливаем глобально
echo "Installing globally..."
go install ./cmd/kern

# Check if GOPATH/bin is in PATH
GOPATH_BIN="$(go env GOPATH)/bin"
if [[ ":$PATH:" != *":$GOPATH_BIN:"* ]]; then
    echo "Adding GOPATH/bin to PATH..."
    echo "export PATH=\$PATH:$GOPATH_BIN" >> ~/.bashrc
    echo "Please run: source ~/.bashrc"
    echo "Or use full path: $GOPATH_BIN/kern"
    export PATH=$PATH:$GOPATH_BIN
fi

echo "kern installed successfully!"
echo "Available commands:"
echo "  ./kern                    # Run local binary"
echo "  kern                      # Run global installation"  
echo "  $GOPATH_BIN/kern         # Run with full path"
echo ""
echo "Run 'kern --help' for options"