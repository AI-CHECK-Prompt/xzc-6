package repository

import (
	"windpower-monitor/internal/model"
	"windpower-monitor/pkg/database"
)

type TurbineRepository interface {
	Create(turbine *model.WindTurbine) error
	GetByID(id uint) (*model.WindTurbine, error)
	GetAll() ([]model.WindTurbine, error)
	Update(turbine *model.WindTurbine) error
	Delete(id uint) error
	GetStatistics() (*model.SystemStatistics, error)
}

type turbineRepository struct{}

func NewTurbineRepository() TurbineRepository {
	return &turbineRepository{}
}

func (r *turbineRepository) Create(turbine *model.WindTurbine) error {
	return database.DB.Create(turbine).Error
}

func (r *turbineRepository) GetByID(id uint) (*model.WindTurbine, error) {
	var turbine model.WindTurbine
	err := database.DB.First(&turbine, id).Error
	return &turbine, err
}

func (r *turbineRepository) GetAll() ([]model.WindTurbine, error) {
	var turbines []model.WindTurbine
	err := database.DB.Find(&turbines).Error
	return turbines, err
}

func (r *turbineRepository) Update(turbine *model.WindTurbine) error {
	return database.DB.Save(turbine).Error
}

func (r *turbineRepository) Delete(id uint) error {
	return database.DB.Delete(&model.WindTurbine{}, id).Error
}

func (r *turbineRepository) GetStatistics() (*model.SystemStatistics, error) {
	var stats model.SystemStatistics
	
	err := database.DB.Model(&model.WindTurbine{}).Count(&stats.TotalTurbines).Error
	if err != nil {
		return nil, err
	}
	
	err = database.DB.Model(&model.WindTurbine{}).Where("status = ?", "running").Count(&stats.RunningTurbines).Error
	if err != nil {
		return nil, err
	}
	
	err = database.DB.Model(&model.WindTurbine{}).Where("status = ?", "fault").Count(&stats.FaultTurbines).Error
	if err != nil {
		return nil, err
	}
	
	err = database.DB.Model(&model.WindTurbine{}).Where("status = ?", "maintenance").Count(&stats.MaintenanceCount).Error
	if err != nil {
		return nil, err
	}
	
	err = database.DB.Model(&model.SensorData{}).Select("AVG(power)").Scan(&stats.AvgPower).Error
	if err != nil {
		return nil, err
	}
	
	err = database.DB.Model(&model.SensorData{}).Select("SUM(power)").Scan(&stats.TotalPower).Error
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}
