#!/bin/bash

# Define variables
REPO="mreider/quacker"
BINARY_NAME="quacker"
INSTALL_DIR="/usr/local/bin"

# Ensure the script is being run with curl piping
if [ -z "$REPO" ]; then
    echo "Error: REPO variable is not set. Make sure you are piping the script correctly."
    exit 1
fi

# Get the latest release URL
LATEST_RELEASE_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep "browser_download_url" | grep "linux-amd64" | cut -d '"' -f 4)

# Check if the URL was retrieved
if [ -z "$LATEST_RELEASE_URL" ]; then
    echo "Failed to retrieve the latest release. Please check the repository name or your internet connection."
    exit 1
fi

# Download the latest release
TMP_FILE="/tmp/$BINARY_NAME"
echo "Downloading the latest release from $LATEST_RELEASE_URL..."
curl -L -o "$TMP_FILE" "$LATEST_RELEASE_URL"

# Make the binary executable
echo "Making the binary executable..."
chmod +x "$TMP_FILE"

# Move the binary to the install directory
echo "Moving the binary to $INSTALL_DIR with sudo permissions..."
sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"

# Verify installation
if [ -x "$(command -v $BINARY_NAME)" ]; then
    echo "$BINARY_NAME successfully installed to $INSTALL_DIR."
else
    echo "Failed to install $BINARY_NAME. Please check for errors."
    exit 1
fi

# Print the installed version
echo "Installed $BINARY_NAME version:"
$BINARY_NAME --version