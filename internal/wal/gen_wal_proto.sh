#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Define the directory where the proto file resides
PROTO_DIR=$(dirname "$0")

# Define the proto file and the output file
PROTO_FILE="$PROTO_DIR/wal.proto"
OUTPUT_DIR="."

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
