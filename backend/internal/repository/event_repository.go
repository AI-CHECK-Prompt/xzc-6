package repository

import (
	"time"

	"windpower-monitor/internal/model"
	"windpower-monitor/pkg/database"
)

type EventRepository interface {
	CreateEvent(event *model.StatusChangeEvent) error
	GetPendingEvents() ([]model.StatusChangeEvent, error)
	GetEventByID(id uint) (*model.StatusChangeEvent, error)
	UpdateEvent(event *model.StatusChangeEvent) error
	MarkEventProcessed(id uint) error
	MarkEventFailed(id uint, errorMessage string) error
	MarkEventCompensated(id uint, compensatedBy string) error
	GetFailedEvents(hours int) ([]model.StatusChangeEvent, error)
	DeleteOldEvents(hours int) error
}

type eventRepository struct{}

func NewEventRepository() EventRepository {
	return &eventRepository{}
}

func (r *eventRepository) CreateEvent(event *model.StatusChangeEvent) error {
	return database.DB.Create(event).Error
}

func (r *eventRepository) GetPendingEvents() ([]model.StatusChangeEvent, error) {
	var events []model.StatusChangeEvent
	err := database.DB.Where("event_status = ?", "pending").
		Order("created_at ASC").
		Find(&events).Error
	return events, err
}

func (r *eventRepository) GetEventByID(id uint) (*model.StatusChangeEvent, error) {
	var event model.StatusChangeEvent
	err := database.DB.First(&event, id).Error
	return &event, err
}

func (r *eventRepository) UpdateEvent(event *model.StatusChangeEvent) error {
	return database.DB.Save(event).Error
}

func (r *eventRepository) MarkEventProcessed(id uint) error {
	now := time.Now()
	return database.DB.Model(&model.StatusChangeEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"event_status": "processed",
			"processed_at": &now,
			"updated_at":   time.Now(),
		}).Error
}

func (r *eventRepository) MarkEventFailed(id uint, errorMessage string) error {
	return database.DB.Model(&model.StatusChangeEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"event_status":  "failed",
			"error_message": errorMessage,
			"updated_at":    time.Now(),
		}).Error
}

func (r *eventRepository) MarkEventCompensated(id uint, compensatedBy string) error {
	now := time.Now()
	return database.DB.Model(&model.StatusChangeEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"event_status":   "compensated",
			"compensated_at": &now,
			"compensated_by": compensatedBy,
			"updated_at":     time.Now(),
		}).Error
}

func (r *eventRepository) GetFailedEvents(hours int) ([]model.StatusChangeEvent, error) {
	var events []model.StatusChangeEvent
	err := database.DB.Where("event_status = ? AND created_at < ?", "failed",
		time.Now().Add(-time.Duration(hours)*time.Hour)).
		Order("created_at DESC").
		Find(&events).Error
	return events, err
}

func (r *eventRepository) DeleteOldEvents(hours int) error {
	return database.DB.Where("event_status IN (?, ?, ?) AND created_at < ?",
		"processed", "compensated", "failed",
		time.Now().Add(-time.Duration(hours)*time.Hour)).
		Delete(&model.StatusChangeEvent{}).Error
}
