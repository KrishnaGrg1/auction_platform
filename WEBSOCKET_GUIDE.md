# WebSocket Implementation Guide

## Overview

The auction platform now includes a real-time WebSocket implementation that broadcasts auction events to connected clients. This allows users to receive live updates about bids, auction endings, and other auction-related events.

## Architecture

### Components

1. **WebSocket Hub** (`internal/socket/socket.go`)
   - Manages client connections and rooms
   - Handles message broadcasting to auction-specific rooms
   - Implements connection lifecycle management
   - Supports ping/pong for connection health

2. **Auction Service** (`services/auction/service/auction.go`)
   - Publishes events after bid/auction operations
   - Integrates with the WebSocket hub
   - Sends structured event messages

3. **Main Server** (`services/auction/main.go`)
   - Exposes the WebSocket endpoint at `/ws/auction`
   - Registers auction service handlers
   - Configures HTTP/2 support for Connect protocol

## WebSocket Endpoint

### Connection URL

```
ws://localhost:8080/ws/auction?auction_id=<uuid>&user_id=<uuid>
```

### Query Parameters

- `auction_id` (required): UUID of the auction to subscribe to
- `user_id` (optional): UUID of the user connecting (for future features)

### Example Connection

```javascript
const auctionId = '123e4567-e89b-12d3-a456-426614174000';
const ws = new WebSocket(`ws://localhost:8080/ws/auction?auction_id=${auctionId}`);

ws.onopen = () => {
    console.log('Connected to auction room');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Received event:', data);
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('Connection closed');
};
```

## Event Types

### 1. Connected Event

Sent immediately after successful connection.

```json
{
  "type": "connected",
  "auction_id": "123e4567-e89b-12d3-a456-426614174000",
  "message": "Successfully connected to auction room",
  "timestamp": "2026-06-15T12:00:00Z"
}
```

### 2. New Bid Event (English Auction)

Broadcast when a new bid is placed on an English auction.

```json
{
  "type": "new_bid",
  "auction_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "789e4567-e89b-12d3-a456-426614174000",
  "amount": 150000,
  "timestamp": "2026-06-15T12:05:30Z"
}
```

### 3. Auction Won Event (Dutch Auction)

Sent when someone accepts the current price in a Dutch auction.

```json
{
  "type": "auction_won",
  "auction_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "789e4567-e89b-12d3-a456-426614174000",
  "amount": 120000,
  "timestamp": "2026-06-15T12:10:15Z"
}
```

### 4. Auction Ended Event

Broadcast when an auction is ended (manually or after Dutch auction win).

```json
{
  "type": "auction_ended",
  "auction_id": "123e4567-e89b-12d3-a456-426614174000",
  "timestamp": "2026-06-15T12:10:15Z"
}
```

### 5. Ping/Pong Events

Automatic keep-alive messages sent every 30 seconds.

```json
{
  "type": "ping",
  "auction_id": "123e4567-e89b-12d3-a456-426614174000",
  "timestamp": "2026-06-15T12:00:30Z"
}
```

Client can respond with:

```json
{
  "type": "ping"
}
```

The server will respond with a pong event.

## Testing the WebSocket

### Option 1: HTML Test Client

Open the provided `websocket_test.html` file in your browser:

```bash
open websocket_test.html
```

Features:
- Connect/disconnect to auction rooms
- View real-time events
- Track statistics (total events, bids, connection time)
- Clear event history

### Option 2: Command Line with `websocat`

Install websocat:
```bash
brew install websocat  # macOS
# or
cargo install websocat  # Rust
```

Connect to an auction:
```bash
websocat "ws://localhost:8080/ws/auction?auction_id=YOUR_AUCTION_ID"
```

### Option 3: JavaScript Client

```javascript
class AuctionWebSocket {
    constructor(auctionId, onEvent) {
        this.auctionId = auctionId;
        this.onEvent = onEvent;
        this.ws = null;
    }

    connect() {
        const url = `ws://localhost:8080/ws/auction?auction_id=${this.auctionId}`;
        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
            console.log('Connected to auction', this.auctionId);
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.onEvent(data);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.ws.onclose = () => {
            console.log('Disconnected from auction');
        };
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    sendPing() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type: 'ping' }));
        }
    }
}

// Usage
const auction = new AuctionWebSocket('123e4567-e89b-12d3-a456-426614174000', (event) => {
    console.log('Event received:', event);
    
    switch(event.type) {
        case 'new_bid':
            console.log(`New bid: $${event.amount / 100}`);
            break;
        case 'auction_won':
            console.log(`Auction won by ${event.user_id}`);
            break;
        case 'auction_ended':
            console.log('Auction has ended');
            break;
    }
});

auction.connect();
```

## Integration with Frontend

### React Example

```jsx
import { useEffect, useState } from 'react';

function AuctionRoom({ auctionId }) {
    const [events, setEvents] = useState([]);
    const [connected, setConnected] = useState(false);
    const [ws, setWs] = useState(null);

    useEffect(() => {
        const websocket = new WebSocket(
            `ws://localhost:8080/ws/auction?auction_id=${auctionId}`
        );

        websocket.onopen = () => {
            setConnected(true);
        };

        websocket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            setEvents(prev => [data, ...prev]);
        };

        websocket.onclose = () => {
            setConnected(false);
        };

        setWs(websocket);

        return () => {
            websocket.close();
        };
    }, [auctionId]);

    return (
        <div>
            <h2>Auction Room: {auctionId}</h2>
            <p>Status: {connected ? '🟢 Connected' : '⚪ Disconnected'}</p>
            
            <div>
                {events.map((event, idx) => (
                    <div key={idx} className={`event ${event.type}`}>
                        <strong>{event.type}</strong>
                        {event.amount && <span> - ${event.amount / 100}</span>}
                        <span> - {new Date(event.timestamp).toLocaleTimeString()}</span>
                    </div>
                ))}
            </div>
        </div>
    );
}
```

## Implementation Details

### Hub Management

The hub manages multiple auction rooms, each containing multiple clients:

```
Hub
├── Room: auction-uuid-1
│   ├── Client 1
│   ├── Client 2
│   └── Client 3
├── Room: auction-uuid-2
│   ├── Client 4
│   └── Client 5
```

### Connection Lifecycle

1. **Connection**: Client connects with auction_id parameter
2. **Registration**: Hub adds client to the auction room
3. **Welcome**: Server sends "connected" event
4. **Live Updates**: Server broadcasts events to all room clients
5. **Keep-Alive**: Periodic ping/pong (every 30 seconds)
6. **Disconnection**: Hub removes client and cleans up empty rooms

### Broadcasting Strategy

- **Per-Auction Broadcasting**: Events only sent to clients in the same auction room
- **Non-Blocking**: Slow clients don't block fast ones (buffered channels)
- **Graceful Cleanup**: Disconnected clients automatically removed
- **Empty Room Cleanup**: Rooms with no clients are deleted

### Message Flow

```
Bid Placed → Service Layer → publishAuctionEvents() 
→ Hub.BroadcastToAuction() → Client Send Channels 
→ Client Write Pump → WebSocket Connection → Client Browser
```

## Performance Considerations

### Scalability

- Each client has a buffered send channel (256 messages)
- Write operations have 10-second timeout
- Read operations detect disconnections immediately
- Empty rooms are automatically cleaned up

### Resource Management

- Goroutines per client: 1 (write pump)
- Memory per client: ~1KB (excluding buffers)
- Message size limit: 512KB

### Connection Limits

Current implementation supports:
- Unlimited auction rooms
- Unlimited clients per room
- Message compression enabled

For production, consider:
- Rate limiting on connections
- Authentication on WebSocket upgrade
- Monitoring active connections
- Load balancing across instances

## Troubleshooting

### Connection Refused

```
WebSocket connection to 'ws://localhost:8080/ws/auction' failed
```

**Solutions:**
1. Ensure the auction service is running
2. Check the port (default: 8080)
3. Verify the URL includes auction_id parameter

### No Events Received

**Solutions:**
1. Confirm you're connected to the correct auction_id
2. Check that the auction exists and is active
3. Place a test bid to trigger an event

### Connection Drops

**Solutions:**
1. Check network stability
2. Verify ping/pong messages are being exchanged
3. Check server logs for errors

## Security Considerations

### Current Implementation

- Basic origin checking (disabled for development)
- No authentication on WebSocket upgrade
- No message validation from clients

### Production Recommendations

1. **Authentication**: Verify JWT token on WebSocket upgrade
2. **Authorization**: Check user permissions for auction access
3. **Origin Validation**: Restrict allowed origins
4. **Rate Limiting**: Prevent connection abuse
5. **Message Validation**: Validate client messages
6. **TLS/WSS**: Use secure WebSocket (wss://) in production

### Example with Authentication

```go
func ServeWs(hub *Hub, jwtManager *auth.JWTManager, w http.ResponseWriter, r *http.Request) {
    // Verify JWT from query param or header
    token := r.URL.Query().Get("token")
    claims, err := jwtManager.VerifyToken(token)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    auctionID := r.URL.Query().Get("auction_id")
    
    // Verify user has access to this auction
    // ... authorization logic ...

    // Proceed with WebSocket upgrade
    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        OriginPatterns: []string{"yourdomain.com"},
    })
    // ...
}
```

## Future Enhancements

- [ ] Authentication on WebSocket connections
- [ ] User-specific notifications
- [ ] Auction room statistics (viewer count)
- [ ] Message history/replay for late joiners
- [ ] Presence indicators (who's watching)
- [ ] Bidder-specific messages (outbid notifications)
- [ ] Admin broadcast messages
- [ ] Connection analytics and monitoring
- [ ] Redis pub/sub for multi-instance support

## API Documentation

See the [main README](README.md) for complete API documentation of the auction service endpoints.
