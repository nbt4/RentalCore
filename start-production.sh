#!/bin/bash

# JobScanner Pro - Production Startup Script
# Uses config.production.direct.json for configuration

set -e

echo "Starting JobScanner Pro in Production Mode..."

# Set Go to release mode for better performance
export GIN_MODE=release

# Create logs directory if it doesn't exist
mkdir -p logs

# Check if production config file exists
if [ ! -f "config.production.direct.json" ]; then
    echo "Error: config.production.direct.json not found"
    echo "Please ensure the production config file exists in the current directory"
    exit 1
fi

echo "Using configuration file: config.production.direct.json"
echo "Server will start on port 8080"
echo "Logs will be written to: logs/production.log"

# Start the application with production config
exec ./server -config=config.production.direct.json >> logs/production.log 2>&1