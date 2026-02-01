package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"zygo-backend/models"

	"github.com/gofiber/fiber/v2"
)

func GetRoute(c *fiber.Ctx)error{
	startLat := c.Query("startLat")
	startLng := c.Query("startLng")
	endLat := c.Query("endLat")
	endLng := c.Query("endLng")

	if startLat == "" || startLng == "" || endLat == "" || endLng == "" {
		return c.Status(400).JSON(fiber.Map{"error":"All coordinates are required"})
	}

	url := fmt.Sprintf(
		"https://router.project-osrm.org/route/v1/driving/%s,%s;%s,%s?overview=full&geometries=geojson",
		startLng, startLat, endLng, endLat,
	)

	resp, err := http.Get(url)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error":"Failed to fetch route"})
	}
	defer resp.Body.Close()

	var routeResponse models.RouteResponse

	if err := json.NewDecoder(resp.Body).Decode(&routeResponse); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode route response"})
	}
	return c.JSON(routeResponse)
}