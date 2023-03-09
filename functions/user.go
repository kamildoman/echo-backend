package functions

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/models"
	"github.com/kamildoman/echo-backend/storage"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
    ID string `json:"id"`
    Email string `json:"email"`
    Username string `json:"username"`
    Password string `json:"-"`
	Avatar string `json:"avatar,omitempty"`
    Level int `json:"level"`
	Role int `json:"role"`
    // Posts []Post `gorm:"foreignkey:UserID"`
    // Comments []Comment `gorm:"foreignkey:UserID"`
}

type GamLevels struct {
	LevelId string `json:"level_id"`
	Level int `json:"level"`
}

type GamLevelProgress struct {
	UserId string
	LevelId string
	ExpPoints int
	MissingExp int
	Coins int
}

type Invite struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type Employees struct {
	EmployeeId string
	UserId string
}

func HealthCheck(c *fiber.Ctx) error {
	c.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "OK"})
	return nil
}

func GetUserIdByEmployeeId(employeeId string) (string, error) {
	var employee Employees
	if err := storage.DB.Table("employees").Select("user_id").Where("employee_id = ?", employeeId).First(&employee).Error; err != nil {
		return "", err
	}
	return employee.UserId, nil
}

func GetEmployeeIdByUserId(userId string) (string, error) {
	var employee Employees
	if err := storage.DB.Table("employees").Select("employee_id").Where("user_id = ?", userId ).First(&employee).Error; err != nil {
		return "", err
	}
	return employee.EmployeeId, nil
}

func GetUserByID(context *fiber.Ctx) error {
	id := context.Query("id")
	var user User
	err := storage.DB.Select(`
			users.id,
			users.username,
			users.email,
			users.avatar,
			gam_level_progress.exp_points,
			gam_level_progress.missing_exp,
			gam_level_progress.coins,
			gam_levels.level
		`).
		Joins(`
			LEFT JOIN gam_level_progress ON users.id = gam_level_progress.user_id
			LEFT JOIN gam_levels ON gam_level_progress.level_id = gam_levels.level_id
		`).
		Where("users.id = ?", id).
		First(&user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get user"})
		return err
	}
	context.Status(http.StatusOK).JSON(user)
	return nil
}


func DeleteUserByID(context *fiber.Ctx) error {
	id := context.Query("id")
	err := storage.DB.Model(&User{}).Where("id = ?", id).Update("username", "Użytkownik usunięty").Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not delete user"})
		return err
	}
	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "user deleted successfully"})
	return nil
}

func GetAllUsers(context *fiber.Ctx) error {
	var users []User
	err := storage.DB.Select("id, level, username, email, avatar, role").Find(&users).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get users"})
		return err
	}
	context.Status(http.StatusOK).JSON(users)
	return nil
}

type AvatarRequestData struct {
    Blob string `json:"blob"`
    Name string `json:"name"`
}

func UpdateAvatar (c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	SecretKey := os.Getenv("SECRET_KEY")
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error){
		return []byte(SecretKey), nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	userId := claims.Issuer

	blobData := AvatarRequestData{}
	
	c.BodyParser(&blobData)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	conf := aws.Config{Region: aws.String("eu-central-1")}
    sess := session.New(&conf)

    svc := s3manager.NewUploader(sess)

	decodedImage, err := base64.StdEncoding.DecodeString(blobData.Blob)

	if err != nil {
		c.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Error decoding the image"})
		return err
	}

	// Resize the image
	img, err := imaging.Decode(bytes.NewReader(decodedImage))
	if err != nil {
		c.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to decode the image"})
		return err
	}
	img = imaging.Resize(img, 100, 100, imaging.Lanczos)

	// Encode the resized image as a JPEG
	var resizedImage bytes.Buffer
	err = imaging.Encode(&resizedImage, img, imaging.JPEG)
	if err != nil {
		c.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to encode the image"})
		return err
	}

	fmt.Println("Uploading file to S3...")
	reader := bytes.NewReader(resizedImage.Bytes())
	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String("echoavatars"),
		Key:    aws.String(blobData.Name),
		Body:   reader,
	})

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		log.Println("Failed to upload file to S3:", err)
		return nil
	}

	err = storage.DB.Model(&User{}).Where("id = ?", userId).Update("avatar", result.Location).Error

	if err != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not update the avatar"})
		return err
	}

	c.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "avatar updated!"})
	return nil
}

func SendRegistrationInvite(context *fiber.Ctx) error {
    // Parse email address from request body
    email := struct {
        Email string `json:"email"`
    }{}
    err := context.BodyParser(&email)
    if err != nil {
        context.Status(http.StatusUnprocessableEntity).JSON(
            &fiber.Map{"message": "request failed"})
        return err
    }
    
    // Generate unique token for registration link
    token := uuid.NewV4().String()
    
    // Save token to database
    invite := Invite{
        Token:  token,
        Email:  email.Email,
    }

    err = storage.DB.Create(&invite).Error
    if err != nil {
        context.Status(http.StatusBadRequest).JSON(
            &fiber.Map{"message": "could not create invite"})
        return err
    }
    
    // Send registration invite email
    // err = emailService.SendRegistrationInvite(email.Email, token)
    // if err != nil {
    //     context.Status(http.StatusInternalServerError).JSON(
    //         &fiber.Map{"message": "could not send registration invite"})
    //     return err
    // }
    
    context.Status(http.StatusOK).JSON(
        &fiber.Map{"message": "registration invite sent"})
    return nil
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
		ExpiresAt: jwt.NewTime(float64(time.Now().Add(time.Hour * 72).Unix())),
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

	var user User

	err = storage.DB.Select(`
			users.id,
			users.username,
			users.email,
			users.avatar,
			users.role,
			gam_level_progress.exp_points,
			gam_level_progress.missing_exp,
			gam_level_progress.coins,
			gam_levels.level
		`).
		Joins(`
			LEFT JOIN gam_level_progress ON users.id = gam_level_progress.user_id
			LEFT JOIN gam_levels ON gam_level_progress.level_id = gam_levels.level_id
		`).
		Where("users.id = ?", claims.Issuer).
		First(&user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get user"})
		return err
	}

	var count int64
	storage.DB.Model(&Messages{}).Where("receive_user_id = ? and read = false", user.ID).Count(&count)

	
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