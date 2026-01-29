#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# GitHub repository
REPO="bimalpaudels/finks"
BINARY_NAME="finks"

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        *)          echo -e "${RED}Error: Unsupported operating system${NC}" >&2; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64)     arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)          echo -e "${RED}Error: Unsupported architecture${NC}" >&2; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
    local version=$(curl -s "${latest_url}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        echo -e "${RED}Error: Could not fetch latest version${NC}" >&2
        exit 1
    fi
    
    echo "$version"
}

# Download and install binary
install() {
    local platform=$(detect_platform)
    local version=$(get_latest_version)
    local download_url="https://github.com/${REPO}/releases/download/${version}/finks-${platform}"
    local temp_file=$(mktemp)
    
    echo -e "${GREEN}Installing Finks ${version} for ${platform}...${NC}"
    
    # Download binary
    echo "Downloading from ${download_url}..."
    if ! curl -fsSL -o "${temp_file}" "${download_url}"; then
        echo -e "${RED}Error: Failed to download binary${NC}" >&2
        rm -f "${temp_file}"
        exit 1
    fi
    
    # Make executable
    chmod +x "${temp_file}"
    
    # Execute the binary (which will run the installation wizard)
    echo -e "${GREEN}Running installation wizard...${NC}"
    "${temp_file}"
    
    # Cleanup
    rm -f "${temp_file}"
    
    echo -e "${GREEN}Installation complete!${NC}"
}

# Run installation
install
