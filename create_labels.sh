#!/bin/bash
cd /Users/ysh/dev/flow

# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ov_tf1@naver.com","password":"1234"}' | jq -r '.access_token')

echo "Token: ${TOKEN:0:20}..."

# Get project ID for THREAD
PROJECT_ID=$(curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/projects | jq -r '.[] | select(.key=="THREAD") | .id')

echo "Project ID: $PROJECT_ID"

# Create remaining labels
echo "Creating iOS label..."
curl -s -X POST "http://localhost:8080/api/v1/projects/$PROJECT_ID/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"iOS","color":"#f97316"}' | jq '.name'

echo "Creating Android label..."
curl -s -X POST "http://localhost:8080/api/v1/projects/$PROJECT_ID/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Android","color":"#22c55e"}' | jq '.name'

echo "Creating Design label..."
curl -s -X POST "http://localhost:8080/api/v1/projects/$PROJECT_ID/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Design","color":"#a855f7"}' | jq '.name'

echo "Creating API label..."
curl -s -X POST "http://localhost:8080/api/v1/projects/$PROJECT_ID/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"API","color":"#eab308"}' | jq '.name'

echo "Done!"
