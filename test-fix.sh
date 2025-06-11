#!/bin/bash

echo "🔧 TESTE DIE FIXES..."
echo "===================="

# Kill old server
pkill -f "./server" 2>/dev/null || true
sleep 1

# Start new server
echo "🚀 Starte reparierten Server..."
go build -o server cmd/server/main.go
./server -config=config.json &
SERVER_PID=$!
sleep 3

echo ""
echo "✅ Server läuft! Teste jetzt im Browser:"
echo "   http://localhost:9000/login"
echo ""
echo "🎯 Was du jetzt sehen solltest:"
echo "1. Logge dich ein"
echo "2. Gehe zu /users"
echo "3. Klicke 'Create New User' → Jetzt sollte das FORMULAR erscheinen!"
echo ""
echo "📊 DEBUG-Output läuft live..."
echo "Bei Problemen: Ctrl+C zum Beenden"
echo ""

# Wait for user to test
wait $SERVER_PID