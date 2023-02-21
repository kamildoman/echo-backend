package functions

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/storage"
	uuid "github.com/satori/go.uuid"
)

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
	ID   string    `json:"id"`
	GamMissionID string `json:"gam_mission_id"`
	UserID string `json:"user_id"`
	Progress int `json:"progress"`
	GamMission GamMission `gorm:"foreignkey:GamMissionID"`
	User User `gorm:"foreignkey:UserID"`
}

type GamMissionProgressMultiple struct {
    GamMissionProgress
    UserIDs []string `json:"user_ids"`
}

func CreateMissionProgress (context *fiber.Ctx) error{
	missionMultiple := GamMissionProgressMultiple{}
    err := context.BodyParser(&missionMultiple)
    if err != nil {
        context.Status(http.StatusUnprocessableEntity).JSON(
            &fiber.Map{"message": "request failed"})
        return err
    }

	for _, UserID := range missionMultiple.UserIDs {
		id := uuid.NewV4().String()
        mission := GamMissionProgress{
            ID: id,
            UserID: UserID,
            GamMissionID: missionMultiple.GamMissionID,
			Progress: missionMultiple.Progress,
        }

        //save message to database
        err = storage.DB.Create(&mission).Error
        if err != nil {
            context.Status(http.StatusBadRequest).JSON(
                &fiber.Map{"message": "could not send the message"})
            return err
        }
    }

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "mission progress created"})
	return nil
}


func GetAllMissionProgress(context *fiber.Ctx) error {
	var missionProgress []*GamMissionProgress

	db := storage.DB

	// Get all posts within the specified range and preload the related data
	db.Preload("User").Preload("GamMission").Find(&missionProgress)

	// Construct the final response
	res := make([]map[string]interface{}, len(missionProgress))
	for i, singleProgress := range missionProgress {
		postMap := make(map[string]interface{})
		postMap["progress"] = singleProgress.Progress
		postMap["id"] = singleProgress.ID
		postMap["mission"] = singleProgress.GamMission
		postMap["user"] = singleProgress.User
		res[i] = postMap
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"data": res})

	return nil
}

func GetUsersMissionProgress(context *fiber.Ctx) error {
	id := context.Query("id")
	var missionProgress []*GamMissionProgress

	db := storage.DB

	// Get all posts within the specified range and preload the related data
	db.Preload("GamMission").Where("user_id = ?", id).Find(&missionProgress)

	// Construct the final response
	res := make([]map[string]interface{}, len(missionProgress))
	for i, singleProgress := range missionProgress {
		postMap := make(map[string]interface{})
		postMap["progress"] = singleProgress.Progress
		postMap["id"] = singleProgress.ID
		postMap["mission"] = singleProgress.GamMission
		postMap["user"] = singleProgress.User
		res[i] = postMap
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"data": res})

	return nil
}


func CreateMission (context *fiber.Ctx) error{
	mission := GamMission{}
	err := context.BodyParser(&mission)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	//generate unique id
	id := uuid.NewV4().String()
	mission.MissionID = id
	
	
	err = storage.DB.Create(&mission).Error
	
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create the mission"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "mission created"})
	return nil
}

func GetAllMissions(context *fiber.Ctx) error {
	var missions []GamMission
	err := storage.DB.Find(&missions).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get missions"})
		return err
	}
	context.Status(http.StatusOK).JSON(missions)
	return nil
}