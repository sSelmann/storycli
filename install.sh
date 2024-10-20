#!/bin/bash

set -e

# Determine the latest version or specify a version
VERSION="v0.1.0"

# Determine the OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Adjust ARCH format if necessary
if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
fi

# Define the download URL for storycli
DOWNLOAD_URL="https://github.com/sSelmann/storycli/releases/download/${VERSION}/storycli-${OS}-${ARCH}"

# Check if Go is installed and the version
REQUIRED_GO_VERSION="go1.22.4"
if command -v go >/dev/null 2>&1; then
    INSTALLED_GO_VERSION=$(go version | awk '{print $3}')
    if [ "$INSTALLED_GO_VERSION" != "$REQUIRED_GO_VERSION" ]; then
        echo "Go version $INSTALLED_GO_VERSION is currently installed. Required version is $REQUIRED_GO_VERSION."
        read -p "Do you want to remove the existing Go version and install $REQUIRED_GO_VERSION? [y/N]: " response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            echo "Updating Go to version $REQUIRED_GO_VERSION..."
            # Remove old Go version
            sudo rm -rf /usr/local/go
            # Download and install the required Go version
            wget "https://golang.org/dl/${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
            sudo tar -C /usr/local -xzf "${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
            rm "${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
        else
            echo "Go version was not updated. Exiting."
            exit 1
        fi
    else
        echo "Go is already at the required version."
    fi
else
    echo "Installing Go version $REQUIRED_GO_VERSION..."
    wget "https://golang.org/dl/${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
    sudo tar -C /usr/local -xzf "${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
    rm "${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
    # Add Go to PATH
    [ ! -f ~/.bash_profile ] && touch ~/.bash_profile
    echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bash_profile
    source ~/.bash_profile
fi

# Verify Go installation
if ! command -v go >/dev/null 2>&1; then
    echo "Go installation failed. Exiting."
    exit 1
fi

# Download the storycli binary
echo "Downloading storycli ${VERSION} for ${OS}/${ARCH}..."
curl -L $DOWNLOAD_URL -o storycli

# Make it executable
chmod +x storycli

# Move it to /usr/local/bin
echo "Installing storycli to /usr/local/bin..."
sudo mv storycli /usr/local/bin/

echo "storycli installed successfully!"
