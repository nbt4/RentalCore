#!/bin/bash

echo "🔍 Debug: User Management Flow Test"
echo "=================================="

# Start server
echo "🚀 Starting server..."
go build -o server cmd/server/main.go
./server -config=config.json > debug.log 2>&1 &
SERVER_PID=$!
sleep 3

echo "📋 Testing all user management URLs:"
echo ""

echo "1. /users (user list page):"
curl -s -i http://localhost:9000/users | head -n 3

echo ""
echo "2. /users/new (should redirect to /user-management/new):"
curl -s -i http://localhost:9000/users/new | head -n 3

echo ""
echo "3. /user-management/new (direct form access):"
curl -s -i http://localhost:9000/user-management/new | head -n 3

echo ""
echo "📄 Checking if templates contain correct URLs..."
echo "Looking for URLs in users_list.html:"
grep -n "href.*user" web/templates/users_list.html

echo ""
echo "📊 Server routes registered:"
grep -A 20 "GIN-debug.*GET.*users" debug.log 2>/dev/null | head -10

# Cleanup
kill $SERVER_PID 2>/dev/null
rm -f debug.log

echo ""
echo "🎯 SUMMARY:"
echo "If you see 'href=\"/user-management/new\"' above, templates are correct."
echo "If you see 'Location: /user-management/new' above, redirects work."
echo "Both should result in the user form, not the user list!"