#!/bin/bash

# Harvest CLI Utility Installation Script
# This script installs the Harvest CLI utility on macOS or Windows (via Git Bash)

set -e  # Exit on error

# Print with colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS
OS="$(uname -s)"
echo -e "${BLUE}Detected operating system: ${OS}${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in your PATH.${NC}"
    echo -e "Please install Go from https://golang.org/dl/ and try again."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

if [[ "$GO_MAJOR" -lt 1 || ("$GO_MAJOR" -eq 1 && "$GO_MINOR" -lt 18) ]]; then
    echo -e "${YELLOW}Warning: Go version $GO_VERSION detected.${NC}"
    echo -e "${YELLOW}This project recommends Go 1.18 or higher.${NC}"
    read -p "Continue anyway? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}Go version $GO_VERSION detected. ✓${NC}"
fi

# Build the project
echo -e "${BLUE}Building Harvest CLI utility...${NC}"
go build -o h

# Determine installation path
INSTALL_PATH=""
case "$OS" in
    Darwin)  # macOS
        INSTALL_PATH="/usr/local/bin/h"
        ;;
    MINGW*|MSYS*|CYGWIN*)  # Windows
        INSTALL_PATH="$HOME/bin/h"
        mkdir -p "$HOME/bin" 2>/dev/null || true
        ;;
    Linux)  # Linux
        INSTALL_PATH="/usr/local/bin/h"
        ;;
    *)
        echo -e "${YELLOW}Unknown operating system. Will install to ./h only.${NC}"
        echo -e "${YELLOW}You may need to manually move the binary to your PATH.${NC}"
        exit 0
        ;;
esac

# Copy the binary to the installation path
if [[ "$OS" == "Darwin" || "$OS" == "Linux" ]]; then
    echo -e "${BLUE}Installing to $INSTALL_PATH...${NC}"
    if [[ -w "$(dirname "$INSTALL_PATH")" ]]; then
        cp h "$INSTALL_PATH"
    else
        echo -e "${YELLOW}Requesting administrator privileges to install to $INSTALL_PATH${NC}"
        sudo cp h "$INSTALL_PATH"
    fi
else  # Windows
    echo -e "${BLUE}Installing to $INSTALL_PATH...${NC}"
    cp h "$INSTALL_PATH"
    
    # Check if the directory is in PATH
    if [[ ":$PATH:" != *":$HOME/bin:"* ]]; then
        echo -e "${YELLOW}Adding $HOME/bin to your PATH in .bashrc${NC}"
        echo 'export PATH="$HOME/bin:$PATH"' >> "$HOME/.bashrc"
        echo -e "${YELLOW}Please restart your terminal or run 'source ~/.bashrc' to update your PATH.${NC}"
    fi
fi

# Check if config.json exists
if [[ ! -f "config.json" ]]; then
    echo -e "${BLUE}Creating sample config.json...${NC}"
    cat > config.json << EOF
{
  "projects": [
    {
      "id": 123,
      "name": "Project A",
      "tasks": [
        {
          "id": 456,
          "name": "Software Development"
        },
        {
          "id": 678,
          "name": "Non-Billable"
        }
      ]
    }
  ],
  "default_project": "Project A",
  "default_task": "Software Development",
  "harvest_api": {
    "account_id": "YOUR_HARVEST_ACCOUNT_ID",
    "token": "YOUR_HARVEST_API_TOKEN"
  }
}
EOF
    echo -e "${YELLOW}A sample config.json has been created.${NC}"
    echo -e "${YELLOW}Please edit it with your Harvest API credentials before using the tool.${NC}"
fi

echo -e "${GREEN}Installation complete!${NC}"
echo -e "${GREEN}You can now use the Harvest CLI utility by running 'h' in your terminal.${NC}"
echo -e "${BLUE}For help, run 'h --help'${NC}"

# Provide instructions for getting Harvest API credentials
echo -e "\n${YELLOW}Don't forget to update your config.json with your Harvest API credentials:${NC}"
echo -e "1. Log in to your Harvest account"
echo -e "2. Go to Settings > Developer"
echo -e "3. Create a new personal access token"
echo -e "4. Note your Account ID and Token"
echo -e "5. Update the config.json file with these values"

# Provide instructions for first use
echo -e "\n${BLUE}Quick start:${NC}"
echo -e "- List today's entries: ${GREEN}h list${NC}"
echo -e "- Create a new entry: ${GREEN}h create${NC}"
echo -e "- Create with defaults: ${GREEN}h create -D${NC}"
echo -e "- Update an entry: ${GREEN}h update${NC}"
echo -e "- Delete entries: ${GREEN}h delete${NC}" 