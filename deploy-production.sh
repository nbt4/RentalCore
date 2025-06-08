#!/bin/bash

# JobScanner Pro - Production Deployment Script

set -e

echo "🚀 Deploying JobScanner Pro to Production..."

# Build the application for production
echo "📦 Building application..."
go build -o server ./cmd/server

# Create logs directory
mkdir -p logs

# Install systemd service (requires root)
if [ "$EUID" -eq 0 ]; then
    echo "🔧 Installing systemd service..."
    cp jobscanner.service /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable jobscanner
    echo "✅ Systemd service installed and enabled"
    
    echo "📝 To start the service, run:"
    echo "   sudo systemctl start jobscanner"
    echo "   sudo systemctl status jobscanner"
else
    echo "⚠️  Run as root to install systemd service:"
    echo "   sudo ./deploy-production.sh"
fi

echo ""
echo "🔑 Make sure to set these environment variables:"
echo "   export DB_HOST=your_database_host"
echo "   export DB_NAME=TS-Lager"  
echo "   export DB_USER=your_database_user"
echo "   export DB_PASSWORD=your_database_password"
echo ""
echo "🚀 To start manually:"
echo "   ./start-production.sh"
echo ""
echo "✅ Deployment complete!"