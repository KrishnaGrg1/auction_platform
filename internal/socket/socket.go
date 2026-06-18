package socket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type AuctionEvent struct {
	Type      string    `json:"type"`
	AuctionID string    `json:"auction_id"`
	UserID    string    `json:"user_id,omitempty"`
	Amount    int64     `json:"amount,omitempty"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Client struct {
	conn      *websocket.Conn
	auctionID string
	userID    string
	ctx       context.Context
	cancel    context.CancelFunc
	sendCh    chan []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	hub := &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
	}
	go hub.run()
	return hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.auctionID] == nil {
				h.rooms[client.auctionID] = make(map[*Client]bool)
			}
			h.rooms[client.auctionID][client] = true
			h.mu.Unlock()
			log.Printf("Client joined auction room: %s (total: %d)", client.auctionID, len(h.rooms[client.auctionID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.auctionID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.sendCh)
					if len(clients) == 0 {
						delete(h.rooms, client.auctionID)
					}
					log.Printf("Client left auction room: %s (remaining: %d)", client.auctionID, len(clients))
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) JoinRoom(auctionID string, client *Client) {
	h.register <- client
}

func (h *Hub) LeaveRoom(auctionID string, client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToAuction(auctionID string, payload any) {
	h.mu.RLock()
	clients, ok := h.rooms[auctionID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		h.mu.RUnlock()
		return
	}

	// Send to all clients in the room
	for client := range clients {
		select {
		case client.sendCh <- data:
		default:
			// Client's send buffer is full, skip this message
			log.Printf("Client send buffer full for auction %s", auctionID)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) GetRoomCount(auctionID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.rooms[auctionID]; ok {
		return len(clients)
	}
	return 0
}

// ServeWs — HTTP handler for WebSocket connections
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	auctionID := r.URL.Query().Get("auction_id")
	if auctionID == "" {
		http.Error(w, `{"error":"auction_id query parameter required"}`, http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		CompressionMode:    websocket.CompressionContextTakeover,
	})
	if err != nil {
		log.Printf("Failed to accept websocket: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	client := &Client{
		conn:      conn,
		auctionID: auctionID,
		userID:    userID,
		ctx:       ctx,
		cancel:    cancel,
		sendCh:    make(chan []byte, 256),
	}

	hub.JoinRoom(auctionID, client)
	defer func() {
		hub.LeaveRoom(auctionID, client)
		conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	// Send welcome message
	welcomeMsg := AuctionEvent{
		Type:      "connected",
		AuctionID: auctionID,
		Message:   "Successfully connected to auction room",
		Timestamp: time.Now(),
	}
	if data, err := json.Marshal(welcomeMsg); err == nil {
		client.sendCh <- data
	}

	// Start writer goroutine
	go client.writePump()

	// Reader pump (main goroutine)
	client.readPump()
}

func (c *Client) readPump() {
	defer c.cancel()

	// Set read deadline for ping/pong
	c.conn.SetReadLimit(512 * 1024) // 512KB max message size

	for {
		_, msg, err := c.conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}

		// Handle ping/pong or other client messages
		var clientMsg map[string]interface{}
		if err := json.Unmarshal(msg, &clientMsg); err == nil {
			if msgType, ok := clientMsg["type"].(string); ok && msgType == "ping" {
				pong := AuctionEvent{
					Type:      "pong",
					AuctionID: c.auctionID,
					Timestamp: time.Now(),
				}
				if data, err := json.Marshal(pong); err == nil {
					select {
					case c.sendCh <- data:
					default:
					}
				}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.sendCh:
			if !ok {
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}

			writeCtx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, message)
			cancel()

			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send periodic ping
			ping := AuctionEvent{
				Type:      "ping",
				AuctionID: c.auctionID,
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(ping); err == nil {
				select {
				case c.sendCh <- data:
				default:
				}
			}

		case <-c.ctx.Done():
			return
		}
	}
}
