package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Users struct {
	ID       string `gorm:"unique" json:"id"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Avatar 	 string `json:"avatar"`
	Level    *int    `json:"level"`
	Password *string `json:"-"`
}

type Posts struct {
	ID string `gorm:"unique" json:"id"`
    Message *string `json:"message"`
	CreatedAt *int `json:"created_at"`
    UserID *string `json:"userId"`
	Likes pq.StringArray `gorm:"type:text[]" json:"likes"`
}

type Comments struct {
    ID string `gorm:"unique" json:"id"`
    Message *string `json:"message"`
	CreatedAt *int `json:"created_at"`
    UserID *string `json:"userId"`
    PostID *string `json:"postId"`
}

type Messages struct {
	MessageID      string `json:"message_id"`
	Title		  *string `json:"title"`
	Message       *string `json:"message"`
	SendUserID    *string `json:"send_user_id"`
	RecieveUserID *string `json:"recieve_user_id"`
	CreatedAt     *int    `json:"created_at"`
	Read          *bool   `json:"read"`
}

func MigrateUsers(db *gorm.DB) error{
	// db.Migrator().CreateTable(Messages{})
	err := db.AutoMigrate(&Users{}, &Posts{}, &Comments{}, &Messages{})
	return err
}