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

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&WindTurbine{}, &SensorData{})
}
