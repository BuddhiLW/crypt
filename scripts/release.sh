#!/bin/bash

# Crypt - Release Script
# This script creates a new release with proper versioning

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

# Function to check if we're in a git repository
check_git() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
}

# Function to check if there are uncommitted changes
check_clean_working_dir() {
    if ! git diff-index --quiet HEAD --; then
        print_warning "You have uncommitted changes. Please commit or stash them first."
        git status --short
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Function to get current version
get_current_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "$latest_tag"
}

# Function to suggest next version
suggest_next_version() {
    local current_version=$(get_current_version)
    local major=$(echo "$current_version" | cut -d. -f1 | sed 's/v//')
    local minor=$(echo "$current_version" | cut -d. -f2)
    local patch=$(echo "$current_version" | cut -d. -f3)
    
    echo "v$major.$minor.$((patch + 1))"
}

# Function to create release
create_release() {
    local version=$1
    
    print_status "Creating release: $version"
    
    # Create and push tag
    git tag -a "$version" -m "Release $version"
    git push origin "$version"
    
    print_success "Tag $version created and pushed"
    
    # Build binaries
    print_status "Building binaries..."
    make build-all
    
    # Create release archives
    print_status "Creating release archives..."
    make release
    
    print_success "Release $version created successfully!"
    print_status "Binaries are available in the build/ directory"
    print_status "You can now create a GitHub release with the generated archives"
}

# Function to show help
show_help() {
    echo "Usage: $0 [OPTIONS] [VERSION]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -s, --suggest  Suggest next version number"
    echo "  -c, --current  Show current version"
    echo ""
    echo "Examples:"
    echo "  $0 v1.0.0      Create release v1.0.0"
    echo "  $0 --suggest   Suggest next version"
    echo "  $0 --current   Show current version"
}

# Main function
main() {
    # Parse command line arguments
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--suggest)
            print_status "Suggested next version: $(suggest_next_version)"
            exit 0
            ;;
        -c|--current)
            print_status "Current version: $(get_current_version)"
            exit 0
            ;;
        "")
            print_error "No version specified"
            echo ""
            show_help
            exit 1
            ;;
        v*)
            # Valid version format
            ;;
        *)
            print_error "Invalid version format. Use semantic versioning (e.g., v1.0.0)"
            exit 1
            ;;
    esac
    
    local version=$1
    
    # Pre-flight checks
    check_git
    check_clean_working_dir
    
    # Confirm release
    print_warning "About to create release: $version"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Release cancelled"
        exit 0
    fi
    
    # Create release
    create_release "$version"
}

# Run main function
main "$@"
