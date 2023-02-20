package functions

import (
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/storage"
	uuid "github.com/satori/go.uuid"
)

type Messages struct {
	MessageID     string `json:"message_id"`
	Title		  string `json:"title"`
	Message       string `json:"message"`
	SendUserID    string `json:"send_user_id"`
	ReceiveUserID string `json:"receive_user_id"`
	CreatedAt     int    `json:"created_at"`
	Read          bool   `json:"read"`
	User User `gorm:"foreignkey:SendUserID"`
	ReceiveUser User `gorm:"foreignkey:ReceiveUserID"`
}

type MessageMultiple struct {
    Messages
    ReceiveUserIDs []string `json:"receive_user_ids"`
}

func SendMessage(context *fiber.Ctx) error {
    messageMultiple := MessageMultiple{}
    err := context.BodyParser(&messageMultiple)
    if err != nil {
        context.Status(http.StatusUnprocessableEntity).JSON(
            &fiber.Map{"message": "request failed"})
        return err
    }

	addToCreatedAt := 0
    for _, receiveUserID := range messageMultiple.ReceiveUserIDs {
		id := uuid.NewV4().String()
        message := Messages{
            MessageID:     id,
            Title:         messageMultiple.Title,
            Message:       messageMultiple.Message,
            SendUserID:    messageMultiple.SendUserID,
            ReceiveUserID: receiveUserID,
            CreatedAt:     messageMultiple.CreatedAt + addToCreatedAt,
            Read:          false,
        }
		addToCreatedAt += 1

        //save message to database
        err = storage.DB.Create(&message).Error
        if err != nil {
            context.Status(http.StatusBadRequest).JSON(
                &fiber.Map{"message": "could not send the message"})
            return err
        }
    }

    context.Status(http.StatusOK).JSON(
        &fiber.Map{"message": "message sent!"})
    return nil
}

func GetUsersMessages (context *fiber.Ctx) error {
	cookie := context.Cookies("jwt")

	SecretKey := os.Getenv("SECRET_KEY")
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error){
		return []byte(SecretKey), nil
	})

	if err != nil {
		return context.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	userId := claims.Issuer

	var sentMessages, receivedMessages []*Messages
	storage.DB.Order("created_at desc").Where("send_user_id = ?", userId).Preload("ReceiveUser").Find(&sentMessages)
	storage.DB.Order("created_at desc").Where("receive_user_id = ?", userId).Preload("User").Find(&receivedMessages)

	res := make(map[string]interface{})
	res["sent"] = formatMessages(sentMessages, false)
	res["received"] = formatMessages(receivedMessages, true)

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"data": res})

	return nil
}

func formatMessages(messages []*Messages, receive bool) []map[string]interface{} {
	res := make([]map[string]interface{}, len(messages))
	for i, message := range messages {
		messageMap := make(map[string]interface{})
		messageMap["message"] = message.Message
		messageMap["title"] = message.Title
		messageMap["message_id"] = message.MessageID
		messageMap["sent_user_id"] = message.SendUserID
		if receive {
            messageMap["username"] = message.User.Username
			messageMap["avatar"] = message.User.Avatar
        } else {
            messageMap["username"] = message.ReceiveUser.Username
			messageMap["avatar"] = message.ReceiveUser.Avatar
        }
		messageMap["read"] = message.Read
		
		res[i] = messageMap
	}
	return res
}

func ReadMessage (context *fiber.Ctx) error {
	cookie := context.Cookies("jwt")

	SecretKey := os.Getenv("SECRET_KEY")
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error){
		return []byte(SecretKey), nil
	})

	if err != nil {
		return context.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	userId := claims.Issuer

	payload := struct {
        MessageID  string `json:"message_id"`
    }{}

    if err := context.BodyParser(&payload); err != nil {
        return err
    }
	

	var message Messages
	storage.DB.Where("receive_user_id = ? and message_id = ?", userId, payload.MessageID).First(&message)

	message.Read = true
	err = storage.DB.Where("receive_user_id = ? and message_id = ?", userId, payload.MessageID).Save(&message).Error
	if err != nil {
		return err
	}

	return nil
}