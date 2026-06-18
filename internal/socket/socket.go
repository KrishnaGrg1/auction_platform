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

// AuctionValidator — socket only needs this one method
// store implements this without socket importing store
type AuctionValidator interface {
	AuctionExists(ctx context.Context, auctionID string) bool
}

type AuctionEvent struct {
	Type      string    `json:"type"`
	AuctionID string    `json:"auction_id"`
	UserID    string    `json:"user_id,omitempty"`
	Amount    int64     `json:"amount,omitempty"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Client struct {
	conn      interface{} // use nhooyr.io/websocket.Conn in ws.go
	auctionID string
	userID    string
	ctx       context.Context
	cancel    context.CancelFunc
	sendCh    chan []byte
	done      chan struct{}
}

type Hub struct {
	validator  AuctionValidator
	rooms      map[string]map[*Client]bool
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

func NewHub(validator AuctionValidator) *Hub {
	hub := &Hub{
		validator:  validator,
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
			log.Printf("Client joined room %s (total: %d)",
				client.auctionID, len(h.rooms[client.auctionID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.auctionID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.sendCh)
					if len(clients) == 0 {
						delete(h.rooms, client.auctionID)
					}
					log.Printf("Client left room %s (remaining: %d)",
						client.auctionID, len(clients))
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
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[auctionID]
	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.sendCh <- data:
		default:
			log.Printf("Client buffer full — disconnecting")
			client.cancel()
		}
	}
}

func (h *Hub) GetRoomCount(auctionID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.rooms[auctionID]; ok {
		return len(clients)
	}
	return 0
}
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	auctionID := r.URL.Query().Get("auction_id")
	if auctionID == "" {
		http.Error(w, `{"error":"auction_id required"}`, http.StatusBadRequest)
		return
	}

	// 🔧 use interface — no store import needed
	if !hub.validator.AuctionExists(r.Context(), auctionID) {
		http.Error(w, `{"error":"auction not found"}`, http.StatusNotFound)
		return
	}

	userID := r.URL.Query().Get("user_id") // TODO: validate from token

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		CompressionMode:    websocket.CompressionContextTakeover,
	})
	if err != nil {
		log.Printf("Failed to accept websocket: %v", err)
		return
	}

	conn.SetReadLimit(512 * 1024)

	ctx, cancel := context.WithCancel(r.Context())

	client := &Client{
		auctionID: auctionID,
		userID:    userID,
		ctx:       ctx,
		cancel:    cancel,
		sendCh:    make(chan []byte, 256),
		done:      make(chan struct{}),
	}

	// store conn separately since Client.conn is interface{}
	// pass conn directly to pumps
	hub.JoinRoom(auctionID, client)

	// welcome message
	welcomeMsg := AuctionEvent{
		Type:      "connected",
		AuctionID: auctionID,
		Message:   "Connected to auction room",
		Timestamp: time.Now(),
	}
	if data, err := json.Marshal(welcomeMsg); err == nil {
		client.sendCh <- data
	}

	go writePump(client, conn)
	readPump(client, conn)

	cancel()
	<-client.done

	hub.LeaveRoom(auctionID, client)
	conn.Close(websocket.StatusNormalClosure, "connection closed")
}

func readPump(c *Client, conn *websocket.Conn) {
	for {
		_, msg, err := conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure &&
				websocket.CloseStatus(err) != websocket.StatusGoingAway {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}

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

func writePump(c *Client, conn *websocket.Conn) {
	defer close(c.done)

	for {
		select {
		case message, ok := <-c.sendCh:
			if !ok {
				conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			writeCtx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			err := conn.Write(writeCtx, websocket.MessageText, message)
			cancel()
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}
