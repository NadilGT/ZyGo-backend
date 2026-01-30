package routes

import (
	"zygo-backend/handlers"
	"zygo-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	api := app.Group("/auth")
	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
	api.Get("/profile", middleware.JWTProtected(),handlers.Profile)
	api.Post("/refresh", handlers.Refresh)
}