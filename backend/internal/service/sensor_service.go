package service

import (
	"encoding/json"
	"strconv"
	"time"

	"windpower-monitor/internal/model"
	"windpower-monitor/internal/repository"
	"windpower-monitor/pkg/redis"
)

type SensorService interface {
	CollectSensorData(sensorData *model.SensorData) error
	GetSensorDataByTurbineID(turbineID uint) ([]model.SensorData, error)
	GetTurbineStatus(turbineID uint) (*model.TurbineStatus, error)
	GetTrendData(turbineID uint, startTime, endTime string) ([]model.SensorData, error)
	GetTurbineStatistics(turbineID uint, startTime, endTime string) (*model.TurbineStatistics, error)
	GetAllTurbineStatistics(startTime, endTime string) ([]model.TurbineStatistics, error)
	ExportData(turbineID uint, startTime, endTime string) ([]model.SensorData, error)
}

type sensorService struct {
	repo repository.SensorRepository
}

func NewSensorService() SensorService {
	return &sensorService{
		repo: repository.NewSensorRepository(),
	}
}

func (s *sensorService) CollectSensorData(sensorData *model.SensorData) error {
	err := s.repo.Create(sensorData)
	if err != nil {
		return err
	}

	status := &model.TurbineStatus{
		TurbineID:   sensorData.TurbineID,
		RPM:         sensorData.RPM,
		Power:       sensorData.Power,
		Temperature: sensorData.Temperature,
		Humidity:    sensorData.Humidity,
		Vibration:   sensorData.Vibration,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		return err
	}

	key := "turbine:status:" + strconv.Itoa(int(sensorData.TurbineID))
	err = redis.Set(key, string(jsonData), 0)
	if err != nil {
		return err
	}

	return nil
}

func (s *sensorService) GetSensorDataByTurbineID(turbineID uint) ([]model.SensorData, error) {
	return s.repo.GetByTurbineID(turbineID)
}

func (s *sensorService) GetTurbineStatus(turbineID uint) (*model.TurbineStatus, error) {
	key := "turbine:status:" + strconv.Itoa(int(turbineID))
	data, err := redis.Get(key)
	if err != nil {
		sensorData, err := s.repo.GetLatestByTurbineID(turbineID)
		if err != nil {
			return nil, err
		}
		return &model.TurbineStatus{
			TurbineID:   sensorData.TurbineID,
			RPM:         sensorData.RPM,
			Power:       sensorData.Power,
			Temperature: sensorData.Temperature,
			Humidity:    sensorData.Humidity,
			Vibration:   sensorData.Vibration,
			Timestamp:   sensorData.CreatedAt.Format(time.RFC3339),
		}, nil
	}

	var status model.TurbineStatus
	err = json.Unmarshal([]byte(data), &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (s *sensorService) GetTrendData(turbineID uint, startTime, endTime string) ([]model.SensorData, error) {
	return s.repo.GetByTimeRange(turbineID, startTime, endTime)
}

func (s *sensorService) GetTurbineStatistics(turbineID uint, startTime, endTime string) (*model.TurbineStatistics, error) {
	return s.repo.GetStatisticsByTurbine(turbineID, startTime, endTime)
}

func (s *sensorService) GetAllTurbineStatistics(startTime, endTime string) ([]model.TurbineStatistics, error) {
	return s.repo.GetAllStatistics(startTime, endTime)
}

func (s *sensorService) ExportData(turbineID uint, startTime, endTime string) ([]model.SensorData, error) {
	return s.repo.GetByTimeRange(turbineID, startTime, endTime)
}
