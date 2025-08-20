# Crypt - Installation Guide

This guide explains how to install and distribute the Crypt CLI tool for all major platforms.

## Overview

Crypt is a CLI tool for steganography that uses CGO with libjpeg for JPEG processing. This means it requires native compilation and cannot be distributed as a statically linked binary.

## Installation Methods

### 1. Using `go install` (Recommended for Users)

```bash
# Install the latest version
go install github.com/BuddhiLW/crypt/cmd/crypt@latest

# Install a specific version
go install github.com/BuddhiLW/crypt/cmd/crypt@v1.0.0
```

**Prerequisites:**
- Go 1.23.5 or later
- libjpeg development library

### 2. Using the Installation Script

```bash
# Clone the repository
git clone https://github.com/BuddhiLW/crypt.git
cd crypt

# Run the installation script
./scripts/install.sh
```

The script will:
- Detect your operating system
- Install required dependencies
- Build and install the crypt CLI tool
- Verify the installation

### 3. Manual Installation

#### Prerequisites by Platform

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install libjpeg-dev build-essential
```

**CentOS/RHEL/Fedora:**
```bash
sudo yum install libjpeg-devel gcc make
```

**Arch Linux:**
```bash
sudo pacman -S libjpeg-turbo base-devel
```

**macOS:**
```bash
brew install jpeg
```

**Windows (MSYS2):**
```bash
pacman -S mingw-w64-x86_64-libjpeg-turbo
```

#### Build and Install

```bash
# Clone the repository
git clone https://github.com/BuddhiLW/crypt.git
cd crypt

# Build for current platform
go build -o crypt ./cmd/crypt

# Or install locally
go install ./cmd/crypt
```

## Cross-Platform Distribution

### Challenges

Due to CGO dependencies, cross-compilation is challenging:

1. **CGO Requirement**: The project uses C code for JPEG processing
2. **Library Dependencies**: Requires libjpeg on target systems
3. **Platform-Specific Binaries**: Each platform needs its own build

### Solutions

#### 1. Native Builds (Recommended)

Build on each target platform:

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o crypt-linux-amd64 ./cmd/crypt

# Linux (arm64)
GOOS=linux GOARCH=arm64 go build -o crypt-linux-arm64 ./cmd/crypt

# macOS (amd64)
GOOS=darwin GOARCH=amd64 go build -o crypt-darwin-amd64 ./cmd/crypt

# macOS (arm64)
GOOS=darwin GOARCH=arm64 go build -o crypt-darwin-arm64 ./cmd/crypt

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o crypt-windows-amd64.exe ./cmd/crypt

# Windows (arm64)
GOOS=windows GOARCH=arm64 go build -o crypt-windows-arm64.exe ./cmd/crypt
```

#### 2. Using Makefile

```bash
# Build for all platforms (may fail due to CGO)
make build-all

# Build for specific platforms
make build-linux
make build-darwin
make build-windows
```

#### 3. Using Docker

Create a Dockerfile for each target platform:

```dockerfile
# Example for Linux
FROM golang:1.23.5-alpine AS builder
RUN apk add --no-cache libjpeg-dev
WORKDIR /app
COPY . .
RUN go build -o crypt ./cmd/crypt

FROM alpine:latest
RUN apk add --no-cache libjpeg
COPY --from=builder /app/crypt /usr/local/bin/
ENTRYPOINT ["crypt"]
```

#### 4. Using GitHub Actions

The project includes GitHub Actions workflows that:
- Build for multiple platforms
- Create release archives
- Handle CGO dependencies

## Distribution Strategies

### 1. Source Distribution

Distribute the source code and let users build locally:

```bash
# Users can install directly
go install github.com/BuddhiLW/crypt/cmd/crypt@latest
```

### 2. Platform-Specific Binaries

Provide pre-built binaries for each platform:

- `crypt-linux-amd64`
- `crypt-linux-arm64`
- `crypt-darwin-amd64`
- `crypt-darwin-arm64`
- `crypt-windows-amd64.exe`
- `crypt-windows-arm64.exe`

### 3. Package Managers

#### Homebrew (macOS)
```bash
# Create a homebrew formula
brew tap BuddhiLW/crypt
brew install crypt
```

#### Snap (Linux)
```bash
# Create a snap package
snapcraft
```

#### Docker
```bash
# Pull and run
docker pull ghcr.io/buddhilw/crypt:latest
docker run --rm ghcr.io/buddhilw/crypt:latest help
```

## Release Process

### 1. Using the Release Script

```bash
# Create a new release
./scripts/release.sh v1.0.0

# Suggest next version
./scripts/release.sh --suggest

# Show current version
./scripts/release.sh --current
```

### 2. Manual Release Process

```bash
# 1. Update version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 2. Build binaries
make build-all

# 3. Create archives
make release

# 4. Upload to GitHub releases
```

### 3. Automated Releases with GoReleaser

The project includes a `.goreleaser.yml` configuration for automated releases:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Create a release
goreleaser release --snapshot --rm-dist

# Create a real release
goreleaser release
```

## Testing Installation

After installation, verify it works:

```bash
# Check if crypt is available
which crypt

# Test the help command
crypt help

# Test basic functionality
crypt encrypt help
crypt decrypt help
```

## Troubleshooting

### Common Issues

1. **"undefined: ExtractQRCodeFromJPEG"**
   - Ensure CGO is enabled: `CGO_ENABLED=1`
   - Install libjpeg development library

2. **"cannot find -ljpeg"**
   - Install libjpeg development library for your platform

3. **Cross-compilation fails**
   - Use native builds on each platform
   - Or use Docker for cross-platform builds

4. **Permission denied**
   - Ensure the binary is executable: `chmod +x crypt`
   - Check PATH includes Go bin directory

### Getting Help

- Check the [README.md](README.md) for basic usage
- Open an issue on GitHub for bugs
- Check the [Makefile](Makefile) for build options

## Best Practices

1. **Always test on target platforms** before releasing
2. **Use semantic versioning** for releases
3. **Provide clear installation instructions** for each platform
4. **Include dependency information** in documentation
5. **Use CI/CD** for automated testing and building
6. **Provide multiple installation methods** for user convenience

## Conclusion

While CGO makes distribution more complex, it provides the necessary performance and functionality for JPEG processing. The recommended approach is to:

1. Use `go install` for development and testing
2. Build native binaries for each platform for distribution
3. Use Docker for containerized deployments
4. Leverage GitHub Actions for automated releases

This ensures users get the best experience while maintaining the tool's functionality.
