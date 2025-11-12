#!/bin/bash
set -e

echo "🧪 Running kern tests..."

# Test basic compilation
echo "✅ Testing compilation..."
go build -o /tmp/kern-test ./cmd/kern

# Test help command
echo "✅ Testing help command..."
/tmp/kern-test --help > /dev/null

# Test version command
echo "✅ Testing version command..."
/tmp/kern-test --version > /dev/null

# Test service status
echo "✅ Testing service commands..."
/tmp/kern-test --service-status > /dev/null || true

# Test language commands
echo "✅ Testing language commands..."
/tmp/kern-test --list-languages > /dev/null

# Cleanup
rm -f /tmp/kern-test

echo "🎉 All tests passed!"