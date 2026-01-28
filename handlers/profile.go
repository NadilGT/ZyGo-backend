package handlers

import (
	"context"
	"time"
	"zygo-backend/config"
	"zygo-backend/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Profile(c *fiber.Ctx)error{
	userID := c.Locals("user_id")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found in token",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user id type",
		})
	}

	objectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user id",
		})
	}

	collection := config.DATABASE.Collection("Users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User

	err = collection.FindOne(ctx, bson.M{
		"_id": objectID,
	}).Decode(&user)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"user_id": user.ID.Hex(),
		"name": user.Name,
		"email": user.Email,
	})
}