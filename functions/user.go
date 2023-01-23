package functions

import (
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/models"
	"github.com/kamildoman/echo-backend/storage"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
    ID string
    Email string `json:"email"`
    Username string `json:"username"`
    Password string `json:"password"`
	Avatar string `json:"avatar,omitempty"`
    Level int `json:"level"`
    // Posts []Post `gorm:"foreignkey:UserID"`
    // Comments []Comment `gorm:"foreignkey:UserID"`
}

func CreateUser (context *fiber.Ctx) error{
	user := User{}
	err := context.BodyParser(&user)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	//generate unique id
	id := uuid.NewV4().String()
	user.ID = id
	
	//hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
	    context.Status(http.StatusBadRequest).JSON(
		&fiber.Map{"message": "could not hash password"})
	    return err
	}
	user.Password = string(hashedPassword)
	
	//save user to database
	err = storage.DB.Create(&user).Error
	
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create the user"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "user created"})
	return nil
}


func Login (context *fiber.Ctx) error {
	var data map[string]string

	if err := context.BodyParser(&data); err != nil {
		return err
	}

	var user models.Users

	storage.DB.Where("email = ?", data["email"]).Find(&user)

	if user.ID == "" {
		return context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "user doesn't exist"})
	}

	err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(data["password"]))

	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "incorrect password"})
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.ID,
		ExpiresAt: jwt.NewTime(float64(time.Now().Add(time.Hour * 24).Unix())),
	})

	SecretKey := os.Getenv("SECRET_KEY")
	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "could not login",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	context.Cookie(&cookie)

	return context.JSON(fiber.Map{
		"message": "success",
	})
}

func AuthenticateUser (context *fiber.Ctx) error {
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

	var user models.Users

	storage.DB.Where("id = ?", claims.Issuer).First(&user)

	return context.JSON(user)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}