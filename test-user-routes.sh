#!/bin/bash

# Test script for user management routes
# This helps verify that user management works in production

echo "🧪 Testing User Management Routes"
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/ > /dev/null; then
    echo "❌ Server not running on port 8080"
    echo "Please start the server first:"
    echo "  ./start-production.sh"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Test routes (these will redirect to login if not authenticated)
echo "📋 Testing routes..."

echo "  /users (user list):"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/users)
echo "    Status: $RESPONSE"

echo "  /users/new (new user form):"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/users/new)
echo "    Status: $RESPONSE"

echo "  /users/1/edit (edit user form):"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/users/1/edit)
echo "    Status: $RESPONSE"

echo ""
echo "📝 Note: Status 303 (redirect) is expected for unauthenticated requests"
echo "   This means the routes are working and redirecting to login"
echo ""
echo "🔐 To test with authentication:"
echo "   1. Log in via web browser at http://localhost:8080/login"
echo "   2. Navigate to http://localhost:8080/users"
echo "   3. Click 'Create New User' or 'Edit' buttons"
echo ""
echo "📊 Check logs for detailed information:"
echo "   tail -f logs/production.log"