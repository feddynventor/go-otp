#!/bin/bash

# Example script demonstrating OTP API usage

SERVER_URL="http://localhost:8080"

echo "=== OTP API Example ==="
echo

# Start server in background
echo "Starting server..."
make run-server-bg &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Server started with PID: $SERVER_PID"
echo

# Register a new user
echo "1. Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$SERVER_URL/register" \
  -H "Content-Type: application/json" \
  -d '{
    "issuer": "OTP Example",
    "account_name": "test@example.com"
  }')

echo "Registration response:"
echo "$REGISTER_RESPONSE" | jq '.'
echo

# Extract user ID and secret
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.master_token.id')
SECRET=$(echo "$REGISTER_RESPONSE" | jq -r '.secret')

echo "User ID: $USER_ID"
echo "Secret: $SECRET"
echo

# Generate OTP using the secret
echo "2. Generating OTP..."
OTP=$(python3 -c "
import pyotp
import time
secret = '$SECRET'
totp = pyotp.TOTP(secret)
print(totp.now())
" 2>/dev/null)

if [ -z "$OTP" ]; then
    echo "Python pyotp not available, using a test OTP"
    OTP="123456"
fi

echo "Generated OTP: $OTP"
echo

# Validate OTP
echo "3. Validating OTP..."
VALIDATE_RESPONSE=$(curl -s -X POST "$SERVER_URL/validate-otp" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"otp\": \"$OTP\"
  }")

echo "Validation response:"
echo "$VALIDATE_RESPONSE" | jq '.'
echo

# Test protected endpoint with OTP
echo "4. Testing protected endpoint..."
STATUS_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/status" \
  -H "X-User-ID: $USER_ID" \
  -H "X-OTP: $OTP")

echo "Status response:"
echo "$STATUS_RESPONSE" | jq '.'
echo

# Test protected data endpoint
echo "5. Testing protected data endpoint..."
DATA_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/protected-data" \
  -H "X-User-ID: $USER_ID" \
  -H "X-OTP: $OTP")

echo "Protected data response:"
echo "$DATA_RESPONSE" | jq '.'
echo

# Cleanup
echo "6. Cleaning up..."
kill $SERVER_PID 2>/dev/null
echo "Server stopped"
echo

echo "=== Example completed ==="
