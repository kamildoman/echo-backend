package models

import (
	"time"

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
	Role	 *int    `json:"role"`
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

type Invites struct {
	Token *string `json:"token"`
	Email *string `json:"email"`
}

type GamMission struct {
	MissionID   string    `gorm:"primaryKey;not null" json:"mission_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Exp         int       `json:"exp"`
	Coins       int       `json:"coins"`
	StartDate   time.Time `json:"start_date"`
	Target      int       `json:"target"`
	EndDate     time.Time `json:"end_date"`
}

type GamMissionProgress struct {
	ID   string    `gorm:"primaryKey;not null" json:"id"`
	GamMissionID string `json:"gam_mission_id"`
	UserID string `json:"user_id"`
	Progress int `json:"progress"`
}

func MigrateUsers(db *gorm.DB) error{
	// db.Migrator().CreateTable(Messages{})
	err := db.AutoMigrate(&Users{}, &Posts{}, &Comments{}, &Messages{}, &GamMission{}, &GamMissionProgress{})
	return err
}