#!/bin/bash
set -e

echo "Running templ generate..."
templ generate

echo "Building application..."
go build -o ./tmp/main ./cmd/obtura

echo "Build complete!"