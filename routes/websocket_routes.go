package routes

import (
	"zygo-backend/handlers"
	ws "zygo-backend/websocket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// WebSocketRoutes sets up all WebSocket related routes
func WebSocketRoutes(app *fiber.App, hub *ws.Hub) {
	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		// Check if it's a WebSocket upgrade request
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Driver WebSocket endpoint
	// Connect: ws://localhost:3000/ws/driver?driver_id=xxx&token=xxx
	app.Get("/ws/driver", websocket.New(func(c *websocket.Conn) {
		handleDriverConnection(c, hub)
	}))

	// Rider WebSocket endpoint
	// Connect: ws://localhost:3000/ws/rider?rider_id=xxx&token=xxx
	app.Get("/ws/rider", websocket.New(func(c *websocket.Conn) {
		handleRiderConnection(c, hub)
	}))

	// REST endpoints for driver status
	api := app.Group("/api/tracking")
	api.Get("/drivers/online", handlers.GetOnlineDrivers)
	api.Get("/driver/:driver_id/status", handlers.GetDriverStatus)
}

// handleDriverConnection handles new driver WebSocket connections
func handleDriverConnection(c *websocket.Conn, hub *ws.Hub) {
	// Get driver ID from query params
	driverID := c.Query("driver_id")
	if driverID == "" {
		c.WriteJSON(fiber.Map{"error": "driver_id is required"})
		c.Close()
		return
	}

	// TODO: Validate JWT token from query params or headers
	// token := c.Query("token")
	// claims, err := utils.ValidateJWT(token)
	// if err != nil { ... }

	client := ws.NewClient(hub, c, driverID, "driver")
	hub.Register(client)

	// Send welcome message
	client.SendMessage(ws.WSMessage{
		Type:    "connected",
		Payload: map[string]string{"message": "Driver connected successfully", "driver_id": driverID},
	})

	// Start the write pump in a goroutine
	go client.WritePump()

	// ReadPump blocks until connection is closed
	client.ReadPump()
}

// handleRiderConnection handles new rider WebSocket connections
func handleRiderConnection(c *websocket.Conn, hub *ws.Hub) {
	// Get rider ID from query params
	riderID := c.Query("rider_id")
	if riderID == "" {
		c.WriteJSON(fiber.Map{"error": "rider_id is required"})
		c.Close()
		return
	}

	// TODO: Validate JWT token from query params or headers
	// token := c.Query("token")
	// claims, err := utils.ValidateJWT(token)
	// if err != nil { ... }

	client := ws.NewClient(hub, c, riderID, "rider")
	hub.Register(client)

	// Send welcome message
	client.SendMessage(ws.WSMessage{
		Type:    "connected",
		Payload: map[string]string{"message": "Rider connected successfully", "rider_id": riderID},
	})

	// Start the write pump in a goroutine
	go client.WritePump()

	// ReadPump blocks until connection is closed
	client.ReadPump()
}
