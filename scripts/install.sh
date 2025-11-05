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
            sudo apt install -y procps net-tools util-linux
            break
        elif command -v yum &> /dev/null; then
            sudo yum install -y procps-ng net-tools util-linux
            break
        fi
    fi
done

# Update Go dependencies and install
echo "Building kern..."
go mod tidy
go install

# Check if GOPATH/bin is in PATH
if [[ ":$PATH:" != *":$(go env GOPATH)/bin:"* ]]; then
    echo "Adding GOPATH/bin to PATH..."
    echo "export PATH=\$PATH:$(go env GOPATH)/bin" >> ~/.bashrc
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "kern installed successfully!"
echo "Run 'kern' to start monitoring"
echo "Run 'kern --help' for options"
