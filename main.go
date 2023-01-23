package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/kamildoman/echo-backend/functions"
	"github.com/kamildoman/echo-backend/storage"
)


func SetupRoutes (app *fiber.App) {
	api := app.Group("/api")
	api.Post("create_user", functions.CreateUser)
	api.Post("login", functions.Login)
	api.Post("logout", functions.Logout)
	api.Post("create_post", functions.CreatePost)
	api.Post("create_comment", functions.CreateComment)
	api.Get("posts", functions.GetPosts)
	api.Get("user", functions.AuthenticateUser)
}

func main() {
	storage.NewConnection()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	SetupRoutes(app)

	app.Listen(":8080")
}