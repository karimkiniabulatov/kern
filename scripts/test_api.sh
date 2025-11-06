#!/bin/bash
set -e

echo "Testing kern API..."

# Build the project
echo "Building kern..."
go build -o kern-test ./cmd/kern

# Start kern in API mode in the background
echo "Starting API server on port 8080..."
./kern-test -r 8080 &
PID=$!

# Wait for server to start
sleep 3

# Test API endpoints
echo ""
echo "Testing API endpoints:"

echo "1. Testing /api/cpu..."
curl -s http://localhost:8080/api/cpu | jq . && echo "✓ CPU API test passed"

echo ""
echo "2. Testing /api/mem..."
curl -s http://localhost:8080/api/mem | jq . && echo "✓ Memory API test passed"

echo ""
echo "3. Testing /api/disk..."
curl -s http://localhost:8080/api/disk | jq . && echo "✓ Disk API test passed"

echo ""
echo "4. Testing /api/net..."
curl -s http://localhost:8080/api/net | jq . && echo "✓ Network API test passed"

echo ""
echo "5. Testing /api/system..."
curl -s http://localhost:8080/api/system | jq . && echo "✓ System API test passed"

echo ""
echo "6. Testing /health..."
curl -s http://localhost:8080/health | jq . && echo "✓ Health check passed"

echo ""
echo "7. Testing root endpoint..."
curl -s http://localhost:8080/ | jq . && echo "✓ Root endpoint passed"

# Stop kern
kill $PID
wait $PID 2>/dev/null

# Cleanup
rm -f kern-test

echo ""
echo "All API tests passed! kern API is working correctly."