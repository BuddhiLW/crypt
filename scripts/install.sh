#!/bin/bash

# Crypt - Installation Script
# This script installs the crypt CLI tool and its dependencies

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt-get &> /dev/null; then
            echo "ubuntu"
        elif command -v yum &> /dev/null; then
            echo "centos"
        elif command -v pacman &> /dev/null; then
            echo "arch"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Function to install dependencies
install_dependencies() {
    local os=$(detect_os)
    
    print_status "Detected OS: $os"
    print_status "Installing dependencies..."
    
    case $os in
        "ubuntu"|"debian")
            sudo apt-get update
            sudo apt-get install -y libjpeg-dev build-essential
            ;;
        "centos"|"rhel"|"fedora")
            sudo yum install -y libjpeg-devel gcc make
            ;;
        "arch")
            sudo pacman -S --noconfirm libjpeg-turbo base-devel
            ;;
        "macos")
            if command -v brew &> /dev/null; then
                brew install jpeg
            else
                print_error "Homebrew not found. Please install Homebrew first: https://brew.sh"
                exit 1
            fi
            ;;
        "windows")
            print_warning "Windows installation requires MSYS2 or WSL. Please install manually:"
            print_warning "  pacman -S mingw-w64-x86_64-libjpeg-turbo"
            ;;
        *)
            print_warning "Unknown OS. Please install libjpeg development library manually."
            ;;
    esac
    
    print_success "Dependencies installed successfully"
}

# Function to check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first: https://golang.org/dl/"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Found Go version: $go_version"
}

# Function to install crypt
install_crypt() {
    print_status "Installing crypt CLI tool..."
    
    # Check if we're in the project directory
    if [[ -f "go.mod" ]] && [[ -f "cmd/crypt/main.go" ]]; then
        print_status "Building from source..."
        go install ./cmd/crypt
    else
        print_status "Installing from GitHub..."
        go install github.com/BuddhiLW/crypt/cmd/crypt@latest
    fi
    
    print_success "crypt CLI tool installed successfully"
}

# Function to verify installation
verify_installation() {
    if command -v crypt &> /dev/null; then
        print_success "crypt CLI tool is available in PATH"
        crypt --help || crypt help
    else
        print_error "crypt CLI tool not found in PATH"
        print_warning "Make sure your Go bin directory is in your PATH:"
        print_warning "  export PATH=\$PATH:\$(go env GOPATH)/bin"
        exit 1
    fi
}

# Main installation function
main() {
    print_status "Starting crypt CLI tool installation..."
    
    check_go
    install_dependencies
    install_crypt
    verify_installation
    
    print_success "Installation completed successfully!"
    print_status "You can now use the 'crypt' command"
}

# Run main function
main "$@"
