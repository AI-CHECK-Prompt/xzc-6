package repository

import (
	"windpower-monitor/internal/model"
	"windpower-monitor/pkg/database"
)

type SensorRepository interface {
	Create(sensorData *model.SensorData) error
	GetByTurbineID(turbineID uint) ([]model.SensorData, error)
	GetLatestByTurbineID(turbineID uint) (*model.SensorData, error)
}

type sensorRepository struct{}

func NewSensorRepository() SensorRepository {
	return &sensorRepository{}
}

func (r *sensorRepository) Create(sensorData *model.SensorData) error {
	return database.DB.Create(sensorData).Error
}

func (r *sensorRepository) GetByTurbineID(turbineID uint) ([]model.SensorData, error) {
	var sensorData []model.SensorData
	err := database.DB.Where("turbine_id = ?", turbineID).Order("created_at DESC").Limit(100).Find(&sensorData).Error
	return sensorData, err
}

func (r *sensorRepository) GetLatestByTurbineID(turbineID uint) (*model.SensorData, error) {
	var sensorData model.SensorData
	err := database.DB.Where("turbine_id = ?", turbineID).Order("created_at DESC").First(&sensorData).Error
	return &sensorData, err
}
