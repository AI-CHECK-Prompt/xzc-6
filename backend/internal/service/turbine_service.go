package service

import (
	"windpower-monitor/internal/model"
	"windpower-monitor/internal/repository"
)

type TurbineService interface {
	CreateTurbine(turbine *model.WindTurbine) error
	GetTurbineByID(id uint) (*model.WindTurbine, error)
	GetAllTurbines() ([]model.WindTurbine, error)
	UpdateTurbine(turbine *model.WindTurbine) error
	DeleteTurbine(id uint) error
	GetSystemStatistics(startTime, endTime string) (*model.SystemStatistics, error)
}

type turbineService struct {
	repo        repository.TurbineRepository
	healthService HealthService
}

func NewTurbineService() TurbineService {
	return &turbineService{
		repo:          repository.NewTurbineRepository(),
		healthService: NewHealthService(),
	}
}

func (s *turbineService) CreateTurbine(turbine *model.WindTurbine) error {
	return s.repo.Create(turbine)
}

func (s *turbineService) GetTurbineByID(id uint) (*model.WindTurbine, error) {
	return s.repo.GetByID(id)
}

func (s *turbineService) GetAllTurbines() ([]model.WindTurbine, error) {
	return s.repo.GetAll()
}

func (s *turbineService) UpdateTurbine(turbine *model.WindTurbine) error {
	oldTurbine, err := s.repo.GetByID(turbine.ID)
	if err != nil {
		return err
	}

	if err := s.repo.Update(turbine); err != nil {
		return err
	}

	if oldTurbine.Status != turbine.Status {
		if err := s.healthService.HandleTurbineStatusChange(turbine.ID, oldTurbine.Status, turbine.Status); err != nil {
			return err
		}
	}

	return nil
}

func (s *turbineService) DeleteTurbine(id uint) error {
	return s.repo.Delete(id)
}

func (s *turbineService) GetSystemStatistics(startTime, endTime string) (*model.SystemStatistics, error) {
	return s.repo.GetStatistics(startTime, endTime)
}
