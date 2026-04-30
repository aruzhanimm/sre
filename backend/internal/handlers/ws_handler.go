package handlers

import (
	"net/http"

	ws "BetKZ/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev
	},
}

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// GET /ws?eventId=xxx
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	eventID := c.Query("eventId")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	client := ws.NewClient(h.hub, conn)
	h.hub.Register(client)

	if eventID != "" {
		h.hub.JoinRoom(client, eventID)
	}

	go client.WritePump()
	go client.ReadPump()
}
