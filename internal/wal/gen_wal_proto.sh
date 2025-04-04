#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Define the directory where the proto file will be downloaded
PROTO_DIR=$(dirname "$0")
OUTPUT_DIR="$PROTO_DIR"

# Define the GitHub repository and file path
GITHUB_REPO="AshwinKul28/dicedb-protos"
GITHUB_BRANCH="patch-1"
PROTO_FILE_PATH="wal/wal.proto"
PROTO_FILE="$PROTO_DIR/wal.proto"

# Function to download the proto file from GitHub
download_proto_file() {
    echo "Downloading wal.proto from GitHub..."
    curl -s "https://raw.githubusercontent.com/$GITHUB_REPO/$GITHUB_BRANCH/$PROTO_FILE_PATH" -o "$PROTO_FILE"
    if [ $? -ne 0 ]; then
        echo "Failed to download wal.proto"
        exit 1
    fi
    echo "Successfully downloaded wal.proto to $PROTO_FILE"
}

# Function to install protoc if not present
install_protoc() {
    echo "protoc not found. Installing Protocol Buffers compiler..."
    if [[ "$(uname)" == "Darwin" ]]; then
        # MacOS installation
        brew install protobuf
    elif [[ "$(uname)" == "Linux" ]]; then
        # Linux installation
        sudo apt update && sudo apt install -y protobuf-compiler
    else
        echo "Unsupported OS. Please install 'protoc' manually."
        exit 1
    fi
}

# Function to install the Go plugin for protoc if not present
install_protoc_gen_go() {
    echo "protoc-gen-go not found. Installing Go plugin for protoc..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
    echo "Make sure $(go env GOPATH)/bin is in your PATH."
}

# Function to clean up downloaded files
cleanup() {
    echo "Cleaning up..."
    rm -f "$PROTO_FILE"
    echo "Cleanup complete."
}

# Set up trap to ensure cleanup happens even if script fails
trap cleanup EXIT

# Download the proto file
download_proto_file

# Check if protoc is installed, install if necessary
if ! command -v protoc &>/dev/null; then
    install_protoc
fi

# Check if the Go plugin for protoc is installed, install if necessary
if ! command -v protoc-gen-go &>/dev/null; then
    install_protoc_gen_go
fi

# Generate the wal.pb.go file
echo "Generating wal.pb.go from wal.proto..."
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative "$PROTO_FILE"

echo "Generation complete. File created at $OUTPUT_DIR/wal.pb.go"
