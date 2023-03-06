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
	api.Post("send_message", functions.SendMessage)
	api.Post("read_message", functions.ReadMessage)
	api.Post("update_avatar", functions.UpdateAvatar)
	api.Post("like_post", functions.ToggleLike)
	api.Post("send_email", functions.SendEmail)
	api.Post("create_mission", functions.CreateMission)
	api.Post("create_mission_progress", functions.CreateMissionProgress)
	api.Post("create_metric_definition", functions.CreateMetricDefinition)
	api.Post("mission_complete", functions.MissionComplete)
	api.Get("users_metrics", functions.GetMetricsForUser)
	api.Get("posts", functions.GetPosts)
	api.Get("user", functions.AuthenticateUser)
	api.Get("health", functions.HealthCheck)
	api.Get("user_id", functions.GetUserByID)
	api.Get("all_users", functions.GetAllUsers)
	api.Get("user_messages", functions.GetUsersMessages)
	api.Get("all_missions", functions.GetAllMissions)
	api.Get("all_missions_progress", functions.GetAllMissionProgress)
	api.Get("users_mission_progress", functions.GetUsersMissionProgress)
	api.Get("all_metric_definitions", functions.GetAllMetricDefinitions)
	api.Get("all_periods", functions.GetAllPeriods)
	api.Delete("delete_user", functions.DeleteUserByID)
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