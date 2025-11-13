#!/bin/bash

# First login to get a token
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "Login failed. Response:"
  echo $LOGIN_RESPONSE | jq .
  exit 1
fi

echo "Login successful. Token: ${TOKEN:0:20}..."
echo ""
echo "Fetching workout 13..."
curl -s -X GET "http://localhost:8080/api/workouts/13" \
  -H "Authorization: Bearer $TOKEN" | jq .
