package handlers

import (
	"strconv"
	"zygo-backend/models"

	"github.com/gofiber/fiber/v2"
)

func GetFareEstimate(c *fiber.Ctx)error{
	vehicleType := c.Query("vehicle")
	distStr := c.Query("distance")
	durStr := c.Query("duration")

	distance, _ := strconv.ParseFloat(distStr, 64)
	duration, _ := strconv.ParseFloat(durStr, 64)

	var base, perKm, perMin, minFare float64

	switch vehicleType {
	case "bike":
		base, perKm, perMin, minFare = 60, 50, 2, 80
	case "tuk":
		base, perKm, perMin, minFare = 100, 80, 3, 120
	case "car":
		base, perKm, perMin, minFare = 200, 120, 5, 300
	case "van":
		base, perKm, perMin, minFare = 350, 160, 8, 500
	default:
		return c.Status(400).JSON(fiber.Map{"error":"Invalid vechicle type"})
	}

	distCost := distance * perKm
	durCost := duration * perMin
	total := base + distCost + durCost

	if total < minFare {
		total = minFare
	}

	return c.JSON(models.PriceResponse{
		VehicleType: vehicleType,
		BaseFare: base,
		DistanceCost: distCost,
		DurationCost: durCost,
		TotalFare: total,
		Currency: "LKR",
	})
}