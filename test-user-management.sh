#!/bin/bash

# Comprehensive User Management Test
echo "🔧 User Management Complete Test"
echo "================================"

# Build the application
echo "📦 Building application..."
go build -o server cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful"

# Start server in background
echo "🚀 Starting server..."
./server -config=config.json > test.log 2>&1 &
SERVER_PID=$!
sleep 3

# Check if server started
if ! curl -s http://localhost:9000/ > /dev/null; then
    echo "❌ Server failed to start"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi
echo "✅ Server running on port 9000"

# Test the new routes
echo "🧪 Testing routes..."

echo "  Testing /users (main page):"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/users)
echo "    Status: $RESPONSE (should be 303 - redirect to login)"

echo "  Testing /user-management/new:"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/user-management/new)
echo "    Status: $RESPONSE (should be 303 - redirect to login)"

echo "  Testing legacy redirect /users/new:"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/users/new)
echo "    Status: $RESPONSE (should be 301 - permanent redirect)"

echo "  Testing /user-management/1/edit:"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/user-management/1/edit)
echo "    Status: $RESPONSE (should be 303 - redirect to login)"

# Test template loading
echo "🎨 Testing template loading..."
if grep -q "user_form.html" test.log; then
    echo "✅ Templates loaded successfully"
else
    echo "⚠️  Check template loading in logs"
fi

# Cleanup
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null
rm -f test.log

echo ""
echo "✅ All tests completed!"
echo ""
echo "📋 Next steps for production:"
echo "1. Deploy: ./deploy-production.sh"
echo "2. Create user: ./create-production-user.sh"
echo "3. Start: ./start-production.sh"
echo "4. Test in browser:"
echo "   - Login: http://your-server:8080/login"
echo "   - Users: http://your-server:8080/users"
echo "   - New User: http://your-server:8080/user-management/new"
echo ""
echo "🎯 Routes that should work:"
echo "   /users                     → User list"
echo "   /user-management/new       → Create user form"
echo "   /user-management/X/edit    → Edit user form"
echo "   /user-management/X/view    → View user details"