#!/bin/bash

# Comprehensive test script to verify no blank pages exist
# This script tests critical routes that were previously problematic

echo "🔍 Testing Application Routes for Blank Pages"
echo "============================================="

# Test configuration
BASE_URL="http://localhost:9000"
TEST_RESULTS=()

# Function to test a route
test_route() {
    local route="$1"
    local description="$2"
    local expected_content="$3"
    
    echo -n "Testing $description... "
    
    # Make request and capture response
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL$route")
    status_code=$(echo "$response" | tail -n1)
    content=$(echo "$response" | head -n -1)
    
    # Check status code
    if [ "$status_code" != "200" ] && [ "$status_code" != "302" ] && [ "$status_code" != "303" ]; then
        echo "❌ FAIL (Status: $status_code)"
        TEST_RESULTS+=("FAIL: $description - Status: $status_code")
        return 1
    fi
    
    # Check for blank/empty content
    if [ -z "$content" ] || [ "$(echo "$content" | wc -c)" -lt 100 ]; then
        echo "❌ FAIL (Blank/minimal content)"
        TEST_RESULTS+=("FAIL: $description - Blank page")
        return 1
    fi
    
    # Check for expected content if provided
    if [ -n "$expected_content" ]; then
        if echo "$content" | grep -q "$expected_content"; then
            echo "✅ PASS"
            TEST_RESULTS+=("PASS: $description")
            return 0
        else
            echo "❌ FAIL (Missing expected content: $expected_content)"
            TEST_RESULTS+=("FAIL: $description - Missing content")
            return 1
        fi
    fi
    
    echo "✅ PASS"
    TEST_RESULTS+=("PASS: $description")
    return 0
}

# Start server in background
echo "🚀 Starting server..."
./bin/server &
SERVER_PID=$!
sleep 3

# Wait for server to be ready
echo "⏳ Waiting for server to be ready..."
for i in {1..10}; do
    if curl -s "$BASE_URL" > /dev/null 2>&1; then
        echo "✅ Server is ready!"
        break
    fi
    sleep 1
    if [ $i -eq 10 ]; then
        echo "❌ Server failed to start"
        kill $SERVER_PID 2>/dev/null
        exit 1
    fi
done

echo ""
echo "🧪 Running Tests..."
echo "=================="

# Test critical routes that were previously problematic
echo ""
echo "📋 Testing Equipment Package Routes (Previously Blank):"
test_route "/workflow/packages" "Equipment Packages List" "Equipment Packages"
test_route "/workflow/packages/new" "New Equipment Package Form" "New Equipment Package"

echo ""
echo "📦 Testing Case Management Routes:"
test_route "/cases" "Cases List" "Cases"
test_route "/cases/new" "New Case Form" "Case"
test_route "/cases/management" "Case Management" "Case Management"

echo ""
echo "👥 Testing Customer Routes:"
test_route "/customers" "Customers List" "Customers"
test_route "/customers/new" "New Customer Form" "Customer"

echo ""
echo "🖥️ Testing Device Routes:"
test_route "/devices" "Devices List" "Devices"
test_route "/devices/new" "New Device Form" "Device"

echo ""
echo "💼 Testing Job Routes:"
test_route "/jobs" "Jobs List" "Jobs"
test_route "/jobs/new" "New Job Form" "Job"

echo ""
echo "📄 Testing Document Routes:"
test_route "/documents" "Documents List" "Documents"
test_route "/documents/upload" "Document Upload" "Upload"

echo ""
echo "💰 Testing Financial Routes:"
test_route "/financial" "Financial Dashboard" "Financial"
test_route "/financial/transactions" "Transactions List" "Transactions"

echo ""
echo "🔧 Testing Workflow Routes:"
test_route "/workflow/templates" "Job Templates" "Job Templates"
test_route "/workflow/templates/new" "New Job Template" "New Job Template"
test_route "/workflow/bulk" "Bulk Operations" "Bulk"

echo ""
echo "👤 Testing User Management Routes:"
test_route "/users" "Users List" "User Management"
test_route "/user-management/new" "New User Form" "User"

echo ""
echo "🔍 Testing Search Routes:"
test_route "/search/global" "Global Search" "Search"

echo ""
echo "📊 Testing Analytics Routes:"
test_route "/analytics" "Analytics Dashboard" "Analytics"

echo ""
echo "🔐 Testing Security Routes:"
test_route "/security/roles" "Security Roles" "Role"
test_route "/security/audit" "Security Audit" "Audit"

# Test some 404 scenarios
echo ""
echo "🚫 Testing 404 Handling:"
test_route "/nonexistent-page" "404 Page" "Error"
test_route "/workflow/invalid" "Invalid Workflow Route" "Error"

# Stop server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

# Display results
echo ""
echo "📊 Test Results Summary:"
echo "======================="
PASS_COUNT=0
FAIL_COUNT=0

for result in "${TEST_RESULTS[@]}"; do
    echo "$result"
    if [[ $result == PASS* ]]; then
        ((PASS_COUNT++))
    else
        ((FAIL_COUNT++))
    fi
done

echo ""
echo "📈 Final Score: $PASS_COUNT passed, $FAIL_COUNT failed"

if [ $FAIL_COUNT -eq 0 ]; then
    echo "🎉 ALL TESTS PASSED! No blank pages detected!"
    exit 0
else
    echo "⚠️  Some tests failed. Review the results above."
    exit 1
fi