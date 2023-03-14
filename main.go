package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/kamildoman/echo-backend/functions"
	"github.com/kamildoman/echo-backend/storage"
)

type client struct{
	id string
} 

var clients = make(map[*websocket.Conn]client) 
var register = make(chan *websocket.Conn)
var broadcast = functions.GetBroadcast()
var unregister = make(chan *websocket.Conn)

func Contains[T comparable](s []T, e T) bool {
    for _, v := range s {
        if v == e {
            return true
        }
    }
    return false
}


func runHub() {
    for {
        select {
        case connection := <-register:
            id := connection.Params("id")
			clients[connection] = client{id: id}

			
        case message := <-broadcast:
            // Notify the clients that a post has been liked
            postBytes, err := json.Marshal(message)
            if err != nil {
                log.Println("error encoding Post struct to JSON:", err)
                continue
            }
			
			messageType, ok := message["type"].(string)

			if !ok {
				log.Println("invalid message type")
				continue
			}	

			var ids []string;

			if val, ok := message["ids"]; ok {
				ids = val.([]string)
			  } else {
				ids = []string{}
			  }
	
            for connection, client := range clients {
				if messageType != "NEW_MESSAGE" || Contains(ids, client.id) {
					err := connection.WriteMessage(websocket.TextMessage, []byte(postBytes)); 
					if err != nil {
						// Remove the connection from the map
						unregister <- connection

						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
					}
				}
                
            }

        case connection := <-unregister:
            // Unregister the client
            delete(clients, connection)
        }
    }
}


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
	api.Delete("delete_post", functions.DeletePostByID)

	api.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// When the function returns, unregister the client and close the connection
		defer func() {
			unregister <- c
			c.Close()
		}()

		// Extract the id from the URL
		id := c.Params("id")

		// Register the client with the id
		var exists bool
			for _, c := range clients {
				if c.id == id {
					exists = true
					break
				}
			}

		if (!exists) {
			register <- c
			clients[c] = client{id: id}
		}
		
		
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}
				return // Calls the deferred function, i.e. closes the connection on error
			}

			var post map[string]interface{}
			err = json.Unmarshal(message, &post)
			if err != nil {
				log.Println("unmarshal error:", err)
				continue
			}
			
			broadcast <- post
		}
	}))
}

func main() {
	storage.NewConnection()

	app := fiber.New()

	// app.Use(func(c *fiber.Ctx) {
	// 	if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
	// 		c.Next()
	// 	}
	// })

	go runHub()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	SetupRoutes(app)

	app.Listen(":8080")
}