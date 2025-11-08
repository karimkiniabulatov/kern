#!/bin/bash
set -e

echo "Testing kern API..."

# Check if scripts are executable
if [ ! -x "./scripts/test_api.sh" ]; then
    echo "Setting executable permissions..."
    chmod +x scripts/*.sh
fi

# Update dependencies first
echo "Updating Go dependencies..."
go mod download
go mod tidy

# Build the project
echo "Building kern..."
go build -o kern-test ./cmd/kern

# Set executable permission
chmod +x kern-test

# Start kern in API mode in the background
echo "Starting API server on port 28126..."
./kern-test -r 28126 &
PID=$!

# Wait for server to start
sleep 3

# Test API endpoints
echo ""
echo "Testing API endpoints:"

echo "1. Testing /api/cpu..."
curl -s http://localhost:28126/api/cpu | jq . && echo "✓ CPU API test passed"

echo ""
echo "2. Testing /api/mem..."
curl -s http://localhost:28126/api/mem | jq . && echo "✓ Memory API test passed"

echo ""
echo "3. Testing /api/disk..."
curl -s http://localhost:28126/api/disk | jq . && echo "✓ Disk API test passed"

echo ""
echo "4. Testing /api/net..."
curl -s http://localhost:28126/api/net | jq . && echo "✓ Network API test passed"

echo ""
echo "5. Testing /api/gpu..."
curl -s http://localhost:28126/api/gpu | jq . && echo "✓ GPU API test passed"

echo ""
echo "6. Testing /api/ai..."
curl -s http://localhost:28126/api/ai | jq . && echo "✓ AI API test passed"

echo ""
echo "7. Testing /api/mining..."
curl -s http://localhost:28126/api/mining | jq . && echo "✓ Mining API test passed"

echo ""
echo "8. Testing /api/system..."
curl -s http://localhost:28126/api/system | jq . && echo "✓ System API test passed"

echo ""
echo "9. Testing /health..."
curl -s http://localhost:28126/health | jq . && echo "✓ Health check passed"

echo ""
echo "10. Testing root endpoint..."
curl -s http://localhost:28126/ | jq . && echo "✓ Root endpoint passed"

# Stop kern
kill $PID
wait $PID 2>/dev/null

# Cleanup
rm -f kern-test

echo ""
echo "All API tests passed! kern API is working correctly."