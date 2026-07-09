package repository

import (
	"time"

	"windpower-monitor/internal/model"
	"windpower-monitor/pkg/database"
)

type HealthRepository interface {
	CreateTemplate(template *model.HealthTemplate) error
	GetTemplateByModel(model string) (*model.HealthTemplate, error)
	GetDefaultTemplate() (*model.HealthTemplate, error)
	UpdateTemplate(template *model.HealthTemplate) error
	DeleteTemplate(id uint) error
	GetAllTemplates() ([]model.HealthTemplate, error)

	GetConfig() (*model.HealthConfig, error)
	CreateConfig(config *model.HealthConfig) error
	UpdateConfig(config *model.HealthConfig) error

	CreateSnapshot(snapshot *model.HealthSnapshot) error
	GetLatestSnapshot(turbineID uint) (*model.HealthSnapshot, error)
	GetSnapshotsByTimeRange(turbineID uint, startTime, endTime time.Time) ([]model.HealthSnapshot, error)
	UpdateSnapshot(snapshot *model.HealthSnapshot) error
	DeleteSnapshot(id uint) error
	GetAllSnapshots(turbineID uint, limit int) ([]model.HealthSnapshot, error)

	CreateAlert(alert *model.HealthAlert) error
	GetActiveAlert(turbineID uint) (*model.HealthAlert, error)
	UpdateAlert(alert *model.HealthAlert) error
	GetAlertsByTurbine(turbineID uint) ([]model.HealthAlert, error)
	GetAllActiveAlerts() ([]model.HealthAlert, error)

	CreateAdjustment(adjustment *model.ManualAdjustment) error
	GetActiveAdjustment(turbineID uint) (*model.ManualAdjustment, error)
	UpdateAdjustment(adjustment *model.ManualAdjustment) error
	GetAdjustmentsByTurbine(turbineID uint) ([]model.ManualAdjustment, error)

	CreateCalcRecord(record *model.CalcRecord) error
	GetCalcRecord(turbineID uint, batchID string) (*model.CalcRecord, error)
	UpdateCalcRecord(record *model.CalcRecord) error
	DeleteExpiredCalcRecords(hours int) error
}

type healthRepository struct{}

func NewHealthRepository() HealthRepository {
	return &healthRepository{}
}

func (r *healthRepository) CreateTemplate(template *model.HealthTemplate) error {
	return database.DB.Create(template).Error
}

func (r *healthRepository) GetTemplateByModel(model string) (*model.HealthTemplate, error) {
	var template model.HealthTemplate
	err := database.DB.Where("turbine_model = ?", model).First(&template).Error
	return &template, err
}

func (r *healthRepository) GetDefaultTemplate() (*model.HealthTemplate, error) {
	var template model.HealthTemplate
	err := database.DB.Where("is_default = ?", true).First(&template).Error
	return &template, err
}

func (r *healthRepository) UpdateTemplate(template *model.HealthTemplate) error {
	return database.DB.Save(template).Error
}

func (r *healthRepository) DeleteTemplate(id uint) error {
	return database.DB.Delete(&model.HealthTemplate{}, id).Error
}

func (r *healthRepository) GetAllTemplates() ([]model.HealthTemplate, error) {
	var templates []model.HealthTemplate
	err := database.DB.Find(&templates).Error
	return templates, err
}

func (r *healthRepository) GetConfig() (*model.HealthConfig, error) {
	var config model.HealthConfig
	err := database.DB.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *healthRepository) CreateConfig(config *model.HealthConfig) error {
	return database.DB.Create(config).Error
}

func (r *healthRepository) UpdateConfig(config *model.HealthConfig) error {
	return database.DB.Save(config).Error
}

func (r *healthRepository) CreateSnapshot(snapshot *model.HealthSnapshot) error {
	return database.DB.Create(snapshot).Error
}

func (r *healthRepository) GetLatestSnapshot(turbineID uint) (*model.HealthSnapshot, error) {
	var snapshot model.HealthSnapshot
	err := database.DB.Where("turbine_id = ?", turbineID).Order("timestamp DESC").First(&snapshot).Error
	return &snapshot, err
}

func (r *healthRepository) GetSnapshotsByTimeRange(turbineID uint, startTime, endTime time.Time) ([]model.HealthSnapshot, error) {
	var snapshots []model.HealthSnapshot
	err := database.DB.Where("turbine_id = ? AND timestamp BETWEEN ? AND ?", turbineID, startTime, endTime).
		Order("timestamp ASC").Find(&snapshots).Error
	return snapshots, err
}

func (r *healthRepository) UpdateSnapshot(snapshot *model.HealthSnapshot) error {
	return database.DB.Save(snapshot).Error
}

func (r *healthRepository) DeleteSnapshot(id uint) error {
	return database.DB.Delete(&model.HealthSnapshot{}, id).Error
}

func (r *healthRepository) GetAllSnapshots(turbineID uint, limit int) ([]model.HealthSnapshot, error) {
	var snapshots []model.HealthSnapshot
	err := database.DB.Where("turbine_id = ?", turbineID).Order("timestamp DESC").Limit(limit).Find(&snapshots).Error
	return snapshots, err
}

func (r *healthRepository) CreateAlert(alert *model.HealthAlert) error {
	return database.DB.Create(alert).Error
}

func (r *healthRepository) GetActiveAlert(turbineID uint) (*model.HealthAlert, error) {
	var alert model.HealthAlert
	err := database.DB.Where("turbine_id = ? AND status = ?", turbineID, "active").First(&alert).Error
	return &alert, err
}

func (r *healthRepository) UpdateAlert(alert *model.HealthAlert) error {
	return database.DB.Save(alert).Error
}

func (r *healthRepository) GetAlertsByTurbine(turbineID uint) ([]model.HealthAlert, error) {
	var alerts []model.HealthAlert
	err := database.DB.Where("turbine_id = ?", turbineID).Order("trigger_time DESC").Find(&alerts).Error
	return alerts, err
}

func (r *healthRepository) GetAllActiveAlerts() ([]model.HealthAlert, error) {
	var alerts []model.HealthAlert
	err := database.DB.Where("status = ?", "active").Order("current_health ASC").Find(&alerts).Error
	return alerts, err
}

func (r *healthRepository) CreateAdjustment(adjustment *model.ManualAdjustment) error {
	return database.DB.Create(adjustment).Error
}

func (r *healthRepository) GetActiveAdjustment(turbineID uint) (*model.ManualAdjustment, error) {
	var adjustment model.ManualAdjustment
	err := database.DB.Where("turbine_id = ? AND is_active = ? AND expiry_time > ?", turbineID, true, time.Now()).
		Order("adjust_time DESC").First(&adjustment).Error
	return &adjustment, err
}

func (r *healthRepository) UpdateAdjustment(adjustment *model.ManualAdjustment) error {
	return database.DB.Save(adjustment).Error
}

func (r *healthRepository) GetAdjustmentsByTurbine(turbineID uint) ([]model.ManualAdjustment, error) {
	var adjustments []model.ManualAdjustment
	err := database.DB.Where("turbine_id = ?", turbineID).Order("adjust_time DESC").Find(&adjustments).Error
	return adjustments, err
}

func (r *healthRepository) CreateCalcRecord(record *model.CalcRecord) error {
	return database.DB.Create(record).Error
}

func (r *healthRepository) GetCalcRecord(turbineID uint, batchID string) (*model.CalcRecord, error) {
	var record model.CalcRecord
	err := database.DB.Where("turbine_id = ? AND batch_id = ?", turbineID, batchID).First(&record).Error
	return &record, err
}

func (r *healthRepository) UpdateCalcRecord(record *model.CalcRecord) error {
	return database.DB.Save(record).Error
}

func (r *healthRepository) DeleteExpiredCalcRecords(hours int) error {
	return database.DB.Where("created_at < ?", time.Now().Add(-time.Duration(hours)*time.Hour)).
		Delete(&model.CalcRecord{}).Error
}
