# Auction Platform

A real-time auction platform supporting **English** and **Dutch** auctions, built with Go, ConnectRPC, PostgreSQL, and WebSocket (planned).

## Features

- 🔐 **Authentication** - JWT-based auth with email verification
- 🎯 **English Auctions** - Traditional ascending-bid auctions with time extensions
- 🇳🇱 **Dutch Auctions** - Descending-price auctions that end on first bid
- 💰 **Balance Management** - User wallets with available/held balance tracking
- 📊 **Bid History** - Complete audit trail of all auction events
- 🔒 **Transaction Safety** - ACID-compliant with row-level locking

## Prerequisites

- **Go 1.25+**
- **PostgreSQL 15+**
- **buf** - `go install github.com/bufbuild/buf/cmd/buf@latest`
- **sqlc** - `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- **goose** - `go install github.com/pressly/goose/v3/cmd/goose@latest`

## Quick Setup

```bash
# 1. Clone and install dependencies
git clone <repo-url>
cd auction_platform
go mod download

# 2. Set up environment variables
cp .env.example .env
# Edit .env with your database credentials

# 3. Run database migrations
goose up

# 4. Generate code
buf generate
sqlc generate

# 5. Run the server
go run services/auction/main.go 
go run services/auth/main.go
```

Server starts on `http://localhost:8080`

## Project Structure

```
auction_platform/
├── auction_platform/v1/     # Proto definitions (SOURCE CODE - commit to git)
│   ├── auction.proto        # Auction service API
│   └── auth.proto           # Auth service API
├── gen/                     # Generated code (DO NOT EDIT, gitignored)
│   └── auction_platform/v1/ # Generated Go code from protos
├── internal/
│   ├── auth/                # JWT & middleware
│   ├── db/
│   │   ├── queries/         # SQL queries (SOURCE CODE)
│   │   └── sqlc/            # Generated database code
│   └── store/               # Database connection pool
├── services/                # Service implementations
│   ├── auth/                # Auth service handlers
│   └── auction/             # Auction service handlers
├── migrations/              # Database migrations (goose)
├── buf.yaml                 # Buf configuration
├── buf.gen.yaml             # Protobuf generation config
└── sqlc.yaml                # SQLC configuration
```

## Common Commands

### Code Generation

```bash
# Lint proto files
buf lint

# Generate protobuf code
buf generate

# Generate database code
sqlc generate
```

### Database Migrations

```bash
# Set your database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/auction_platform"

# Run all pending migrations
goose -dir migrations postgres "$DATABASE_URL" up

# Rollback last migration
goose -dir migrations postgres "$DATABASE_URL" down

# Create new migration
goose -dir migrations create migration_name sql
```

### Running & Testing

```bash
# Run server
go run cmd/server/main.go

# Run tests
go test ./...

# Run with race detector
go test -race ./...
```

## API Services

### AuthService

- **Register** - Create new user account
- **Login** - Authenticate and get JWT token  
- **Verify** - Verify email with OTP code

### AuctionService

- **CreateAuction** - Create new auction (English/Dutch)
- **BidAuction** - Place a bid on active auction
- **GetAuctionDetailsById** - Get auction details
- **GetAuctionsList** - List auctions with filters/pagination
- **GetUserAuctions** - Get user's auctions
- **EndAuction** - Manually end an auction
- **CancelAuction** - Cancel a scheduled auction


## Auction Types

### English Auction
- **Ascending price** - Bids must be higher than current price
- **Multiple bids** - Accepts bids until time expires
- **Time extension** - Can extend time on late bids
- **Winner** - Highest bidder when time runs out

### Dutch Auction  
- **Descending price** - Price drops at intervals
- **First bid wins** - Ends immediately on first bid
- **No extensions** - Single bid mechanism
- **Winner** - First person to accept the price

## Environment Variables

```env
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/auction_platform?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

## Development Workflow

```bash
# 1. Edit proto files in auction_platform/v1/
# 2. Lint proto files
buf lint

# 3. Generate protobuf code
buf generate

# 4. Edit SQL queries in internal/db/queries/
# 5. Generate database code
sqlc generate

# 6. Implement handlers in services/
# 7. Test
go test ./...

# 8. Run
go run cmd/server/main.go
```

## Important Notes

### About `auction_platform/v1/`
**DO NOT gitignore this directory!** These are your **source proto files**, not generated code. Only `gen/` is generated.

### About `buf.lock`
Keep `buf.lock` in version control for reproducible builds.

## Technology Stack

- **Go 1.25** - Backend language
- **ConnectRPC** - gRPC-compatible HTTP/JSON API
- **PostgreSQL 15+** - Database with pgx driver
- **buf** - Protobuf code generation
- **sqlc** - Type-safe SQL code generation
- **goose** - Database migrations
- **JWT** - Authentication tokens

## Roadmap

- [x] Auth service (register, login, verify)
- [x] English auction implementation
- [x] Dutch auction implementation  
- [x] Bid history tracking
- [x] User balance management
- [ ] WebSocket for real-time updates
- [ ] Scheduled auction start/end
- [ ] Dutch auction price drop scheduler
- [ ] Email notifications
- [ ] Payment integration
- [ ] Admin panel
- [ ] Web frontend

## License

MIT
