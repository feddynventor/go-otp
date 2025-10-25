.PHONY: build-server build-client run-server run-client clean test

# Build the server
build-server:
	cd server && go build -buildvcs=false -o ../bin/otp-server .

# Build the client
build-client:
	cd client && go build -buildvcs=false -o ../bin/otp-client .

# Build both
build: build-server build-client

# Run the server
run-server: build-server
	./bin/otp-server

# Run the client
run-client: build-client
	./bin/otp-client

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run server in background
run-server-bg: build-server
	./bin/otp-server &

# Stop background server
stop-server:
	pkill -f otp-server

# Show help
help:
	@echo "Available targets:"
	@echo "  build-server    - Build the OTP server"
	@echo "  build-client    - Build the OTP client"
	@echo "  build          - Build both server and client"
	@echo "  run-server     - Run the OTP server"
	@echo "  run-client     - Run the OTP client"
	@echo "  run-server-bg  - Run the OTP server in background"
	@echo "  stop-server    - Stop the background server"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  deps           - Install dependencies"
	@echo "  help           - Show this help"
