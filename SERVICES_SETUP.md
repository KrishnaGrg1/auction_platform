# Services Setup Guide

## Architecture Overview

The auction platform consists of **two separate microservices**:

1. **Auth Service** (Port 8081) - User authentication, registration, verification
2. **Auction Service** (Port 8080) - Auction management, bidding, WebSocket real-time updates

Both services share:
- Same PostgreSQL database
- Same JWT secret for token validation
- ConnectRPC protocol (HTTP/2)

## Prerequisites

- Go 1.25+
- PostgreSQL 15+ (running and accessible)
- `.env` file configured (see below)

## Environment Configuration

### 1. Copy the example environment file

```bash
cp .env.example .env
```

### 2. Edit `.env` with your settings

```env
# Database Connection
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="postgres://username:password@localhost:5432/auction_platform"
GOOSE_MIGRATION_DIR=./migrations

# Service Ports (customize if needed)
AUTH_PORT=8081
AUCTION_PORT=8080

# Authentication
JWT_SECRET="your-super-secret-jwt-key-change-this-in-production"

# Email (Optional - for OTP verification emails)
RESEND_API_KEY=""
RESEND_EMAIL_FROM=""

# Redis (Optional - for future features)
REDIS_URL="redis://localhost:6379"
```

**Important**: Change `JWT_SECRET` to a strong random string in production!

## Database Setup

### 1. Create the database

```bash
createdb auction_platform
```

Or via PostgreSQL:
```sql
CREATE DATABASE auction_platform;
```

### 2. Run migrations

```bash
goose up
```

This will create all necessary tables:
- users
- auctions
- bids
- winners
- auction_history
- otps

## Running the Services

You need to run **both services** in separate terminal windows.

### Terminal 1: Auth Service

```bash
cd services/auth
go run main.go
```

You should see:
```
Connected to Neon database
🔐 Auth service running on port: 8081
🔗 gRPC endpoint: http://localhost:8081
📍 Endpoints:
   - POST /auction_platform.v1.AuthService/Register
   - POST /auction_platform.v1.AuthService/Login
   - POST /auction_platform.v1.AuthService/Verify
```

### Terminal 2: Auction Service

```bash
cd services/auction
go run main.go
```

You should see:
```
Connected to Neon database
🚀 Auction service running on port: 8080
📡 WebSocket endpoint: ws://localhost:8080/ws/auction?auction_id=<auction_id>
🔗 gRPC endpoint: http://localhost:8080
```

## Service Endpoints

### Auth Service (Port 8081)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auction_platform.v1.AuthService/Register` | Create new user account | ❌ No |
| POST | `/auction_platform.v1.AuthService/Login` | Login with email/password | ❌ No |
| POST | `/auction_platform.v1.AuthService/Verify` | Verify email with OTP | ❌ No |

### Auction Service (Port 8080)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auction_platform.v1.AuctionService/CreateAuction` | Create new auction | ✅ Yes |
| POST | `/auction_platform.v1.AuctionService/BidAuction` | Place a bid | ✅ Yes |
| POST | `/auction_platform.v1.AuctionService/GetAuctionDetailsById` | Get auction details | ✅ Yes |
| POST | `/auction_platform.v1.AuctionService/GetAuctionsList` | List all auctions | ✅ Yes |
| POST | `/auction_platform.v1.AuctionService/GetUserAuctions` | Get user's auctions | ✅ Yes |
| POST | `/auction_platform.v1.AuctionService/EndAuction` | End an auction | ✅ Yes |
| POST | `/auction_platform.v1.AuctionService/CancelAuction` | Cancel an auction | ✅ Yes |
| WS | `/ws/auction?auction_id=<uuid>` | WebSocket for real-time updates | ❌ No* |

*WebSocket doesn't require auth yet (see Production section)

## Testing the Services

### 1. Register a User

```bash
curl -X POST http://localhost:8081/auction_platform.v1.AuthService/Register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

Response:
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "test@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "available_balance": 0,
    "held_balance": 0,
    "is_verified": false,
    ...
  },
  "message": "Registered successfully"
}
```

**Note**: Save the OTP code from the response (or check your email if configured).

### 2. Verify Email

```bash
curl -X POST http://localhost:8081/auction_platform.v1.AuthService/Verify \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "code": "123456"
  }'
```

Response includes JWT token:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { ... },
  "message": "User verified successfully"
}
```

**Save this token!** You'll need it for auction operations.

### 3. Login (Alternative to Verify)

```bash
curl -X POST http://localhost:8081/auction_platform.v1.AuthService/Login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 4. Create an Auction

```bash
curl -X POST http://localhost:8080/auction_platform.v1.AuctionService/CreateAuction \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{
    "title": "Vintage Watch",
    "description": "Beautiful vintage watch from 1960s",
    "type": "AUCTION_TYPE_ENGLISH",
    "starting_price": 10000,
    "reserved_price": 50000,
    "extend_on_bid": true,
    "extend_minutes": 5,
    "start_time": "2026-06-15T12:00:00Z",
    "end_time": "2026-06-15T18:00:00Z"
  }'
```

**Save the auction_id from the response!**

### 5. Connect to WebSocket

Open `websocket_test.html` in your browser:

```bash
open websocket_test.html
```

1. Enter the auction ID from step 4
2. Click "Connect"
3. You should see: `🟢 Connected`

### 6. Place a Bid

```bash
curl -X POST http://localhost:8080/auction_platform.v1.AuctionService/BidAuction \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{
    "auction_id": "YOUR_AUCTION_ID",
    "amount": 15000,
    "is_auto_bid": false
  }'
```

**Watch the WebSocket client!** It should instantly show the new bid event.

## Service Communication

```
┌─────────────────┐         ┌─────────────────┐
│   Auth Service  │         │ Auction Service │
│   Port: 8081    │         │   Port: 8080    │
└────────┬────────┘         └────────┬────────┘
         │                           │
         │    Same Database          │
         └────────┬──────────────────┘
                  │
         ┌────────▼────────┐
         │   PostgreSQL    │
         └─────────────────┘
```

### Flow:
1. User registers/logs in via Auth Service → Gets JWT token
2. User uses JWT token to call Auction Service
3. Auction Service validates JWT (same secret)
4. Both services read/write to same database
5. Auction Service broadcasts events via WebSocket

## Health Checks

### Check Auth Service
```bash
curl http://localhost:8081/health
# Response: OK
```

### Check Auction Service
```bash
curl http://localhost:8080/health
# Response: OK
```

## Production Deployment

### 1. Environment Variables

Set in your production environment (not in .env file):

```bash
export AUTH_PORT=8081
export AUCTION_PORT=8080
export JWT_SECRET="production-secret-at-least-32-chars-long"
export GOOSE_DBSTRING="postgres://user:pass@prod-db:5432/auction_platform"
```

### 2. Build Binaries

```bash
# Build auth service
go build -o bin/auth-service ./services/auth/main.go

# Build auction service
go build -o bin/auction-service ./services/auction/main.go
```

### 3. Run as Services

**Using systemd (Linux):**

Create `/etc/systemd/system/auction-auth.service`:
```ini
[Unit]
Description=Auction Platform Auth Service
After=network.target postgresql.service

[Service]
Type=simple
User=auction
WorkingDirectory=/opt/auction-platform
EnvironmentFile=/opt/auction-platform/.env
ExecStart=/opt/auction-platform/bin/auth-service
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/auction-service.service`:
```ini
[Unit]
Description=Auction Platform Service
After=network.target postgresql.service

[Service]
Type=simple
User=auction
WorkingDirectory=/opt/auction-platform
EnvironmentFile=/opt/auction-platform/.env
ExecStart=/opt/auction-platform/bin/auction-service
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Start services:
```bash
sudo systemctl enable auction-auth
sudo systemctl enable auction-service
sudo systemctl start auction-auth
sudo systemctl start auction-service
```

### 4. Reverse Proxy (Nginx)

```nginx
# Auth Service
upstream auth {
    server localhost:8081;
}

# Auction Service
upstream auction {
    server localhost:8080;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    # Auth endpoints
    location /auction_platform.v1.AuthService/ {
        proxy_pass http://auth;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Auction endpoints
    location /auction_platform.v1.AuctionService/ {
        proxy_pass http://auction;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket
    location /ws/ {
        proxy_pass http://auction;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }
}
```

### 5. Docker Compose (Optional)

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: auction_platform
      POSTGRES_USER: auction
      POSTGRES_PASSWORD: secretpassword
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  auth-service:
    build:
      context: .
      dockerfile: services/auth/Dockerfile
    environment:
      AUTH_PORT: 8081
      GOOSE_DBSTRING: postgres://auction:secretpassword@postgres:5432/auction_platform
      JWT_SECRET: production-secret-here
    ports:
      - "8081:8081"
    depends_on:
      - postgres

  auction-service:
    build:
      context: .
      dockerfile: services/auction/Dockerfile
    environment:
      AUCTION_PORT: 8080
      GOOSE_DBSTRING: postgres://auction:secretpassword@postgres:5432/auction_platform
      JWT_SECRET: production-secret-here
    ports:
      - "8080:8080"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

## Troubleshooting

### "Connection refused" on Auth Service

**Problem**: Cannot connect to port 8081

**Solutions**:
1. Check if auth service is running: `lsof -i :8081`
2. Check environment variable: `echo $AUTH_PORT`
3. Check firewall rules
4. Look at service logs

### "Connection refused" on Auction Service

**Problem**: Cannot connect to port 8080

**Solutions**:
1. Check if auction service is running: `lsof -i :8080`
2. Check environment variable: `echo $AUCTION_PORT`
3. Check firewall rules
4. Look at service logs

### "Invalid token" errors

**Problem**: JWT validation fails

**Solutions**:
1. Ensure both services use the SAME `JWT_SECRET`
2. Check token hasn't expired (default: 24 hours)
3. Verify token format: `Bearer <token>`
4. Re-login to get a fresh token

### "User not verified"

**Problem**: Cannot login

**Solutions**:
1. Complete email verification first
2. Check OTP in database: `SELECT * FROM otps WHERE user_id = 'your-id';`
3. OTP expires after 7 days
4. Re-register if needed

### Database connection fails

**Problem**: "Cannot connect to database"

**Solutions**:
1. Check PostgreSQL is running: `pg_isready`
2. Verify connection string in `.env`
3. Test connection: `psql $GOOSE_DBSTRING`
4. Check database exists: `\l` in psql
5. Run migrations: `goose up`

### WebSocket won't connect

**Problem**: WebSocket connection fails

**Solutions**:
1. Ensure auction service is running
2. Check auction_id is valid UUID
3. Use `ws://` not `wss://` for local
4. Check browser console for errors
5. Verify no firewall blocking WebSocket

## Monitoring

### Log Locations

Development:
- Auth Service: stdout
- Auction Service: stdout

Production (systemd):
```bash
# Auth service logs
journalctl -u auction-auth -f

# Auction service logs
journalctl -u auction-service -f
```

### Metrics to Monitor

1. **Service Health**
   - HTTP 200 on `/health` endpoints
   - Response time < 100ms

2. **Database**
   - Connection pool usage
   - Query performance
   - Active connections

3. **WebSocket**
   - Active connections
   - Connection duration
   - Message throughput

4. **Business Metrics**
   - User registrations/hour
   - Active auctions
   - Bids per minute
   - Failed transactions

## Security Checklist

- [ ] Change `JWT_SECRET` from default
- [ ] Use strong database password
- [ ] Enable PostgreSQL SSL in production
- [ ] Add rate limiting (nginx/middleware)
- [ ] Add CORS configuration
- [ ] Enable HTTPS/TLS (wss:// for WebSocket)
- [ ] Add authentication to WebSocket
- [ ] Implement request validation
- [ ] Add audit logging
- [ ] Set up monitoring/alerting
- [ ] Regular security updates
- [ ] Database backups

## Further Reading

- [WEBSOCKET_GUIDE.md](WEBSOCKET_GUIDE.md) - Complete WebSocket documentation
- [QUICK_START.md](QUICK_START.md) - Quick tutorial
- [README.md](README.md) - Project overview
- [HOW_TO_BID.md](HOW_TO_BID.md) - Bidding guide

---

**Last Updated**: June 15, 2026  
**Services Version**: 1.0.0
