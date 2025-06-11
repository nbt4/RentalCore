#!/bin/bash

echo "🔍 LIVE DEBUG SESSION"
echo "===================="
echo ""
echo "Ich starte jetzt den Server mit Debug-Logging."
echo "Du kannst dann im Browser testen und wir sehen sofort was passiert!"
echo ""

# Check if there's already a server running
if pgrep -f "./server" > /dev/null; then
    echo "⚠️  Server läuft bereits. Stoppe ihn..."
    pkill -f "./server"
    sleep 2
fi

echo "🚀 Starte Debug-Server..."
echo "   URL: http://localhost:9000"
echo "   Login: http://localhost:9000/login"
echo ""

# Build fresh
go build -o server cmd/server/main.go

# Start server with debug output
echo "📋 DEBUG OUTPUT:"
echo "================"
./server -config=config.json