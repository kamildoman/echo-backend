package storage

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kamildoman/echo-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     string
	Password string
	User     string
	DBName   string
	SSLMode  string
}

var DB *gorm.DB

func NewConnection() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	
	config := &Config{
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User: os.Getenv("DB_USER"),
		SSLMode: os.Getenv("DB_SSL"),
		DBName: os.Getenv("DB_NAME"),
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("could not connect to the database")
	}
	DB = db

	models.MigrateUsers(db)
}