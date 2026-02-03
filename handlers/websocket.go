package handlers

import (
	"time"
	ws "zygo-backend/websocket"

	"github.com/gofiber/fiber/v2"
)

// WebSocketHub is the global hub instance
var WebSocketHub *ws.Hub

// InitWebSocketHub initializes the WebSocket hub
func InitWebSocketHub() *ws.Hub {
	WebSocketHub = ws.NewHub()
	go WebSocketHub.Run()
	return WebSocketHub
}

// DriverStatusResponse represents the status response for a driver
type DriverStatusResponse struct {
	DriverID string             `json:"driver_id"`
	IsOnline bool               `json:"is_online"`
	LastSeen time.Time          `json:"last_seen,omitempty"`
	Location *ws.LocationUpdate `json:"location,omitempty"`
}

// GetOnlineDrivers returns the list of online drivers
func GetOnlineDrivers(c *fiber.Ctx) error {
	if WebSocketHub == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "WebSocket service not initialized",
		})
	}

	drivers := WebSocketHub.GetOnlineDrivers()
	return c.JSON(fiber.Map{
		"online_drivers": drivers,
		"count":          len(drivers),
	})
}

// GetDriverStatus returns the status of a specific driver
func GetDriverStatus(c *fiber.Ctx) error {
	if WebSocketHub == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "WebSocket service not initialized",
		})
	}

	driverID := c.Params("driver_id")
	if driverID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "driver_id is required",
		})
	}

	isOnline := WebSocketHub.IsDriverOnline(driverID)
	location := WebSocketHub.GetDriverLocation(driverID)

	status := DriverStatusResponse{
		DriverID: driverID,
		IsOnline: isOnline,
		Location: location,
	}

	return c.JSON(status)
}
