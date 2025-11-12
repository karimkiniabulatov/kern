#!/bin/bash
set -e
VERSION="1.2.0"
TAG="v$VERSION"

echo "🏷️ Creating GitHub release $TAG"
git tag -a "$TAG" -m "Release $TAG"
git push origin "$TAG"
echo "✅ Tag $TAG created"