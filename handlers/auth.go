package handlers

import (
	"errors"
	"strings"
	"time"

	"zygo-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Refresh(c *fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing token",
		})
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token format",
		})
	}

	tokenStr := tokenParts[1]

	// Parse token without verifying expiry
	parsedToken, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return utils.SecretKey, nil
	}, jwt.WithLeeway(time.Minute*5)) // optional leeway

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user_id in token",
		})
	}

	name, _ := claims["name"].(string) // name is optional, default to empty string

	// Generate new token
	newToken, err := utils.GenerateToken(userID, name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token": newToken,
	})
}
