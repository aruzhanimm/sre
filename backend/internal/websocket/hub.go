package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Message struct {
	Type    string      `json:"type"`
	EventID string      `json:"eventId"`
	Data    interface{} `json:"data"`
}

type OddsUpdateMessage struct {
	MarketID  string  `json:"marketId"`
	Outcome   string  `json:"outcome"`
	OldOdds   float64 `json:"oldOdds"`
	NewOdds   float64 `json:"newOdds"`
	Change    float64 `json:"change"`
	Direction string  `json:"direction"`
	Timestamp int64   `json:"timestamp"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				// Remove from all rooms
				for room, clients := range h.rooms {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, room)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}

func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.Rooms[room] = true
}

func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(client.Rooms, room)
}

func (h *Hub) broadcastToRoom(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	room := msg.EventID
	clients, ok := h.rooms[room]
	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(clients, client)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) BroadcastOddsUpdate(eventID string, msg *OddsUpdateMessage) {
	h.broadcast <- &Message{
		Type:    "odds_update",
		EventID: eventID,
		Data:    msg,
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) GetConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
