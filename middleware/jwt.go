package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var SecretKey = []byte(os.Getenv("JWT_SECRET")) // must match your utils/jwt.go

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// -------------------------
		// Get Authorization header
		// -------------------------
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing token",
			})
		}

		// -------------------------
		// Extract token
		// "Bearer TOKEN"
		// -------------------------
		tokenString := strings.Split(authHeader, " ")

		if len(tokenString) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		token := tokenString[1]

		// -------------------------
		// Parse token
		// -------------------------
		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return SecretKey, nil
		})

		if err != nil || !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// -------------------------
		// Get claims
		// -------------------------
		claims := parsedToken.Claims.(jwt.MapClaims)

		// store user_id inside context
		c.Locals("user_id", claims["user_id"])

		// continue
		return c.Next()
	}
}
