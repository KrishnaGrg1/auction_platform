# Auction Platform

A modern auction platform built with Go and ConnectRPC.

## Prerequisites

- Go 1.25.0 or later
- Buf CLI (`brew install bufbuild/buf/buf` or see [buf.build](https://buf.build))
- Protocol Buffers compiler (optional, buf handles code generation)

## Setup

1. **Install dependencies:**
   ```bash
   make deps
   ```

2. **Generate ConnectRPC code:**
   ```bash
   make generate
   ```

3. **Run the server:**
   ```bash
   make run
   ```

The server will start on port 8080 (or the port specified in `PORT` environment variable).

## Project Structure

```
.
├── auction_platform/       # Proto definitions
│   └── v1/
│       └── auth.proto     # Auth service definitions
├── cmd/
│   └── server/            # Server entry point
│       └── main.go
├── internal/
│   ├── db/                # Database queries (sqlc)
│   └── services/          # Service implementations
│       └── auth_service.go
├── gen/                   # Generated code (gitignored)
│   └── auction_platform/
│       └── v1/
├── migrations/            # Database migrations
├── scripts/               # Utility scripts
├── buf.yaml              # Buf configuration
├── buf.gen.yaml          # Code generation config
└── Makefile              # Build targets

```

## Available Commands

- `make help` - Show available commands
- `make generate` - Generate ConnectRPC code from proto files
- `make lint` - Lint proto files
- `make build` - Build the server binary
- `make run` - Run the server
- `make clean` - Clean generated files
- `make test` - Run tests
- `make buf-update` - Update buf dependencies

## API Services

### AuthService

The authentication service provides:

- **Register** - User registration with email/password
  - Validates: first_name (1-50 chars), last_name (1-50 chars), email, password (8-100 chars)
  
- **Verify** - Email verification with 6-digit code
  - Validates: user_id (UUID), code (6 digits)
  
- **Login** - User login
  - Validates: email, password (8-100 chars)

## ConnectRPC Details

This project uses [ConnectRPC](https://connectrpc.com/), which provides:

- **HTTP/1.1 and HTTP/2 support** - Works with curl, browsers, and gRPC clients
- **JSON and binary protocols** - Flexible serialization
- **Type-safe clients** - Generated TypeScript/Swift/Kotlin clients
- **Streaming support** - Server, client, and bidirectional streaming
- **Built-in validation** - Using protovalidate

## Testing the API

### Using curl:

```bash
# Register a new user
curl -X POST http://localhost:8080/auction_platform.v1.AuthService/Register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'

# Verify email
curl -X POST http://localhost:8080/auction_platform.v1.AuthService/Verify \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "123456"
  }'

# Login
curl -X POST http://localhost:8080/auction_platform.v1.AuthService/Login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

## Environment Variables

Create a `.env` file in the project root:

```env
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/auction_platform
JWT_SECRET=your-secret-key
```

## Development

When you modify proto files:

1. Update the `.proto` file in `auction_platform/v1/`
2. Run `make generate` to regenerate code
3. Update service implementations in `internal/services/`
4. Run `make lint` to check proto file quality

## Proto Validation

Input validation is handled automatically using protovalidate constraints in the `.proto` files:

- Email validation
- String length constraints
- UUID format validation
- Numeric patterns (verification codes)

No manual validation code needed!

## Next Steps

- [ ] Implement database layer (PostgreSQL with sqlc)
- [ ] Add JWT token generation and validation
- [ ] Implement email sending for verification codes
- [ ] Add password hashing (bcrypt)
- [ ] Add middleware for authentication
- [ ] Write unit and integration tests
- [ ] Add more services (Auction, Bid, Payment, etc.)
- [ ] Set up CI/CD pipeline
