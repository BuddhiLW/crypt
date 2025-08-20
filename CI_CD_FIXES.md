# CI/CD Fixes for CGO Dependencies

## Problem

The GitHub Actions CI/CD was failing with the error:
```
fatal error: jpeglib.h: No such file or directory
```

This occurred because the project uses CGO (C Go) with libjpeg for JPEG processing, but the GitHub Actions runners didn't have the required development libraries installed.

## Solution

### 1. Updated GitHub Actions Workflows

Added dependency installation steps to all workflows:

#### CI Workflow (`.github/workflows/ci.yml`)
```yaml
- name: Install dependencies
  run: |
    sudo apt-get update
    sudo apt-get install -y libjpeg-dev build-essential
```

This step was added to:
- **Test job**: For running tests across multiple platforms
- **Build job**: For building binaries
- **Lint job**: For code linting
- **Security job**: For security checks

#### GoReleaser Workflow (`.github/workflows/goreleaser.yml`)
```yaml
- name: Install dependencies
  run: |
    sudo apt-get update
    sudo apt-get install -y libjpeg-dev build-essential
```

### 2. Platform-Specific Dependency Installation

For the test matrix that runs on multiple platforms:

```yaml
- name: Install dependencies
  run: |
    if [[ "$RUNNER_OS" == "Linux" ]]; then
      sudo apt-get update
      sudo apt-get install -y libjpeg-dev build-essential
    elif [[ "$RUNNER_OS" == "macOS" ]]; then
      brew install jpeg
    elif [[ "$RUNNER_OS" == "Windows" ]]; then
      echo "Windows dependencies should be handled by the build environment"
    fi
```

### 3. Updated GoReleaser Configuration

Modified `.goreleaser.yml` to:
- Enable CGO: `CGO_ENABLED=1`
- Focus on Linux builds (where CGO works reliably)
- Disable problematic cross-platform builds

## CGO Limitations and Workarounds

### Why CGO Makes Distribution Complex

1. **Cross-compilation issues**: CGO requires target-specific C libraries
2. **Platform dependencies**: Each platform needs its own libjpeg installation
3. **Static linking limitations**: Cannot create truly portable binaries

### Recommended Distribution Strategy

#### 1. Primary: `go install` (Recommended)
```bash
go install github.com/BuddhiLW/crypt/cmd/crypt@latest
```
- Users build locally with their platform's dependencies
- Most reliable method
- Works on all platforms with proper dependencies

#### 2. Secondary: Native Builds
Build on each target platform:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o crypt-linux-amd64 ./cmd/crypt

# macOS
GOOS=darwin GOARCH=amd64 go build -o crypt-darwin-amd64 ./cmd/crypt

# Windows
GOOS=windows GOARCH=amd64 go build -o crypt-windows-amd64.exe ./cmd/crypt
```

#### 3. Container: Docker
```dockerfile
FROM golang:1.23.5-alpine AS builder
RUN apk add --no-cache libjpeg-dev
WORKDIR /app
COPY . .
RUN go build -o crypt ./cmd/crypt
```

## Testing the Fixes

### Local Testing
```bash
# Test build
make build

# Test cross-platform builds (may fail due to CGO)
make build-all

# Test installation script
./scripts/install.sh
```

### CI/CD Testing
The workflows now:
1. Install libjpeg-dev and build-essential
2. Build the project successfully
3. Run tests and linting
4. Create release artifacts

## Future Improvements

### 1. Alternative Approaches
Consider these alternatives to reduce CGO dependency:

1. **Pure Go JPEG library**: Use a Go-native JPEG library
2. **Conditional compilation**: Use CGO only when needed
3. **Plugin architecture**: Separate CGO code into plugins

### 2. Better Cross-Platform Support
- Use Docker for consistent builds across platforms
- Set up native build environments for each platform
- Use GitHub Actions matrix builds with platform-specific runners

### 3. Package Distribution
- Create platform-specific packages (deb, rpm, brew)
- Use snap/flatpak for Linux distribution
- Provide Docker images for containerized usage

## Monitoring

After these fixes, monitor:
- CI/CD pipeline success rates
- Build times (may increase due to dependency installation)
- User installation success rates
- Platform-specific issues

## Conclusion

The CI/CD fixes ensure that:
1. **Builds succeed** on GitHub Actions
2. **Tests run properly** across platforms
3. **Releases are created** successfully
4. **Users can install** via `go install`

While CGO makes distribution more complex, these fixes provide a reliable foundation for development and release automation.
