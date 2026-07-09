package repository

import (
	"windpower-monitor/internal/model"
	"windpower-monitor/pkg/database"
)

type SensorRepository interface {
	Create(sensorData *model.SensorData) error
	GetByTurbineID(turbineID uint) ([]model.SensorData, error)
	GetLatestByTurbineID(turbineID uint) (*model.SensorData, error)
	GetLatestValidMetric(turbineID uint, metric string, min, max float64) (*model.SensorData, error)
	GetByTimeRange(turbineID uint, startTime, endTime string) ([]model.SensorData, error)
	GetStatisticsByTurbine(turbineID uint, startTime, endTime string) (*model.TurbineStatistics, error)
	GetAllStatistics(startTime, endTime string) ([]model.TurbineStatistics, error)
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

func (r *sensorRepository) GetLatestValidMetric(turbineID uint, metric string, min, max float64) (*model.SensorData, error) {
	var sensorData model.SensorData
	query := database.DB.Where("turbine_id = ?", turbineID)

	switch metric {
	case "rpm":
		query = query.Where("rpm >= ? AND rpm <= ?", min, max)
	case "power":
		query = query.Where("power >= ? AND power <= ?", min, max)
	case "temperature":
		query = query.Where("temperature >= ? AND temperature <= ?", min, max)
	case "vibration":
		query = query.Where("vibration >= ? AND vibration <= ?", min, max)
	}

	err := query.Order("created_at DESC").First(&sensorData).Error
	return &sensorData, err
}

func (r *sensorRepository) GetByTimeRange(turbineID uint, startTime, endTime string) ([]model.SensorData, error) {
	var sensorData []model.SensorData
	err := database.DB.Where("turbine_id = ? AND created_at BETWEEN ? AND ?", turbineID, startTime, endTime).
		Order("created_at ASC").Find(&sensorData).Error
	return sensorData, err
}

func (r *sensorRepository) GetStatisticsByTurbine(turbineID uint, startTime, endTime string) (*model.TurbineStatistics, error) {
	var stats model.TurbineStatistics
	err := database.DB.Model(&model.SensorData{}).
		Where("turbine_id = ? AND created_at BETWEEN ? AND ?", turbineID, startTime, endTime).
		Select("turbine_id, COUNT(*) as count, AVG(power) as avg_power, MAX(power) as max_power, MIN(power) as min_power, AVG(temperature) as avg_temperature, AVG(vibration) as avg_vibration").
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *sensorRepository) GetAllStatistics(startTime, endTime string) ([]model.TurbineStatistics, error) {
	var stats []model.TurbineStatistics
	err := database.DB.Model(&model.SensorData{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("turbine_id, COUNT(*) as count, AVG(power) as avg_power, MAX(power) as max_power, MIN(power) as min_power, AVG(temperature) as avg_temperature, AVG(vibration) as avg_vibration").
		Group("turbine_id").Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}
