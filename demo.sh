#!/bin/bash

echo "=== OTP Basic Demo ==="
echo

# Build the project
echo "Building project..."
make build
echo

# Start server in background
echo "Starting OTP server..."
./bin/otp-server &
SERVER_PID=$!

# Wait for server to start
sleep 2
echo "Server started with PID: $SERVER_PID"
echo

# Test registration
echo "Testing registration..."
REGISTER_RESPONSE=$(curl -s -X POST "http://localhost:8080/register" \
  -H "Content-Type: application/json" \
  -d '{
    "issuer": "OTP Demo",
    "account_name": "demo@example.com"
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

# Generate OTP using Go
echo "Generating OTP..."
OTP=$(go run -c 'package main
import (
    "fmt"
    "github.com/pquerna/otp/totp"
    "time"
)
func main() {
    secret := "'$SECRET'"
    code, _ := totp.GenerateCode(secret, time.Now())
    fmt.Print(code)
}' 2>/dev/null)

if [ -z "$OTP" ]; then
    echo "Using test OTP (Go not available in this context)"
    OTP="123456"
fi

echo "Generated OTP: $OTP"
echo

# Test OTP validation
echo "Testing OTP validation..."
VALIDATE_RESPONSE=$(curl -s -X POST "http://localhost:8080/validate-otp" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"otp\": \"$OTP\"
  }")

echo "Validation response:"
echo "$VALIDATE_RESPONSE" | jq '.'
echo

# Test protected endpoint
echo "Testing protected endpoint..."
STATUS_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/status" \
  -H "X-User-ID: $USER_ID" \
  -H "X-OTP: $OTP")

echo "Status response:"
echo "$STATUS_RESPONSE" | jq '.'
echo

# Cleanup
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null
echo "Server stopped"
echo

echo "=== Demo completed ==="
echo
echo "To use the interactive client, run:"
echo "  make run-client"
echo
echo "To start the server manually, run:"
echo "  make run-server"
