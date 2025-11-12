#!/bin/bash
set -e

echo "🌐 Testing kern API..."

# Start API server in background
echo "🚀 Starting API server..."
./scripts/build-release.sh > /dev/null 2>&1 &
SERVER_PID=$!
sleep 3

# Test API endpoints
echo "✅ Testing API endpoints..."

# Health check
curl -s http://localhost:28126/health | grep -q '"status":"ok"'
echo "✅ Health endpoint OK"

# CPU endpoint
curl -s http://localhost:28126/api/cpu > /dev/null
echo "✅ CPU endpoint OK"

# Memory endpoint  
curl -s http://localhost:28126/api/mem > /dev/null
echo "✅ Memory endpoint OK"

# Stop server
kill $SERVER_PID 2>/dev/null || true

echo "🎉 API tests passed!"