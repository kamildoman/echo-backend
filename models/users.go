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
	CreatedAt time.Time `json:"created_at"`
    UserID *string `json:"userId"`
	Likes pq.StringArray `gorm:"type:text[]" json:"likes"`
}

type Comments struct {
    ID string `gorm:"unique" json:"id"`
    Message *string `json:"message"`
	CreatedAt time.Time `json:"created_at"`
    UserID *string `json:"userId"`
    PostID *string `json:"postId"`
}

type Messages struct {
	MessageID      string `json:"message_id"`
	Title		  *string `json:"title"`
	Message       *string `json:"message"`
	SendUserID    *string `json:"send_user_id"`
	ReceiveUserID *string `json:"receive_user_id"`
	CreatedAt     time.Time    `json:"created_at"`
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

type MetricDefinitions struct {
	MetricDefId       string `gorm:"primaryKey;not null" json:"metric_def_id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	Category          string `json:"category"`
	Description       string `json:"description"`
	Exp               int    `json:"exp"`
	Coins             int    `json:"coins"`
	CalculationMethod string `json:"calculation_method"`
	Direction         string `json:"direction"`
	EndValue          int    `json:"end_value"`
	Weight            int    `json:"weight"`
	Target            string `json:"target"`
	PeriodId          int    `json:"period_id"`
}

func MigrateUsers(db *gorm.DB) error{
	// db.Migrator().CreateTable(Messages{})
	err := db.AutoMigrate(&Users{}, &Posts{}, &Comments{}, &Messages{}, &GamMission{}, &GamMissionProgress{}, &MetricDefinitions{})
	return err
}