package functions

import (
	"math"
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
	CalculationMethod string `json:"calculation_method"`
	Direction         string `json:"direction"`
	EndValue          int    `json:"end_value"`
	Weight            int    `json:"weight"`
	Target            string `json:"target"`
	PeriodId          int    `json:"period_id"`
	RenewNextMonth	  bool   `json:"renew_next_month"`
}

type MetricDefinitionRewards struct {
	MetricDefId  string `json:"metric_def_id"`
	Percentage float64 `json:"percentage"`
	Exp int `json:"exp"`
	Coins int `json:"coins"`
}

type Metric struct {
	MetricId string
	MetricDefId string
	EmployeeId string
	CurrentProgress int
	PeriodId int
	MetricDefinitions MetricDefinitions `gorm:"foreignkey:MetricDefId"`
}


func GetMetricsForUser(context *fiber.Ctx) error {
	userId := context.Query("id")
	var metrics []*Metric

	var periodId, err = GetCurrentPeriodID()

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the period for metrics"})
		return err
	}

	var employeeId, error = GetEmployeeIdByUserId(userId)

	if error != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get employeeId for metrics"})
		return err
	}

	err = storage.DB.Where("employee_id = ?", employeeId).Where("period_id = ?", periodId).Preload("MetricDefinitions").Find(&metrics).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the metrics"})
		return err
	}

	responseMetrics := make([]map[string]interface{}, len(metrics))

	for i, metric := range metrics {
		responseMetrics[i] = map[string]interface{}{
			"current_progress": metric.CurrentProgress,
			"metric_definitions": map[string]interface{}{
				"name":              metric.MetricDefinitions.Name,
				"type":              metric.MetricDefinitions.Type,
				"category":          metric.MetricDefinitions.Category,
				"description":       metric.MetricDefinitions.Description,
				"calculation_method": metric.MetricDefinitions.CalculationMethod,
				"direction":         metric.MetricDefinitions.Direction,
				"end_value":         metric.MetricDefinitions.EndValue,
				"weight":            metric.MetricDefinitions.Weight,
				"target":            metric.MetricDefinitions.Target,
			},
		}
	}

	context.Status(http.StatusOK).JSON(responseMetrics)
	return nil
}

func CreateMetricDefinition(context *fiber.Ctx) error {
	type MetricForm struct {
		Metric MetricDefinitions
		Rewards []MetricDefinitionRewards
	}
	metricForm := MetricForm{}
	err := context.BodyParser(&metricForm)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	// get the last metric definition in the database
	var metric = metricForm.Metric
	var rewards = metricForm.Rewards
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

	 // create metric definition rewards
	 for _, reward := range rewards {
        reward.MetricDefId = metric.MetricDefId
		reward.Percentage = math.Round(reward.Percentage * 100) / 100
        err = storage.DB.Create(&reward).Error
        if err != nil {
            context.Status(http.StatusBadRequest).JSON(
                &fiber.Map{"message": "could not create the metric definition reward"})
            return err
        }
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