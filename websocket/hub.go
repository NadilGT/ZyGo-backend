package websocket

import (
	"log"
	"sync"
	"time"
)

// LocationUpdate represents a driver's real-time location
type LocationUpdate struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
	Timestamp int64   `json:"timestamp"`
}

// WSMessage is a generic WebSocket message wrapper
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SubscribeMessage is sent by riders to subscribe to a driver's location
type SubscribeMessage struct {
	Action   string `json:"action"` // "subscribe" or "unsubscribe"
	DriverID string `json:"driver_id"`
}

// DriverStatus represents the current status of a driver
type DriverStatus struct {
	DriverID string          `json:"driver_id"`
	IsOnline bool            `json:"is_online"`
	LastSeen time.Time       `json:"last_seen"`
	Location *LocationUpdate `json:"location,omitempty"`
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Mutex for thread-safe operations
	mu sync.RWMutex

	// All connected clients
	clients map[*Client]bool

	// Drivers indexed by their ID
	drivers map[string]*Client

	// Riders subscribed to each driver: driverID -> set of clients
	subscriptions map[string]map[*Client]bool

	// Latest location for each driver (for new subscribers)
	driverLocations map[string]*LocationUpdate

	// Channel for registering clients
	register chan *Client

	// Channel for unregistering clients
	unregister chan *Client

	// Channel for broadcasting location updates
	broadcast chan *LocationUpdate
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		drivers:         make(map[string]*Client),
		subscriptions:   make(map[string]map[*Client]bool),
		driverLocations: make(map[string]*LocationUpdate),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan *LocationUpdate),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	log.Println("WebSocket Hub started")
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case location := <-h.broadcast:
			h.broadcastLocation(location)
		}
	}
}

// Register adds a client to the hub (thread-safe public method)
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	if client.UserType == "driver" {
		h.drivers[client.UserID] = client
		log.Printf("Driver %s connected", client.UserID)
	} else {
		log.Printf("Rider %s connected", client.UserID)
	}
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		if client.UserType == "driver" {
			delete(h.drivers, client.UserID)
			// Notify subscribers that driver went offline
			h.notifyDriverOffline(client.UserID)
			log.Printf("Driver %s disconnected", client.UserID)
		} else {
			// Remove rider from all subscriptions
			for driverID, subscribers := range h.subscriptions {
				delete(subscribers, client)
				if len(subscribers) == 0 {
					delete(h.subscriptions, driverID)
				}
			}
			log.Printf("Rider %s disconnected", client.UserID)
		}
	}
}

// broadcastLocation sends location update to all subscribers
func (h *Hub) broadcastLocation(location *LocationUpdate) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Store latest location
	h.driverLocations[location.DriverID] = location

	// Get subscribers for this driver
	subscribers, exists := h.subscriptions[location.DriverID]
	if !exists || len(subscribers) == 0 {
		return
	}

	// Create message
	msg := WSMessage{
		Type:    "location_update",
		Payload: location,
	}

	// Broadcast to all subscribers
	for client := range subscribers {
		select {
		case client.send <- msg:
		default:
			// Client's buffer is full, skip
			log.Printf("Client %s buffer full, skipping message", client.UserID)
		}
	}
}

// Subscribe adds a rider to a driver's subscriber list
func (h *Hub) Subscribe(client *Client, driverID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscriptions[driverID] == nil {
		h.subscriptions[driverID] = make(map[*Client]bool)
	}
	h.subscriptions[driverID][client] = true

	log.Printf("Rider %s subscribed to driver %s", client.UserID, driverID)

	// Send current location if available
	if loc, exists := h.driverLocations[driverID]; exists {
		msg := WSMessage{
			Type:    "location_update",
			Payload: loc,
		}
		select {
		case client.send <- msg:
		default:
		}
	}

	// Send driver online status
	_, isOnline := h.drivers[driverID]
	statusMsg := WSMessage{
		Type: "driver_status",
		Payload: DriverStatus{
			DriverID: driverID,
			IsOnline: isOnline,
			LastSeen: time.Now(),
		},
	}
	select {
	case client.send <- statusMsg:
	default:
	}
}

// Unsubscribe removes a rider from a driver's subscriber list
func (h *Hub) Unsubscribe(client *Client, driverID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, exists := h.subscriptions[driverID]; exists {
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.subscriptions, driverID)
		}
	}

	log.Printf("Rider %s unsubscribed from driver %s", client.UserID, driverID)
}

// notifyDriverOffline notifies all subscribers that a driver went offline
func (h *Hub) notifyDriverOffline(driverID string) {
	subscribers, exists := h.subscriptions[driverID]
	if !exists {
		return
	}

	msg := WSMessage{
		Type: "driver_status",
		Payload: DriverStatus{
			DriverID: driverID,
			IsOnline: false,
			LastSeen: time.Now(),
		},
	}

	for client := range subscribers {
		select {
		case client.send <- msg:
		default:
		}
	}
}

// GetOnlineDrivers returns a list of online driver IDs
func (h *Hub) GetOnlineDrivers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	drivers := make([]string, 0, len(h.drivers))
	for driverID := range h.drivers {
		drivers = append(drivers, driverID)
	}
	return drivers
}

// IsDriverOnline checks if a driver is currently online
func (h *Hub) IsDriverOnline(driverID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.drivers[driverID]
	return exists
}

// GetDriverLocation returns the latest location of a driver
func (h *Hub) GetDriverLocation(driverID string) *LocationUpdate {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.driverLocations[driverID]
}
