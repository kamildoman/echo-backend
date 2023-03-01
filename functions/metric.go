package functions

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/storage"
)

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

func CreateMetricDefinition(context *fiber.Ctx) error {
	metric := MetricDefinitions{}
	err := context.BodyParser(&metric)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	// get the last metric definition in the database
	var lastMetric MetricDefinitions
	err = storage.DB.Order("metric_def_id desc").Limit(1).Find(&lastMetric).Error
	if err == nil && len(lastMetric.MetricDefId) > 0 {
		// parse the number portion of the last MetricDefId
		lastId, _ := strconv.Atoi(lastMetric.MetricDefId[1:])
		// increment the number portion to generate the next unique id
		metric.MetricDefId = "M" + strconv.Itoa(lastId+1)
	} else {
		// set the MetricDefId to M10001 if no metrics are found in the database
		metric.MetricDefId = "M10001"
	}

	err = storage.DB.Create(&metric).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create the metric"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "metric created"})
	return nil
}

func GetAllMetricDefinitions(context *fiber.Ctx) error {
	var metric []MetricDefinitions
	err := storage.DB.Find(&metric).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get metric definitions"})
		return err
	}
	context.Status(http.StatusOK).JSON(metric)
	return nil
}