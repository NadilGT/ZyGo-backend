package handlers

import (
	"context"
	"zygo-backend/config"
	"zygo-backend/models"
	"zygo-backend/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func Login(c *fiber.Ctx)error{
	var body struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	c.BodyParser(&body)

	var user models.User

	err := config.DATABASE.Collection("Users").FindOne(context.TODO(), bson.M{"email": body.Email}).Decode(&user)
	if err != nil || !utils.CheckPassword(body.Password, user.Password) {
		return c.Status(401).JSON(fiber.Map{"error":"Invalid credentials"})
	}

	token, _ := utils.GenerateToken(user.ID.Hex(), user.Name);

	return c.JSON(fiber.Map{
		"token":token,
		"user": fiber.Map{
			"id": user.ID,
			"name": user.Name,
			"email": user.Email,
		},
	})
}