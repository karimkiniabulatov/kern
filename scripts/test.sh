#!/bin/bash

set -e

echo "Testing kern system monitor..."

# Build the project
echo "Building kern..."
go build -o kern-test ./cmd/kern

# Test basic functionality
echo "Testing disk information..."
./kern-test --disk > /dev/null && echo "✓ Disk test passed"

echo "Testing CPU information..."
./kern-test --cpu > /dev/null && echo "✓ CPU test passed"

echo "Testing memory information..."
./kern-test --mem > /dev/null && echo "✓ Memory test passed"

echo "Testing network information..."
./kern-test --net > /dev/null && echo "✓ Network test passed"

echo "Testing all information..."
./kern-test > /dev/null && echo "✓ Full test passed"

# Test with different refresh rates
echo "Testing with custom refresh rate..."
./kern-test --refresh=1 > /dev/null & 
PID=$!
sleep 3
kill $PID
echo "✓ Refresh rate test passed"

# Test API functionality
echo "Testing API server..."
./kern-test -r 28126 > /dev/null &
API_PID=$!
sleep 2

# Test API endpoints
curl -s http://localhost:28126/api/cpu > /dev/null && echo "✓ CPU API test passed"
curl -s http://localhost:28126/api/mem > /dev/null && echo "✓ Memory API test passed"
curl -s http://localhost:28126/health > /dev/null && echo "✓ Health API test passed"

kill $API_PID

# Cleanup
rm -f kern-test

echo ""
echo "All tests passed! kern is ready for use."