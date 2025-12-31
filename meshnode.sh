#!/bin/bash

# Check if port argument is provided
if [ -z "$1" ]
  then
    echo "❌ Error: No port provided."
    echo "Usage: ./meshnode.sh <port_number>"
    exit 1
fi

PORT=$1

echo "🌐 Initializing Mesh Network..."
echo "📡 Listening on Port: $PORT"

# Running all files in the current directory
go run main.go \
       publish.go \
       runtime_detect.go \
       dockerfileTemplate.go \
       buildImage.go \
       capabilities.go \
       resource_protocol.go \
       service_identity.go \
       node.go \
       gateway.go \
       scheduler.go \
       -port "$PORT"
