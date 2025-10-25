# OTP Basic - REST API Server with OTP Authentication

A secure REST API server in Go that requires OTP (One-Time Password) authentication for all protected endpoints. Includes a separate client for OTP generation and testing.

## Features

- **Master Token Registration**: Secure registration of master tokens with TOTP secret generation
- **OTP Authentication**: All protected endpoints require valid OTP codes
- **TOTP Support**: Time-based One-Time Passwords using RFC 6238 standard
- **QR Code Generation**: Automatic QR code URL generation for easy setup with authenticator apps
- **RESTful API**: Clean REST API design with proper HTTP status codes
- **Client Application**: Separate client for OTP generation and API testing

## Project Structure

```
otp-basic/
├── server/
│   └── main.go                 # Server entry point
├── client/
│   └── main.go                 # Client application
├── internal/
│   ├── server/
│   │   └── server.go           # Server setup and routing
│   ├── auth/
│   │   ├── auth.go             # Authentication manager
│   │   └── middleware.go       # OTP middleware
│   └── handlers/
│       └── handlers.go         # API handlers
├── go.mod                      # Go module definition
├── Makefile                    # Build and run commands
└── README.md                   # This file
```

## Installation

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd otp-basic
   make deps
   ```

2. **Build the applications**:
   ```bash
   make build
   ```

## Usage

### Starting the Server

```bash
# Run server directly
make run-server

# Or run in background
make run-server-bg
```

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable:

```bash
PORT=9090 make run-server
```

### Using the Client

```bash
# Start the client
make run-client

# Or specify a different server URL
./bin/otp-client http://localhost:9090
```

### Client Commands

The client provides an interactive interface with the following commands:

1. **`register <issuer> <account_name>`** - Register a new master token
2. **`generate`** - Generate OTP for the current user
3. **`validate <user_id> <otp>`** - Validate an OTP code
4. **`status`** - Get protected status (requires valid OTP)
5. **`data`** - Get protected data (requires valid OTP)
6. **`quit`** - Exit the client

### Example Workflow

1. **Start the server**:
   ```bash
   make run-server-bg
   ```

2. **Start the client**:
   ```bash
   make run-client
   ```

3. **Register a new user**:
   ```
   > register MyApp user@example.com
   ```

4. **Generate OTP**:
   ```
   > generate
   ```

5. **Test protected endpoints**:
   ```
   > status
   > data
   ```

## API Endpoints

### Public Endpoints

#### POST `/register`
Register a new master token and get OTP setup information.

**Request Body**:
```json
{
  "issuer": "MyApp",
  "account_name": "user@example.com"
}
```

**Response**:
```json
{
  "master_token": {
    "id": "uuid",
    "secret": "base32-secret",
    "created_at": "2023-01-01T00:00:00Z",
    "is_active": true
  },
  "qr_code_url": "otpauth://totp/...",
  "secret": "base32-secret"
}
```

#### POST `/validate-otp`
Validate an OTP code.

**Request Body**:
```json
{
  "user_id": "uuid",
  "otp": "123456"
}
```

**Response**:
```json
{
  "valid": true
}
```

### Protected Endpoints

All protected endpoints require OTP authentication via headers or JSON body.

**Authentication Methods**:
- **Headers**: `X-User-ID` and `X-OTP`
- **JSON Body**: `{"user_id": "uuid", "otp": "123456"}`

#### GET `/api/status`
Get authentication status.

**Response**:
```json
{
  "status": "authenticated",
  "user_id": "uuid",
  "created_at": "2023-01-01T00:00:00Z",
  "is_active": true,
  "timestamp": "2023-01-01T00:00:00Z"
}
```

#### GET `/api/protected-data`
Get protected data (example endpoint).

**Response**:
```json
{
  "message": "This is protected data",
  "user_id": "uuid",
  "data": {
    "secret_info": "This information is only accessible with valid OTP",
    "timestamp": "2023-01-01T00:00:00Z"
  }
}
```

## Security Features

- **TOTP Standard**: Uses RFC 6238 compliant TOTP implementation
- **Secure Secret Generation**: Cryptographically secure random secret generation
- **Time-based Validation**: OTP codes are valid for 30 seconds
- **No Password Storage**: Only OTP secrets are stored, no passwords
- **Master Token System**: Each user has a unique master token for OTP generation

## Dependencies

- **Gin**: HTTP web framework
- **Pquerna OTP**: TOTP implementation
- **Google UUID**: UUID generation
- **Golang Crypto**: Cryptographic functions

## Development

### Running Tests
```bash
make test
```

### Building
```bash
make build
```

### Cleaning
```bash
make clean
```

## Configuration

The server can be configured using environment variables:

- `PORT`: Server port (default: 8080)
- `GIN_MODE`: Gin mode (default: release)

## Troubleshooting

### Common Issues

1. **OTP validation fails**: Ensure your system clock is synchronized
2. **Connection refused**: Make sure the server is running on the correct port
3. **Invalid secret**: Use the exact secret returned during registration

### Debug Mode

To run the server in debug mode:
```bash
GIN_MODE=debug make run-server
```

## License

This project is open source and available under the MIT License.
