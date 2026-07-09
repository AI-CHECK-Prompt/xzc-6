package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type WindTurbine struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	Name        string    `gorm:"size:100" json:"name" validate:"required"`
	Code        string    `gorm:"size:50;unique" json:"code" validate:"required,unique"`
	WindFarm    string    `gorm:"size:100" json:"wind_farm" validate:"required"`
	Location    string    `gorm:"size:200" json:"location"`
	Model       string    `gorm:"size:50" json:"model"`
	Status      string    `gorm:"size:20" json:"status" validate:"required,oneof=running stopped maintenance fault"`
	InstallDate time.Time `json:"install_date"`
	Capacity    float64   `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SensorData struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	TurbineID  uint      `json:"turbine_id" validate:"required"`
	Timestamp  time.Time `json:"timestamp"`
	RPM        float64   `json:"rpm"`
	Power      float64   `json:"power"`
	Temperature float64  `json:"temperature"`
	Humidity   float64   `json:"humidity"`
	Vibration  float64   `json:"vibration"`
	CreatedAt  time.Time `json:"created_at"`
}

type TurbineStatus struct {
	TurbineID   uint    `json:"turbine_id"`
	RPM         float64 `json:"rpm"`
	Power       float64 `json:"power"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Vibration   float64 `json:"vibration"`
	Timestamp   string  `json:"timestamp"`
}

type TurbineStatistics struct {
	TurbineID       uint    `json:"turbine_id"`
	Count           int     `json:"count"`
	AvgPower        float64 `json:"avg_power"`
	MaxPower        float64 `json:"max_power"`
	MinPower        float64 `json:"min_power"`
	AvgTemperature  float64 `json:"avg_temperature"`
	AvgVibration    float64 `json:"avg_vibration"`
}

type SystemStatistics struct {
	TotalTurbines    int     `json:"total_turbines"`
	RunningTurbines  int     `json:"running_turbines"`
	FaultTurbines    int     `json:"fault_turbines"`
	MaintenanceCount int     `json:"maintenance_count"`
	AvgPower         float64 `json:"avg_power"`
	TotalPower       float64 `json:"total_power"`
}

type HealthTemplate struct {
	ID             uint    `gorm:"primary_key" json:"id"`
	TurbineModel   string  `gorm:"size:50;unique_index" json:"turbine_model"`
	IsDefault      bool    `gorm:"default:false" json:"is_default"`
	RPMWeight      float64 `gorm:"default:0.2" json:"rpm_weight"`
	PowerWeight    float64 `gorm:"default:0.3" json:"power_weight"`
	TempWeight     float64 `gorm:"default:0.25" json:"temp_weight"`
	VibrationWeight float64 `gorm:"default:0.25" json:"vibration_weight"`
	RPMMin         float64 `gorm:"default:0" json:"rpm_min"`
	RPMMax         float64 `gorm:"default:30" json:"rpm_max"`
	PowerMin       float64 `gorm:"default:0" json:"power_min"`
	PowerMax       float64 `gorm:"default:5000" json:"power_max"`
	TempMin        float64 `gorm:"default:-40" json:"temp_min"`
	TempMax        float64 `gorm:"default:100" json:"temp_max"`
	VibrationMin   float64 `gorm:"default:0" json:"vibration_min"`
	VibrationMax   float64 `gorm:"default:10" json:"vibration_max"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type HealthConfig struct {
	ID                   uint      `gorm:"primary_key" json:"id"`
	AlertThreshold       float64   `gorm:"default:70" json:"alert_threshold"`
	RecoveryThreshold    float64   `gorm:"default:85" json:"recovery_threshold"`
	RecoveryDuration     int       `gorm:"default:3600" json:"recovery_duration"`
	SmoothingWindowMinutes int     `gorm:"default:60" json:"smoothing_window_minutes"`
	MissingDataStrategy  string    `gorm:"size:20;default:degrade" json:"missing_data_strategy"`
	MaxDataAgeMinutes    int       `gorm:"default:120" json:"max_data_age_minutes"`
	AdjustmentExpiryHours int      `gorm:"default:24" json:"adjustment_expiry_hours"`
	CompressionEnabled   bool      `gorm:"default:true" json:"compression_enabled"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type HealthSnapshot struct {
	ID             uint      `gorm:"primary_key" json:"id"`
	TurbineID      uint      `json:"turbine_id"`
	HealthIndex    float64   `json:"health_index"`
	Timestamp      time.Time `json:"timestamp"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Count          int       `gorm:"default:1" json:"count"`
	IsCompressed   bool      `gorm:"default:false" json:"is_compressed"`
	IsBackfilled   bool      `gorm:"default:false" json:"is_backfilled"`
	DataQuality    float64   `json:"data_quality"`
	RPMScore       float64   `json:"rpm_score"`
	PowerScore     float64   `json:"power_score"`
	TempScore      float64   `json:"temp_score"`
	VibrationScore float64   `json:"vibration_score"`
	CreatedAt      time.Time `json:"created_at"`
}

type HealthAlert struct {
	ID               uint      `gorm:"primary_key" json:"id"`
	TurbineID        uint      `json:"turbine_id"`
	AlertLevel       string    `gorm:"size:20" json:"alert_level"`
	CurrentHealth    float64   `json:"current_health"`
	PreviousHealth   float64   `json:"previous_health"`
	Status           string    `gorm:"size:20;default:active" json:"status"`
	TriggerTime      time.Time `json:"trigger_time"`
	RecoveryStartTime *time.Time `json:"recovery_start_time"`
	RecoveryTime     *time.Time `json:"recovery_time"`
	Duration         int       `json:"duration"`
	DeclineCount     int       `gorm:"default:0" json:"decline_count"`
	Message          string    `gorm:"size:500" json:"message"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ManualAdjustment struct {
	ID             uint      `gorm:"primary_key" json:"id"`
	TurbineID      uint      `json:"turbine_id"`
	AdjustedValue  float64   `json:"adjusted_value"`
	Reason         string    `gorm:"size:500" json:"reason"`
	Operator       string    `gorm:"size:50" json:"operator"`
	AdjustTime     time.Time `json:"adjust_time"`
	ExpiryTime     time.Time `json:"expiry_time"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	PreviousValue  float64   `json:"previous_value"`
	CreatedAt      time.Time `json:"created_at"`
}

type CalcRecord struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	TurbineID   uint      `json:"turbine_id"`
	BatchID     string    `gorm:"size:36" json:"batch_id"`
	CalcTime    time.Time `json:"calc_time"`
	ResultIndex float64   `json:"result_index"`
	IsProcessed bool      `gorm:"default:false" json:"is_processed"`
	CreatedAt   time.Time `json:"created_at"`
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&WindTurbine{}, &SensorData{}, &HealthTemplate{}, &HealthConfig{}, &HealthSnapshot{}, &HealthAlert{}, &ManualAdjustment{}, &CalcRecord{})
}
