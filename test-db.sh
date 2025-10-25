#!/bin/bash

# Test script for PostgreSQL integration
echo "=== Testing OTP Basic with PostgreSQL ==="

# Start PostgreSQL
echo "Starting PostgreSQL..."
make db-up

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 10

# Test database connection
echo "Testing database connection..."
docker exec otp-postgres psql -U postgres -d otp_basic -c "SELECT version();"

# Build and start server
echo "Building server..."
make build

echo "Starting server in background..."
./bin/otp-server &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test client
echo "Testing client..."
echo "register TestApp test@example.com" | ./bin/otp-client

# Cleanup
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
make db-down

echo "Test completed!"
