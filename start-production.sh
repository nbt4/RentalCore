#!/bin/bash

# JobScanner Pro - Production Startup Script
# Make sure to set these environment variables before running

set -e

echo "Starting JobScanner Pro in Production Mode..."

# Check required environment variables
required_vars=("DB_HOST" "DB_NAME" "DB_USER" "DB_PASSWORD")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: Environment variable $var is not set"
        echo "Please set all required variables:"
        echo "  DB_HOST=your_database_host"
        echo "  DB_NAME=your_database_name" 
        echo "  DB_USER=your_database_user"
        echo "  DB_PASSWORD=your_database_password"
        exit 1
    fi
done

# Set Go to release mode for better performance
export GIN_MODE=release

# Create logs directory if it doesn't exist
mkdir -p logs

# Replace environment variables in config file
envsubst < config.production.json > config.runtime.json

echo "Configuration file generated with environment variables"
echo "Database Host: $DB_HOST"
echo "Database Name: $DB_NAME" 
echo "Database User: $DB_USER"
echo "Server will start on port 8080"

# Start the application with production config
exec ./server -config=config.runtime.json >> logs/production.log 2>&1