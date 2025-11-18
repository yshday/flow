#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== File Attachment API Test  ==="
echo ""

# Step 1: Register a user
echo "1. Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "attachtest2@example.com", "password": "password123", "username": "attachtest2"}')
echo "Register response: $REGISTER_RESPONSE"

# Step 1.5: Login to get token
echo "1.5. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "attachtest2@example.com", "password": "password123"}')
echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['access_token'])")
echo "Token: ${TOKEN:0:20}..."
echo ""

# Step 2: Create a project
echo "2. Creating project..."
PROJECT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "Attachment Test Project", "key": "ATP", "description": "Test project for attachments"}')
echo "Project response: $PROJECT_RESPONSE"
PROJECT_ID=$(echo $PROJECT_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "Project ID: $PROJECT_ID"
echo ""

# Step 3: Create an issue
echo "3. Creating issue..."
ISSUE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/projects/$PROJECT_ID/issues" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title": "Test Issue for Attachments", "description": "This issue will have file attachments"}')
echo "Issue response: $ISSUE_RESPONSE"
ISSUE_ID=$(echo $ISSUE_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "Issue ID: $ISSUE_ID"
echo ""

# Step 4: Create a test file
echo "4. Creating test file..."
echo "This is a test file for attachment upload" > /tmp/test_attachment.txt
echo "Test file created at /tmp/test_attachment.txt"
echo ""

# Step 5: Upload file attachment
echo "5. Uploading file attachment..."
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/test_attachment.txt")
echo "Upload response: $UPLOAD_RESPONSE"
ATTACHMENT_ID=$(echo $UPLOAD_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', 'ERROR'))" 2>/dev/null || echo "ERROR")
echo "Attachment ID: $ATTACHMENT_ID"
echo ""

# Step 6: List attachments for the issue
echo "6. Listing attachments for issue $ISSUE_ID..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
  -H "Authorization: Bearer $TOKEN")
echo "List response: $LIST_RESPONSE"
echo ""

# Step 7: Download attachment
if [ "$ATTACHMENT_ID" != "ERROR" ]; then
  echo "7. Downloading attachment $ATTACHMENT_ID..."
  curl -s -X GET "$BASE_URL/api/v1/attachments/$ATTACHMENT_ID/download" \
    -H "Authorization: Bearer $TOKEN" \
    -o /tmp/downloaded_attachment.txt
  echo "Downloaded to /tmp/downloaded_attachment.txt"
  echo "Content: $(cat /tmp/downloaded_attachment.txt)"
  echo ""

  # Step 8: Delete attachment
  echo "8. Deleting attachment $ATTACHMENT_ID..."
  DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/v1/attachments/$ATTACHMENT_ID" \
    -H "Authorization: Bearer $TOKEN")
  echo "Delete response: $DELETE_RESPONSE"
  echo ""

  # Step 9: Verify deletion
  echo "9. Verifying deletion - listing attachments again..."
  LIST_RESPONSE2=$(curl -s -X GET "$BASE_URL/api/v1/issues/$ISSUE_ID/attachments" \
    -H "Authorization: Bearer $TOKEN")
  echo "List response after delete: $LIST_RESPONSE2"
  echo ""
fi

echo "=== Test Complete ==="
