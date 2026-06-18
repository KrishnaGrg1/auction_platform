# 🔄 How to Restart Services After Code Changes

## The Issue

When you make code changes, you **must restart the running services** for the changes to take effect. The error you're seeing indicates the old code is still running.

## Quick Restart Guide

### Step 1: Stop All Running Services

**Option A: Find and kill the processes**
```bash
# Find auth service
lsof -ti:8081 | xargs kill -9

# Find auction service  
lsof -ti:8080 | xargs kill -9
```

**Option B: If you see them running in terminals**
- Press `Ctrl + C` in each terminal window running the services

### Step 2: Restart Services with New Code

**Terminal 1 - Auth Service:**
```bash
cd services/auth
go run main.go
```

Wait for:
```
🔐 Auth service running on port: 8081
```

**Terminal 2 - Auction Service:**
```bash
cd services/auction
go run main.go
```

Wait for:
```
🚀 Auction service running on port: 8080
📡 WebSocket endpoint: ws://localhost:8080/ws/auction?auction_id=<auction_id>
```

### Step 3: Test Again

Now try your request:

```bash
curl -X POST http://localhost:8080/auction_platform.v1.AuctionService/CreateAuction \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Test Auction",
    "description": "Testing WebSocket",
    "type": "AUCTION_TYPE_ENGLISH",
    "starting_price": 10000,
    "reserved_price": 50000,
    "extend_on_bid": true,
    "extend_minutes": 5,
    "start_time": "2026-06-15T12:00:00Z",
    "end_time": "2026-06-15T13:00:00Z"
  }'
```

## Pro Tips

### Use a Process Manager (Recommended)

Instead of manually stopping/starting, use a tool like `air` for auto-reload:

**Install air:**
```bash
go install github.com/cosmtrek/air@latest
```

**Create `.air.toml` in project root:**
```toml
[build]
  cmd = "go build -o ./tmp/auth ./services/auth/main.go"
  bin = "./tmp/auth"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor", "gen"]
  
[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"
```

**Run with air:**
```bash
# Auth service (in one terminal)
cd services/auth && air

# Auction service (in another terminal)
cd services/auction && air
```

Now services auto-restart when you change `.go` files!

### Build Binaries (Production-like)

Instead of `go run`, build binaries:

```bash
# Build both services
go build -o bin/auth-service ./services/auth/main.go
go build -o bin/auction-service ./services/auction/main.go

# Run them
./bin/auth-service      # Terminal 1
./bin/auction-service   # Terminal 2
```

This is closer to production and slightly faster.

### Check Running Services

**See what's listening on your ports:**
```bash
# Check both ports
lsof -i:8080 -i:8081

# Or individually
lsof -i:8080  # Auction service
lsof -i:8081  # Auth service
```

**Quick health check:**
```bash
curl http://localhost:8081/health && echo " - Auth OK"
curl http://localhost:8080/health && echo " - Auction OK"
```

## Common Mistakes

### ❌ Mistake 1: Editing code but not restarting

**Symptom:** Changes don't take effect, old errors persist

**Fix:** Always restart after code changes!

### ❌ Mistake 2: Multiple instances running

**Symptom:** Weird behavior, sometimes works, sometimes doesn't

**Check:**
```bash
lsof -i:8080
# If you see multiple PIDs, kill them all
```

**Fix:**
```bash
lsof -ti:8080 | xargs kill -9
lsof -ti:8081 | xargs kill -9
# Then start ONE instance of each
```

### ❌ Mistake 3: Wrong directory

**Symptom:** "cannot find package" errors

**Fix:** Make sure you're in the right directory:
```bash
# For auth service
cd /path/to/auction_platform/services/auth
go run main.go

# For auction service  
cd /path/to/auction_platform/services/auction
go run main.go
```

### ❌ Mistake 4: Forgot to save the file

**Symptom:** Changes still don't appear after restart

**Fix:** Make sure you saved the file in your editor (Ctrl+S / Cmd+S)

## Verification Checklist

After restarting, verify:

- [ ] Auth service shows startup message on port 8081
- [ ] Auction service shows startup message on port 8080
- [ ] Health checks pass: `curl http://localhost:8081/health`
- [ ] Health checks pass: `curl http://localhost:8080/health`
- [ ] Only ONE process per port: `lsof -i:8080` and `lsof -i:8081`
- [ ] Database connection successful (check logs)

## Automated Restart Script

Create `restart.sh`:

```bash
#!/bin/bash

echo "🛑 Stopping services..."
lsof -ti:8080 | xargs kill -9 2>/dev/null
lsof -ti:8081 | xargs kill -9 2>/dev/null
sleep 1

echo "🔨 Building services..."
go build -o bin/auth-service ./services/auth/main.go
go build -o bin/auction-service ./services/auction/main.go

echo "✅ Built successfully!"
echo ""
echo "Start the services in separate terminals:"
echo "  Terminal 1: ./bin/auth-service"
echo "  Terminal 2: ./bin/auction-service"
```

Make executable:
```bash
chmod +x restart.sh
```

Use it:
```bash
./restart.sh
```

## Docker Compose (Alternative)

For a production-like setup, use Docker Compose (see SERVICES_SETUP.md).

Benefits:
- Services auto-restart on crash
- Easy to stop/start all at once
- Consistent environment

```bash
docker-compose up --build  # Rebuild and start
docker-compose down        # Stop all
docker-compose logs -f     # View logs
```

---

## TL;DR - Quick Fix

**Your services are running old code. Restart them:**

```bash
# Kill old processes
lsof -ti:8080 | xargs kill -9
lsof -ti:8081 | xargs kill -9

# Start auth service (Terminal 1)
cd services/auth && go run main.go

# Start auction service (Terminal 2)
cd services/auction && go run main.go

# Test again
curl -X POST http://localhost:8080/auction_platform.v1.AuctionService/CreateAuction \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "type": "AUCTION_TYPE_ENGLISH", ... }'
```

✅ **It will work after restart!**

---

**Remember:** Code changes = Restart services! 🔄
