#!/bin/bash

echo "Updating kern dependencies..."

# Clean up previous builds
echo "Cleaning previous builds..."
rm -f kern kern-test

# Download all dependencies
echo "Downloading dependencies..."
go mod download

# Verify dependencies
echo "Verifying dependencies..."
go mod verify

# Update go.sum
echo "Updating go.sum..."
go mod tidy

# Build to ensure everything works
echo "Building kern..."
go build -o kern ./cmd/kern

# Set executable permissions
chmod +x kern

echo "Dependencies updated successfully!"
echo "You can now run: ./kern"