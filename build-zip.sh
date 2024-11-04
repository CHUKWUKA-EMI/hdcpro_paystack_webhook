#!/bin/bash

# Set the output binary name
BINARY_NAME="bootstrap"

# Build the Go executable for Linux
echo "Building the Go executable for Linux..."
GOOS=linux go build -o $BINARY_NAME

# Zip the executable
echo "Zipping the executable..."
zip $BINARY_NAME.zip $BINARY_NAME

echo "Build and zip completed successfully!"