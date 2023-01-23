package models

import (
	"gorm.io/gorm"
)

type Users struct {
	ID       string `gorm:"unique" json:"id"`
	Email    *string `gorm:"unique" json:"email"`
	Username *string `gorm:"unique" json:"username"`
	Avatar 	 string `json:"avatar"`
	Level    *int    `json:"level"`
	Password *string `json:"-"`
}

type Posts struct {
	ID string `gorm:"unique" json:"id"`
    Message *string `json:"message"`
	CreatedAt *int `json:"created_at"`
    UserID *string `json:"userId"`
}

type Comments struct {
    ID string `gorm:"unique" json:"id"`
    Message *string `json:"message"`
	CreatedAt *int `json:"created_at"`
    UserID *string `json:"userId"`
    PostID *string `json:"postId"`
}

func MigrateUsers(db *gorm.DB) error{
	// db.Migrator().CreateTable(Comments{})
	err := db.AutoMigrate(&Users{}, &Posts{}, &Comments{})
	return err
}