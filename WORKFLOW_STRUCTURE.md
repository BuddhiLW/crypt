# GitHub Actions Workflow Structure

## Overview

The project uses two GitHub Actions workflows to handle CI/CD and releases:

1. **CI Workflow** (`.github/workflows/ci.yml`)
2. **GoReleaser Workflow** (`.github/workflows/goreleaser.yml`)

## Workflow Details

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `master` branches
- Pull requests to `main` or `master` branches

**Jobs:**

#### Test Job
- **Purpose**: Run tests across multiple platforms and Go versions
- **Matrix**: 
  - Go versions: `1.23.5`, `1.22`
  - Platforms: `ubuntu-latest`, `macos-latest`, `windows-latest`
- **Actions**:
  - Install platform-specific dependencies (libjpeg-dev)
  - Run tests with `go test -v ./...`
  - Run `go vet ./...`
  - Check code formatting with `gofmt`

#### Build Job
- **Purpose**: Build binaries for all platforms
- **Platform**: `ubuntu-latest`
- **Actions**:
  - Install dependencies (libjpeg-dev, build-essential)
  - Build for current platform with `make build`
  - Build for all platforms with `make build-all`
  - Upload build artifacts

#### Lint Job
- **Purpose**: Code linting and style checking
- **Platform**: `ubuntu-latest`
- **Actions**:
  - Install dependencies
  - Install golangci-lint
  - Run linting with `golangci-lint run`

#### Security Job
- **Purpose**: Security vulnerability scanning
- **Platform**: `ubuntu-latest`
- **Actions**:
  - Install dependencies
  - Run security check with `govulncheck ./...`

### 2. GoReleaser Workflow (`.github/workflows/goreleaser.yml`)

**Triggers:**
- Push tags matching `v*` (e.g., `v1.0.0`)

**Jobs:**

#### GoReleaser Job
- **Purpose**: Automated release creation
- **Platform**: `ubuntu-latest`
- **Permissions**: 
  - `contents: write` (for creating releases)
  - `packages: write` (for publishing packages)
- **Actions**:
  - Install dependencies (libjpeg-dev, build-essential)
  - Run GoReleaser to:
    - Build binaries for supported platforms
    - Create GitHub releases
    - Generate checksums
    - Update changelog

## Workflow Separation

### Why Two Workflows?

1. **CI Workflow**: Handles development workflow
   - Runs on every push and PR
   - Ensures code quality
   - Tests across multiple platforms
   - Builds for development purposes

2. **GoReleaser Workflow**: Handles releases
   - Runs only on tag pushes
   - Creates official releases
   - Distributes binaries
   - Updates documentation

### Benefits of This Structure

1. **Separation of Concerns**: Development vs. Release
2. **Performance**: CI runs frequently, releases run rarely
3. **Security**: Different permission levels for different purposes
4. **Reliability**: Each workflow has a specific, focused purpose

## CGO Dependencies

Both workflows handle CGO dependencies properly:

### Linux (Ubuntu)
```bash
sudo apt-get update
sudo apt-get install -y libjpeg-dev build-essential
```

### macOS
```bash
brew install jpeg
```

### Windows
- Dependencies handled by the build environment
- May require manual setup in some cases

## Release Process

### Manual Release
```bash
# Create and push a tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Using Release Script
```bash
# Use the provided script
./scripts/release.sh v1.0.0
```

### What Happens Automatically

1. **GoReleaser Workflow** triggers on tag push
2. **Dependencies** are installed
3. **Binaries** are built for supported platforms
4. **GitHub Release** is created with:
   - Release notes from changelog
   - Binary downloads
   - Checksums for verification
5. **Artifacts** are uploaded to GitHub releases

## Supported Platforms

### GoReleaser Builds
- **Linux**: amd64, arm64
- **Windows**: amd64, arm64 (if CGO works)
- **macOS**: amd64, arm64 (if CGO works)

### CI Builds
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64
- **Windows**: amd64, arm64

## Monitoring

### CI Workflow
- **Success**: All tests pass, builds succeed
- **Failure**: Code quality issues, build failures
- **Duration**: Typically 5-10 minutes

### GoReleaser Workflow
- **Success**: Release created successfully
- **Failure**: Build issues, permission problems
- **Duration**: Typically 3-5 minutes

## Troubleshooting

### Common Issues

1. **CGO Build Failures**
   - Ensure dependencies are installed
   - Check platform compatibility

2. **Permission Errors**
   - Verify GitHub token permissions
   - Check workflow permissions

3. **Cross-Platform Build Failures**
   - CGO limitations on some platforms
   - Use native builds when possible

### Debugging

1. **Check Workflow Logs**: GitHub Actions provides detailed logs
2. **Test Locally**: Use `make build` to test locally
3. **Verify Dependencies**: Ensure libjpeg is available

## Future Improvements

1. **Better Cross-Platform Support**: Improve CGO handling
2. **Docker Integration**: Use containers for consistent builds
3. **Package Distribution**: Add support for package managers
4. **Performance Optimization**: Reduce build times

## Conclusion

This workflow structure provides:
- **Reliable CI/CD** for development
- **Automated releases** for distribution
- **Proper CGO handling** across platforms
- **Clear separation** of concerns

The workflows work together to ensure code quality and reliable releases while handling the complexities of CGO dependencies.
