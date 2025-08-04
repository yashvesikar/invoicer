#!/bin/bash

# Invoice Manager Install Script
# This script checks prerequisites and builds the invoicer application

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script variables
MIN_GO_VERSION="1.18"
BINARY_NAME="invoicer"
INSTALL_DIR="/usr/local/bin"
LOCAL_INSTALL=true

# Print colored output
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

print_info() {
    echo -e "${BLUE}→${NC} $1"
}

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Darwin*)    OS="macos";;
        Linux*)     OS="linux";;
        MINGW*|MSYS*|CYGWIN*) OS="windows";;
        *)          OS="unknown";;
    esac
    echo "$OS"
}

# Compare version numbers
version_ge() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" = "$2"
}

# Check if Go is installed and meets minimum version
check_go() {
    print_info "Checking Go installation..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        echo "Please install Go $MIN_GO_VERSION or higher from https://golang.org/dl/"
        return 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if version_ge "$GO_VERSION" "$MIN_GO_VERSION"; then
        print_success "Go $GO_VERSION found"
        return 0
    else
        print_error "Go version $GO_VERSION is below minimum required version $MIN_GO_VERSION"
        echo "Please upgrade Go from https://golang.org/dl/"
        return 1
    fi
}

# Check if pdflatex is installed
check_pdflatex() {
    print_info "Checking pdflatex installation..."
    
    if command -v pdflatex &> /dev/null; then
        print_success "pdflatex found"
        return 0
    else
        print_warning "pdflatex not found (optional, needed for PDF export)"
        
        OS=$(detect_os)
        case "$OS" in
            macos)
                echo "To install on macOS:"
                echo "  brew install --cask basictex"
                echo "  or"
                echo "  brew install --cask mactex"
                ;;
            linux)
                if command -v apt-get &> /dev/null; then
                    echo "To install on Debian/Ubuntu:"
                    echo "  sudo apt-get install texlive-latex-base"
                elif command -v yum &> /dev/null; then
                    echo "To install on Fedora/RHEL:"
                    echo "  sudo yum install texlive-latex"
                elif command -v pacman &> /dev/null; then
                    echo "To install on Arch:"
                    echo "  sudo pacman -S texlive-core"
                else
                    echo "Please install TeX Live from your distribution's package manager"
                fi
                ;;
            windows)
                echo "To install on Windows:"
                echo "  Download and install MiKTeX from https://miktex.org/download"
                echo "  or"
                echo "  Download and install TeX Live from https://www.tug.org/texlive/"
                ;;
        esac
        
        read -p "Continue without pdflatex? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            return 1
        fi
    fi
}

# Build the application
build_app() {
    print_info "Building $BINARY_NAME..."
    
    if go build -o "$BINARY_NAME"; then
        print_success "Build successful"
        return 0
    else
        print_error "Build failed"
        return 1
    fi
}

# Create required directories
create_directories() {
    print_info "Creating required directories..."
    
    for dir in data exports; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            print_success "Created $dir/"
        else
            print_info "$dir/ already exists"
        fi
    done
}

# Install binary to system path
install_binary() {
    if [ "$LOCAL_INSTALL" = true ]; then
        print_info "Binary created in current directory: ./$BINARY_NAME"
        return 0
    fi
    
    print_info "Installing $BINARY_NAME to $INSTALL_DIR..."
    
    if [ ! -w "$INSTALL_DIR" ]; then
        print_warning "Need sudo permission to install to $INSTALL_DIR"
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
    else
        cp "$BINARY_NAME" "$INSTALL_DIR/"
    fi
    
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_success "Installed to $INSTALL_DIR/$BINARY_NAME"
        return 0
    else
        print_error "Installation failed"
        return 1
    fi
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    if [ "$LOCAL_INSTALL" = true ]; then
        if [ -x "./$BINARY_NAME" ]; then
            print_success "Installation verified"
            return 0
        fi
    else
        if command -v "$BINARY_NAME" &> /dev/null; then
            print_success "Installation verified"
            return 0
        fi
    fi
    
    print_error "Installation verification failed"
    return 1
}

# Show usage
usage() {
    echo "Invoice Manager Install Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -g, --global    Install binary to system path ($INSTALL_DIR)"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "By default, the binary is built in the current directory."
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -g|--global)
            LOCAL_INSTALL=false
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main installation process
main() {
    echo "========================================"
    echo "  Invoice Manager Installation Script   "
    echo "========================================"
    echo ""
    
    # Check prerequisites
    if ! check_go; then
        exit 1
    fi
    
    check_pdflatex
    
    # Build application
    if ! build_app; then
        exit 1
    fi
    
    # Create directories
    create_directories
    
    # Install binary
    if ! install_binary; then
        exit 1
    fi
    
    # Verify installation
    if ! verify_installation; then
        exit 1
    fi
    
    echo ""
    print_success "Installation complete!"
    echo ""
    echo "Next steps:"
    if [ "$LOCAL_INSTALL" = true ]; then
        echo "  1. Run the application: ./$BINARY_NAME"
    else
        echo "  1. Run the application: $BINARY_NAME"
    fi
    echo "  2. Create your first client"
    echo "  3. Create your first invoice"
    echo ""
    echo "For more information, see README.md"
}

# Run main installation
main