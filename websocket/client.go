package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 4096
)

// Client represents a WebSocket connection
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	UserID   string
	UserType string // "driver" or "rider"
	send     chan WSMessage
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn, userID, userType string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		UserID:   userID,
		UserType: userType,
		send:     make(chan WSMessage, 256),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages based on client type
func (c *Client) handleMessage(message []byte) {
	if c.UserType == "driver" {
		c.handleDriverMessage(message)
	} else {
		c.handleRiderMessage(message)
	}
}

// handleDriverMessage processes messages from drivers (location updates)
func (c *Client) handleDriverMessage(message []byte) {
	var location LocationUpdate
	if err := json.Unmarshal(message, &location); err != nil {
		log.Printf("Error unmarshaling driver location: %v", err)
		return
	}

	// Ensure the driver ID matches the authenticated user
	location.DriverID = c.UserID

	// Set timestamp if not provided
	if location.Timestamp == 0 {
		location.Timestamp = time.Now().UnixMilli()
	}

	// Broadcast to subscribers
	c.Hub.broadcast <- &location
}

// handleRiderMessage processes messages from riders (subscribe/unsubscribe)
func (c *Client) handleRiderMessage(message []byte) {
	var subMsg SubscribeMessage
	if err := json.Unmarshal(message, &subMsg); err != nil {
		log.Printf("Error unmarshaling rider message: %v", err)
		return
	}

	switch subMsg.Action {
	case "subscribe":
		c.Hub.Subscribe(c, subMsg.DriverID)
	case "unsubscribe":
		c.Hub.Unsubscribe(c, subMsg.DriverID)
	default:
		log.Printf("Unknown action: %s", subMsg.Action)
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg WSMessage) {
	select {
	case c.send <- msg:
	default:
		log.Printf("Client %s send buffer full", c.UserID)
	}
}
