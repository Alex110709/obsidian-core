#!/bin/bash

# Obsidian Core - Multi-architecture Docker Build & Push Script
# Builds for AMD64 and ARM64 platforms

set -e  # Exit on error

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo "❌ Error: Docker is not running. Please start Docker Desktop."
  exit 1
fi

# Default variables
IMAGE_NAME="obsidian-node"
DOCKER_USER="yuchanshin"
VERSION="${1:-v1.2.0}"  # Default to v1.2.0 if not provided

FULL_IMAGE_NAME="$DOCKER_USER/$IMAGE_NAME"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🐳 Obsidian Docker Multi-Arch Build"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Image:    $FULL_IMAGE_NAME"
echo "  Version:  $VERSION"
echo "  Tags:     $VERSION, latest"
echo "  Platforms: linux/amd64, linux/arm64"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if buildx is available
if ! docker buildx version > /dev/null 2>&1; then
  echo "❌ Error: docker buildx is not available."
  echo "Please enable it in Docker Desktop settings."
  exit 1
fi

# Create and use a new builder instance (if not exists)
BUILDER_NAME="obsidian-builder"
if ! docker buildx inspect $BUILDER_NAME > /dev/null 2>&1; then
  echo "📦 Creating new buildx builder: $BUILDER_NAME..."
  docker buildx create --name $BUILDER_NAME --use
else
  echo "📦 Using existing builder: $BUILDER_NAME"
  docker buildx use $BUILDER_NAME
fi

# Bootstrap the builder
echo "🔧 Bootstrapping builder..."
docker buildx inspect --bootstrap

# Build and push multi-architecture image
echo ""
echo "🏗️  Building multi-architecture image..."
echo "   This may take several minutes..."
echo ""

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag $FULL_IMAGE_NAME:$VERSION \
  --tag $FULL_IMAGE_NAME:latest \
  --push \
  .

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Success! Images pushed to Docker Hub:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  $FULL_IMAGE_NAME:$VERSION"
echo "  $FULL_IMAGE_NAME:latest"
echo ""
echo "📋 Supported architectures:"
echo "  - linux/amd64 (x86_64)"
echo "  - linux/arm64 (ARM 64-bit)"
echo ""
echo "🚀 Pull and run:"
echo "  docker pull $FULL_IMAGE_NAME:$VERSION"
echo "  docker compose up -d"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
