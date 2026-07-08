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
}

type turbineService struct {
	repo repository.TurbineRepository
}

func NewTurbineService() TurbineService {
	return &turbineService{
		repo: repository.NewTurbineRepository(),
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
	return s.repo.Update(turbine)
}

func (s *turbineService) DeleteTurbine(id uint) error {
	return s.repo.Delete(id)
}
