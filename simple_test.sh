#!/bin/bash

BASE_URL="http://localhost:8080"

# Get token
echo "Getting token..."
TOKEN=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"attachtest2@example.com", "password":"password123"}' | \
  python3 -c "import sys, json; print(json.load(sys.stdin).get('access_token', 'FAIL'))")

if [ "$TOKEN" = "FAIL" ]; then
  echo "Failed to get token"
  exit 1
fi

echo "Token: ${TOKEN:0:30}..."

# Create project with unique key
RAND=$RANDOM
PROJECT_KEY="T$RAND"
echo "Creating project with key: $PROJECT_KEY"
PROJECT=$(curl -s -X POST $BASE_URL/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"name\":\"Test$RAND\",\"key\":\"$PROJECT_KEY\",\"description\":\"Test\"}")

PROJECT_ID=$(echo $PROJECT | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', 'FAIL'))" 2>/dev/null || echo "FAIL")

if [ "$PROJECT_ID" = "FAIL" ]; then
  echo "Failed to create project: $PROJECT"
  exit 1
fi

echo "Project ID: $PROJECT_ID"

# Create issue
echo "Creating issue..."
ISSUE=$(curl -s -X POST "$BASE_URL/api/v1/projects/$PROJECT_ID/issues" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Test Issue","description":"Test"}')

ISSUE_ID=$(echo $ISSUE | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', 'FAIL'))" 2>/dev/null || echo "FAIL")

if [ "$ISSUE_ID" = "FAIL" ]; then
  echo "Failed to create issue: $ISSUE"
  exit 1
fi

echo "Issue ID: $ISSUE_ID"
echo ""

# Test 1: Upload .exe file (should fail)
echo "=== Test 1: Uploading .exe file (should be blocked) ==="
echo "malicious content" > /tmp/test.exe
RESULT=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/test.exe")
echo "Response: $RESULT"
if echo "$RESULT" | grep -q "blocked\|not allowed"; then
  echo "✓ PASS: .exe blocked correctly"
else
  echo "✗ FAIL: .exe should have been blocked"
fi
echo ""

# Test 2: Upload .txt file (should succeed)
echo "=== Test 2: Uploading .txt file (should succeed) ==="
echo "valid content" > /tmp/test.txt
RESULT=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/test.txt")
echo "Response: $RESULT"
if echo "$RESULT" | grep -q '"id"'; then
  echo "✓ PASS: .txt uploaded successfully"
else
  echo "✗ FAIL: .txt should have been allowed"
fi

echo ""
echo "=== Tests Complete ==="
