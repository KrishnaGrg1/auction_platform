package socket

import (
	"encoding/json"
	"sync"
	"time"
)

type Client struct {
	send chan []byte
}

type Hub struct {
	rooms map[string]map[*Client]bool
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]bool),
	}
}

type AuctionEvent struct {
	Type      string    `json:"type"`
	AuctionID string    `json:"auction_id"`
	UserID    string    `json:"user_id,omitempty"`
	Amount    int64     `json:"amount,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Hub) BroadcastToAuction(
	auctionID string,
	payload any,
) {

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := h.rooms[auctionID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(clients, client)
		}
	}
}
