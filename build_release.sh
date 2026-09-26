#!/bin/bash
set -e

echo "=== Building BitConfig Cross-Platform Release Binaries ==="

mkdir -p dist

# 1. Local / Current Host Binary
echo "-> Building local binary (./bitconfig)..."
go build -ldflags="-s -w" -o bitconfig main.go

# 2. macOS Apple Silicon (M1/M2/M3/M4)
echo "-> Building darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/bitconfig-darwin-arm64 main.go

# 3. macOS Intel
echo "-> Building darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/bitconfig-darwin-amd64 main.go

# 4. Linux x86_64
echo "-> Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/bitconfig-linux-amd64 main.go

# 5. Linux ARM64
echo "-> Building linux/arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/bitconfig-linux-arm64 main.go

# 6. Windows x86_64
echo "-> Building windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/bitconfig-windows-amd64.exe main.go

echo ""
echo "=== Release Build Complete! ==="
ls -lh dist/
