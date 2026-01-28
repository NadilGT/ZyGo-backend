package handlers

import (
	"context"
	"zygo-backend/config"
	"zygo-backend/models"
	"zygo-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func Register(c *fiber.Ctx) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid inputs"})
	}

	hash, _ := utils.HashPassword(body.Password)

	user := models.User {
		Name: body.Name,
		Email: body.Email,
		Password: hash,
	}

	_, err := config.DATABASE.Collection("Users").InsertOne(context.TODO(), user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error":"User exists"})
	}

	return c.JSON(fiber.Map{"message":"User registerd"})
}
