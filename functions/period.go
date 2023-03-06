package functions

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/storage"
)

type Period struct {
	PeriodID int `gorm:"primary_key" json:"period_id"`
	Year     int `json:"year"`
	Month    int `json:"month"`
}

func GetCurrentPeriodID() (int, error) {
	now := time.Now()
	year, month, _ := now.Date()
	period := Period{}
	err := storage.DB.Where("year = ? AND month = ?", year, int(month)).First(&period).Error
	if err != nil {
		return 0, err
	}
	return period.PeriodID, nil
}

func GetAllPeriods(context *fiber.Ctx) error {
	var periods []Period
	err := storage.DB.Find(&periods).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get periods"})
		return err
	}
	context.Status(http.StatusOK).JSON(periods)
	return nil
}