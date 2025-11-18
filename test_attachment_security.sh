#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== File Attachment Security Test ==="
echo ""

# Step 1: Register a user
echo "1. Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "sectest@example.com", "password": "password123", "username": "sectest"}')
echo "Register response: $REGISTER_RESPONSE"

# Step 1.5: Login to get token
echo "1.5. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "sectest@example.com", "password": "password123"}')
echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['access_token'])")
echo "Token: ${TOKEN:0:20}..."
echo ""

# Step 2: Create a project
echo "2. Creating project..."
PROJECT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "Security Test Project", "key": "STP", "description": "Test project for security"}')
echo "Project response: $PROJECT_RESPONSE"
PROJECT_ID=$(echo $PROJECT_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "Project ID: $PROJECT_ID"
echo ""

# Step 3: Create an issue
echo "3. Creating issue..."
ISSUE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/projects/$PROJECT_ID/issues" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title": "Security Test Issue", "description": "This issue will test security"}')
echo "Issue response: $ISSUE_RESPONSE"
ISSUE_ID=$(echo $ISSUE_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "Issue ID: $ISSUE_ID"
echo ""

# Test 1: Try to upload a blocked file type (.exe)
echo "=== TEST 1: Blocked Extension (.exe) ==="
echo "Creating fake .exe file..."
echo "fake exe content" > /tmp/malicious.exe
echo "Attempting to upload .exe file (should fail)..."
UPLOAD_EXE=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/malicious.exe")
echo "Response: $UPLOAD_EXE"
if echo "$UPLOAD_EXE" | grep -q "not allowed\|blocked"; then
  echo "✓ PASS: .exe file was correctly blocked"
else
  echo "✗ FAIL: .exe file should have been blocked"
fi
echo ""

# Test 2: Try to upload a shell script (.sh)
echo "=== TEST 2: Blocked Extension (.sh) ==="
echo "Creating shell script..."
echo "#!/bin/bash\necho 'test'" > /tmp/malicious.sh
echo "Attempting to upload .sh file (should fail)..."
UPLOAD_SH=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/malicious.sh")
echo "Response: $UPLOAD_SH"
if echo "$UPLOAD_SH" | grep -q "not allowed\|blocked"; then
  echo "✓ PASS: .sh file was correctly blocked"
else
  echo "✗ FAIL: .sh file should have been blocked"
fi
echo ""

# Test 3: Upload valid file type (.txt)
echo "=== TEST 3: Valid File Type (.txt) ==="
echo "Creating valid .txt file..."
echo "This is a legitimate text file" > /tmp/valid.txt
echo "Attempting to upload .txt file (should succeed)..."
UPLOAD_TXT=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/valid.txt")
echo "Response: $UPLOAD_TXT"
if echo "$UPLOAD_TXT" | grep -q '"id"'; then
  echo "✓ PASS: .txt file was correctly uploaded"
  VALID_ATTACHMENT_ID=$(echo $UPLOAD_TXT | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', 'ERROR'))" 2>/dev/null || echo "ERROR")
else
  echo "✗ FAIL: .txt file should have been allowed"
fi
echo ""

# Test 4: Upload valid image (.png)
echo "=== TEST 4: Valid File Type (.png) ==="
echo "Creating fake .png file..."
echo "PNG fake data" > /tmp/test.png
echo "Attempting to upload .png file (should succeed)..."
UPLOAD_PNG=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/test.png")
echo "Response: $UPLOAD_PNG"
if echo "$UPLOAD_PNG" | grep -q '"id"'; then
  echo "✓ PASS: .png file was correctly uploaded"
else
  echo "✗ FAIL: .png file should have been allowed"
fi
echo ""

# Test 5: Test file size limit (if we can create large files)
echo "=== TEST 5: File Size Limit (10MB) ==="
echo "Creating 11MB file (should exceed limit)..."
dd if=/dev/zero of=/tmp/large_file.txt bs=1M count=11 2>/dev/null
echo "Attempting to upload 11MB file (should fail)..."
UPLOAD_LARGE=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/large_file.txt")
echo "Response: $UPLOAD_LARGE"
if echo "$UPLOAD_LARGE" | grep -q "exceeds\|size"; then
  echo "✓ PASS: Large file was correctly rejected"
else
  echo "✗ FAIL: Large file should have been rejected"
fi
echo ""

# Test 6: Test project membership validation
echo "=== TEST 6: Project Membership Validation ==="
echo "Creating second user (not member of project)..."
REGISTER2=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "nonmember@example.com", "password": "password123", "username": "nonmember"}')
LOGIN2=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "nonmember@example.com", "password": "password123"}')
TOKEN2=$(echo $LOGIN2 | python3 -c "import sys, json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null || echo "ERROR")

if [ "$TOKEN2" != "ERROR" ]; then
  echo "Second user token: ${TOKEN2:0:20}..."
  echo "Attempting to upload as non-member (should fail)..."
  UPLOAD_NONMEMBER=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
    -H "Authorization: Bearer $TOKEN2" \
    -F "file=@/tmp/valid.txt")
  echo "Response: $UPLOAD_NONMEMBER"
  if echo "$UPLOAD_NONMEMBER" | grep -q "access denied\|not a member"; then
    echo "✓ PASS: Non-member was correctly denied access"
  else
    echo "✗ FAIL: Non-member should have been denied access"
  fi
else
  echo "⚠ SKIP: Could not create second user for membership test"
fi
echo ""

# Cleanup
echo "=== Cleanup ==="
if [ "$VALID_ATTACHMENT_ID" != "ERROR" ]; then
  echo "Deleting test attachments..."
  curl -s -X DELETE "$BASE_URL/api/v1/attachments/$VALID_ATTACHMENT_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
fi
rm -f /tmp/malicious.exe /tmp/malicious.sh /tmp/valid.txt /tmp/test.png /tmp/large_file.txt
echo "Cleanup complete"
echo ""

echo "=== Security Test Complete ==="
