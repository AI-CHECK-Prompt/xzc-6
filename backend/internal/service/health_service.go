package service

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"

	"windpower-monitor/internal/model"
	"windpower-monitor/internal/repository"
	"windpower-monitor/pkg/redis"
)

type HealthService interface {
	CalculateHealthIndex(turbineID uint, batchID string) (*model.HealthSnapshot, error)
	CalculateAllTurbines() error
	TriggerIncrementalCalc(turbineID uint) error

	GetHealthSnapshot(turbineID uint) (*model.HealthSnapshot, error)
	GetHealthHistory(turbineID uint, startTime, endTime string) ([]model.HealthSnapshot, error)

	GetActiveAlert(turbineID uint) (*model.HealthAlert, error)
	GetAllActiveAlerts() ([]model.HealthAlert, error)

	ManualAdjustHealth(turbineID uint, value float64, reason, operator string) error
	GetAdjustments(turbineID uint) ([]model.ManualAdjustment, error)

	GetTemplate(model string) (*model.HealthTemplate, error)
	CreateTemplate(template *model.HealthTemplate) error
	UpdateTemplate(template *model.HealthTemplate) error
	DeleteTemplate(id uint) error
	GetAllTemplates() ([]model.HealthTemplate, error)

	GetConfig() (*model.HealthConfig, error)
	InitDefaultConfig() error
	UpdateConfig(config *model.HealthConfig) error

	BackfillHealth(turbineID uint, startTime, endTime time.Time) error

	GetTurbineHealthRanking(sortBy string, limit int) ([]model.WindTurbine, error)

	HandleTurbineStatusChange(turbineID uint, oldStatus, newStatus string) error
}

type healthService struct {
	healthRepo  repository.HealthRepository
	sensorRepo  repository.SensorRepository
	turbineRepo repository.TurbineRepository
}

func NewHealthService() HealthService {
	return &healthService{
		healthRepo:  repository.NewHealthRepository(),
		sensorRepo:  repository.NewSensorRepository(),
		turbineRepo: repository.NewTurbineRepository(),
	}
}

func (s *healthService) CalculateHealthIndex(turbineID uint, batchID string) (*model.HealthSnapshot, error) {
	lockKey := "health:calc:" + strconv.Itoa(int(turbineID))
	token, err := redis.AcquireLock(lockKey, 30)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, errors.New("calculation in progress")
	}
	defer redis.ReleaseLock(lockKey, token)

	existingRecord, _ := s.healthRepo.GetCalcRecord(turbineID, batchID)
	if existingRecord.ID > 0 && existingRecord.IsProcessed {
		log.Printf("【健康计算-去重】风机%d本次计算已处理，batchID=%s", turbineID, batchID)
		return nil, errors.New("already processed")
	}

	turbine, err := s.turbineRepo.GetByID(turbineID)
	if err != nil {
		return nil, err
	}

	if turbine.Status != "running" {
		log.Printf("【健康计算-跳过】风机%d状态为%s，跳过计算", turbineID, turbine.Status)
		return nil, errors.New("turbine not running")
	}

	activeAdjustment, err := s.healthRepo.GetActiveAdjustment(turbineID)
	if err == nil && activeAdjustment.ID > 0 {
		log.Printf("【健康计算-人工调整】风机%d存在有效人工调整，使用调整值%.2f", turbineID, activeAdjustment.AdjustedValue)
		snapshot := &model.HealthSnapshot{
			TurbineID:   turbineID,
			HealthIndex: activeAdjustment.AdjustedValue,
			Timestamp:   time.Now(),
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			DataQuality: 100,
			CreatedAt:   time.Now(),
		}
		if err := s.healthRepo.CreateSnapshot(snapshot); err != nil {
			return nil, err
		}
		if err := s.saveCalcRecord(turbineID, batchID, activeAdjustment.AdjustedValue); err != nil {
			return nil, err
		}
		return snapshot, nil
	}

	config, err := s.getConfigOrDefault()
	if err != nil {
		return nil, err
	}

	template, err := s.getTemplate(turbine.Model)
	if err != nil {
		return nil, err
	}

	sensorDataList, err := s.sensorRepo.GetByTimeRange(
		turbineID,
		time.Now().Add(-time.Duration(config.SmoothingWindowMinutes)*time.Minute).Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	if len(sensorDataList) == 0 {
		return nil, errors.New("no sensor data available")
	}

	scores, dataQuality := s.calculateMetricScores(sensorDataList, template, config)

	healthIndex := scores.RPMScore*template.RPMWeight +
		scores.PowerScore*template.PowerWeight +
		scores.TempScore*template.TempWeight +
		scores.VibrationScore*template.VibrationWeight

	healthIndex = math.Max(0, math.Min(100, healthIndex))

	snapshot := &model.HealthSnapshot{
		TurbineID:      turbineID,
		HealthIndex:    healthIndex,
		Timestamp:      time.Now(),
		StartTime:      time.Now(),
		EndTime:        time.Now(),
		DataQuality:    dataQuality,
		RPMScore:       scores.RPMScore,
		PowerScore:     scores.PowerScore,
		TempScore:      scores.TempScore,
		VibrationScore: scores.VibrationScore,
		CreatedAt:      time.Now(),
	}

	if err := s.compressAndSaveSnapshot(snapshot); err != nil {
		return nil, err
	}

	if err := s.processAlert(turbineID, healthIndex, config); err != nil {
		return nil, err
	}

	if err := s.saveCalcRecord(turbineID, batchID, healthIndex); err != nil {
		return nil, err
	}

	log.Printf("【健康计算-完成】风机%d健康指数: %.2f, 数据质量: %.2f%%", turbineID, healthIndex, dataQuality)
	return snapshot, nil
}

func (s *healthService) CalculateAllTurbines() error {
	turbines, err := s.turbineRepo.GetAll()
	if err != nil {
		return err
	}

	batchID := uuid.New().String()
	log.Printf("【健康计算-定时任务】开始批量计算，batchID=%s", batchID)

	var wg sync.WaitGroup
	runningTurbines := make([]model.WindTurbine, 0)
	for _, turbine := range turbines {
		if turbine.Status == "running" {
			runningTurbines = append(runningTurbines, turbine)
		}
	}

	wg.Add(len(runningTurbines))
	for _, turbine := range runningTurbines {
		go func(tid uint) {
			defer wg.Done()
			_, err := s.CalculateHealthIndex(tid, batchID)
			if err != nil {
				log.Printf("【健康计算-失败】风机%d计算失败: %v", tid, err)
			}
		}(turbine.ID)
	}

	wg.Wait()
	log.Printf("【健康计算-定时任务】批量计算完成，batchID=%s", batchID)
	return nil
}

func (s *healthService) TriggerIncrementalCalc(turbineID uint) error {
	batchID := uuid.New().String()
	_, err := s.CalculateHealthIndex(turbineID, batchID)
	return err
}

func (s *healthService) GetHealthSnapshot(turbineID uint) (*model.HealthSnapshot, error) {
	return s.healthRepo.GetLatestSnapshot(turbineID)
}

func (s *healthService) GetHealthHistory(turbineID uint, startTime, endTime string) ([]model.HealthSnapshot, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, err
	}

	snapshots, err := s.healthRepo.GetSnapshotsByTimeRange(turbineID, start, end)
	if err != nil {
		return nil, err
	}

	return s.expandCompressedSnapshots(snapshots), nil
}

func (s *healthService) expandCompressedSnapshots(snapshots []model.HealthSnapshot) []model.HealthSnapshot {
	var expanded []model.HealthSnapshot

	for _, snapshot := range snapshots {
		if snapshot.IsCompressed && snapshot.Count > 1 {
			interval := snapshot.EndTime.Sub(snapshot.StartTime) / time.Duration(snapshot.Count)
			for i := 0; i < snapshot.Count; i++ {
				expandedSnapshot := model.HealthSnapshot{
					ID:             snapshot.ID,
					TurbineID:      snapshot.TurbineID,
					HealthIndex:    snapshot.HealthIndex,
					Timestamp:      snapshot.StartTime.Add(interval * time.Duration(i)),
					StartTime:      snapshot.StartTime.Add(interval * time.Duration(i)),
					EndTime:        snapshot.StartTime.Add(interval * time.Duration(i+1)),
					Count:          1,
					IsCompressed:   false,
					IsBackfilled:   snapshot.IsBackfilled,
					DataQuality:    snapshot.DataQuality,
					RPMScore:       snapshot.RPMScore,
					PowerScore:     snapshot.PowerScore,
					TempScore:      snapshot.TempScore,
					VibrationScore: snapshot.VibrationScore,
					CreatedAt:      snapshot.CreatedAt,
				}
				expanded = append(expanded, expandedSnapshot)
			}
		} else {
			snapshot.Count = 1
			snapshot.IsCompressed = false
			expanded = append(expanded, snapshot)
		}
	}

	return expanded
}

func (s *healthService) GetActiveAlert(turbineID uint) (*model.HealthAlert, error) {
	return s.healthRepo.GetActiveAlert(turbineID)
}

func (s *healthService) GetAllActiveAlerts() ([]model.HealthAlert, error) {
	return s.healthRepo.GetAllActiveAlerts()
}

func (s *healthService) ManualAdjustHealth(turbineID uint, value float64, reason, operator string) error {
	if value < 0 || value > 100 {
		return errors.New("health index must be between 0 and 100")
	}

	latestSnapshot, err := s.healthRepo.GetLatestSnapshot(turbineID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var previousValue float64
	if latestSnapshot.ID > 0 {
		previousValue = latestSnapshot.HealthIndex
	}

	config, err := s.getConfigOrDefault()
	if err != nil {
		return err
	}

	adjustment := &model.ManualAdjustment{
		TurbineID:     turbineID,
		AdjustedValue: value,
		Reason:        reason,
		Operator:      operator,
		AdjustTime:    time.Now(),
		ExpiryTime:    time.Now().Add(time.Duration(config.AdjustmentExpiryHours) * time.Hour),
		IsActive:      true,
		PreviousValue: previousValue,
		CreatedAt:     time.Now(),
	}

	return s.healthRepo.CreateAdjustment(adjustment)
}

func (s *healthService) GetAdjustments(turbineID uint) ([]model.ManualAdjustment, error) {
	return s.healthRepo.GetAdjustmentsByTurbine(turbineID)
}

func (s *healthService) GetTemplate(model string) (*model.HealthTemplate, error) {
	return s.healthRepo.GetTemplateByModel(model)
}

func (s *healthService) CreateTemplate(template *model.HealthTemplate) error {
	if template.IsDefault {
		if err := s.healthRepo.GetDefaultTemplate(); err == nil {
			return errors.New("default template already exists")
		}
	}
	return s.healthRepo.CreateTemplate(template)
}

func (s *healthService) UpdateTemplate(template *model.HealthTemplate) error {
	return s.healthRepo.UpdateTemplate(template)
}

func (s *healthService) DeleteTemplate(id uint) error {
	return s.healthRepo.DeleteTemplate(id)
}

func (s *healthService) GetAllTemplates() ([]model.HealthTemplate, error) {
	return s.healthRepo.GetAllTemplates()
}

func (s *healthService) GetConfig() (*model.HealthConfig, error) {
	return s.getConfigOrDefault()
}

func (s *healthService) InitDefaultConfig() error {
	_, err := s.getConfigOrDefault()
	if err != nil {
		return err
	}

	_, err = s.healthRepo.GetDefaultTemplate()
	if err != nil {
		defaultTemplate := &model.HealthTemplate{
			TurbineModel:     "",
			IsDefault:        true,
			RPMWeight:        0.2,
			PowerWeight:      0.3,
			TempWeight:       0.25,
			VibrationWeight:  0.25,
			RPMMin:           0,
			RPMMax:           30,
			PowerMin:         0,
			PowerMax:         5000,
			TempMin:          -40,
			TempMax:          100,
			VibrationMin:     0,
			VibrationMax:     10,
		}
		return s.healthRepo.CreateTemplate(defaultTemplate)
	}
	return nil
}

func (s *healthService) UpdateConfig(config *model.HealthConfig) error {
	return s.healthRepo.UpdateConfig(config)
}

func (s *healthService) HandleTurbineStatusChange(turbineID uint, oldStatus, newStatus string) error {
	if oldStatus != "running" && newStatus == "running" {
		log.Printf("【健康计算-状态变更】风机%d从%s恢复运行，触发补算", turbineID, oldStatus)

		latestSnapshot, err := s.healthRepo.GetLatestSnapshot(turbineID)
		if err != nil {
			return nil
		}

		if latestSnapshot.ID > 0 {
			startTime := latestSnapshot.Timestamp
			endTime := time.Now()
			if err := s.BackfillHealth(turbineID, startTime, endTime); err != nil {
				log.Printf("【健康计算-补算失败】风机%d补算失败: %v", turbineID, err)
				return err
			}
		}

		go func() {
			if err := s.TriggerIncrementalCalc(turbineID); err != nil {
				log.Printf("【健康计算-增量计算】风机%d增量计算失败: %v", turbineID, err)
			}
		}()
	}
	return nil
}

func (s *healthService) BackfillHealth(turbineID uint, startTime, endTime time.Time) error {
	config, err := s.getConfigOrDefault()
	if err != nil {
		return err
	}

	turbine, err := s.turbineRepo.GetByID(turbineID)
	if err != nil {
		return err
	}

	template, err := s.getTemplate(turbine.Model)
	if err != nil {
		return err
	}

	currentTime := startTime
	for currentTime.Before(endTime) {
		windowEnd := currentTime.Add(time.Duration(config.SmoothingWindowMinutes) * time.Minute)
		if windowEnd.After(endTime) {
			windowEnd = endTime
		}

		sensorDataList, err := s.sensorRepo.GetByTimeRange(
			turbineID,
			currentTime.Format(time.RFC3339),
			windowEnd.Format(time.RFC3339),
		)
		if err != nil {
			return err
		}

		if len(sensorDataList) == 0 {
			currentTime = currentTime.Add(time.Hour)
			continue
		}

		scores, dataQuality := s.calculateMetricScores(sensorDataList, template, config)

		healthIndex := scores.RPMScore*template.RPMWeight +
			scores.PowerScore*template.PowerWeight +
			scores.TempScore*template.TempWeight +
			scores.VibrationScore*template.VibrationWeight

		healthIndex = math.Max(0, math.Min(100, healthIndex))

		snapshot := &model.HealthSnapshot{
			TurbineID:      turbineID,
			HealthIndex:    healthIndex,
			Timestamp:      currentTime.Add(time.Hour),
			StartTime:      currentTime,
			EndTime:        windowEnd,
			IsBackfilled:   true,
			DataQuality:    dataQuality,
			RPMScore:       scores.RPMScore,
			PowerScore:     scores.PowerScore,
			TempScore:      scores.TempScore,
			VibrationScore: scores.VibrationScore,
			CreatedAt:      time.Now(),
		}

		if err := s.healthRepo.CreateSnapshot(snapshot); err != nil {
			return err
		}

		currentTime = currentTime.Add(time.Hour)
	}

	return nil
}

func (s *healthService) GetTurbineHealthRanking(sortBy string, limit int) ([]model.WindTurbine, error) {
	turbines, err := s.turbineRepo.GetAll()
	if err != nil {
		return nil, err
	}

	type turbineHealth struct {
		turbine    *model.WindTurbine
		health     float64
		trend      float64
		declineCnt int
	}

	var turbineHealthList []turbineHealth

	for _, turbine := range turbines {
		snapshots, err := s.healthRepo.GetAllSnapshots(turbine.ID, 24)
		if err != nil || len(snapshots) == 0 {
			continue
		}

		currentHealth := snapshots[0].HealthIndex
		var trend float64
		var declineCnt int

		if len(snapshots) >= 2 {
			trend = currentHealth - snapshots[len(snapshots)-1].HealthIndex
			for i := 0; i < len(snapshots)-1; i++ {
				if snapshots[i].HealthIndex < snapshots[i+1].HealthIndex {
					declineCnt++
				}
			}
		}

		turbineHealthList = append(turbineHealthList, turbineHealth{
			turbine:    &turbine,
			health:     currentHealth,
			trend:      trend,
			declineCnt: declineCnt,
		})
	}

	switch sortBy {
	case "trend":
		for i := 0; i < len(turbineHealthList)-1; i++ {
			for j := i + 1; j < len(turbineHealthList); j++ {
				if turbineHealthList[i].trend > turbineHealthList[j].trend {
					turbineHealthList[i], turbineHealthList[j] = turbineHealthList[j], turbineHealthList[i]
				}
			}
		}
	case "decline":
		for i := 0; i < len(turbineHealthList)-1; i++ {
			for j := i + 1; j < len(turbineHealthList); j++ {
				if turbineHealthList[i].declineCnt < turbineHealthList[j].declineCnt {
					turbineHealthList[i], turbineHealthList[j] = turbineHealthList[j], turbineHealthList[i]
				}
			}
		}
	default:
		for i := 0; i < len(turbineHealthList)-1; i++ {
			for j := i + 1; j < len(turbineHealthList); j++ {
				if turbineHealthList[i].health > turbineHealthList[j].health {
					turbineHealthList[i], turbineHealthList[j] = turbineHealthList[j], turbineHealthList[i]
				}
			}
		}
	}

	result := make([]model.WindTurbine, 0, limit)
	for i, th := range turbineHealthList {
		if i >= limit {
			break
		}
		result = append(result, *th.turbine)
	}

	return result, nil
}

type metricScores struct {
	RPMScore       float64
	PowerScore     float64
	TempScore      float64
	VibrationScore float64
}

func (s *healthService) calculateMetricScores(dataList []model.SensorData, template *model.HealthTemplate, config *model.HealthConfig) (metricScores, float64) {
	var scores metricScores
	var validCount, totalCount int

	rpmValues := make([]float64, 0)
	powerValues := make([]float64, 0)
	tempValues := make([]float64, 0)
	vibrationValues := make([]float64, 0)

	turbineID := uint(0)
	if len(dataList) > 0 {
		turbineID = dataList[0].TurbineID
	}

	now := time.Now()

	for _, data := range dataList {
		ageMinutes := now.Sub(data.CreatedAt).Minutes()
		if ageMinutes > float64(config.MaxDataAgeMinutes) {
			continue
		}

		totalCount++

		if data.RPM >= template.RPMMin && data.RPM <= template.RPMMax {
			rpmValues = append(rpmValues, data.RPM)
			validCount++
		}

		if data.Power >= template.PowerMin && data.Power <= template.PowerMax {
			powerValues = append(powerValues, data.Power)
			validCount++
		}

		if data.Temperature >= template.TempMin && data.Temperature <= template.TempMax {
			tempValues = append(tempValues, data.Temperature)
			validCount++
		}

		if data.Vibration >= template.VibrationMin && data.Vibration <= template.VibrationMax {
			vibrationValues = append(vibrationValues, data.Vibration)
			validCount++
		}
	}

	if config.MissingDataStrategy == "use_last" && turbineID > 0 {
		if len(rpmValues) == 0 {
			if latest, err := s.sensorRepo.GetLatestValidMetric(turbineID, "rpm", template.RPMMin, template.RPMMax); err == nil && latest.ID > 0 {
				rpmValues = append(rpmValues, latest.RPM)
			}
		}
		if len(powerValues) == 0 {
			if latest, err := s.sensorRepo.GetLatestValidMetric(turbineID, "power", template.PowerMin, template.PowerMax); err == nil && latest.ID > 0 {
				powerValues = append(powerValues, latest.Power)
			}
		}
		if len(tempValues) == 0 {
			if latest, err := s.sensorRepo.GetLatestValidMetric(turbineID, "temperature", template.TempMin, template.TempMax); err == nil && latest.ID > 0 {
				tempValues = append(tempValues, latest.Temperature)
			}
		}
		if len(vibrationValues) == 0 {
			if latest, err := s.sensorRepo.GetLatestValidMetric(turbineID, "vibration", template.VibrationMin, template.VibrationMax); err == nil && latest.ID > 0 {
				vibrationValues = append(vibrationValues, latest.Vibration)
			}
		}
	}

	dataQuality := float64(validCount) / float64(totalCount*4) * 100
	if totalCount == 0 {
		dataQuality = 0
	}

	scores.RPMScore = s.calculateSingleMetricScore(rpmValues, template.RPMMin, template.RPMMax, config)
	scores.PowerScore = s.calculateSingleMetricScore(powerValues, template.PowerMin, template.PowerMax, config)
	scores.TempScore = s.calculateSingleMetricScore(tempValues, template.TempMin, template.TempMax, config)
	scores.VibrationScore = s.calculateSingleMetricScore(vibrationValues, template.VibrationMin, template.VibrationMax, config)

	return scores, dataQuality
}

func (s *healthService) calculateSingleMetricScore(values []float64, min, max float64, config *model.HealthConfig) float64 {
	if len(values) == 0 {
		switch config.MissingDataStrategy {
		case "degrade":
			return 50
		case "use_last":
			return 50
		case "mark_abnormal":
			return 0
		default:
			return 50
		}
	}

	avg := 0.0
	for _, v := range values {
		avg += v
	}
	avg /= float64(len(values))

	numerator := avg - min
	denominator := max - min

	if denominator == 0 {
		return 100
	}

	score := (numerator / denominator) * 100

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

func (s *healthService) processAlert(turbineID uint, healthIndex float64, config *model.HealthConfig) error {
	activeAlert, err := s.healthRepo.GetActiveAlert(turbineID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	latestSnapshot, err := s.healthRepo.GetLatestSnapshot(turbineID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var previousHealth float64
	if latestSnapshot.ID > 0 {
		previousHealth = latestSnapshot.HealthIndex
	}

	if healthIndex < config.AlertThreshold {
		if activeAlert.ID == 0 {
			alertLevel := s.determineAlertLevel(healthIndex)
			alert := &model.HealthAlert{
				TurbineID:        turbineID,
				AlertLevel:       alertLevel,
				CurrentHealth:    healthIndex,
				PreviousHealth:   previousHealth,
				Status:           "active",
				TriggerTime:      time.Now(),
				RecoveryStartTime: nil,
				RecoveryTime:     nil,
				Message:          fmt.Sprintf("健康指数低于告警阈值%.2f，当前值%.2f", config.AlertThreshold, healthIndex),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			return s.healthRepo.CreateAlert(alert)
		} else {
			activeAlert.RecoveryStartTime = nil
			if healthIndex < activeAlert.CurrentHealth {
				activeAlert.DeclineCount++
				activeAlert.PreviousHealth = activeAlert.CurrentHealth
				activeAlert.CurrentHealth = healthIndex
				activeAlert.AlertLevel = s.determineAlertLevel(healthIndex)
				activeAlert.Duration = int(time.Since(activeAlert.TriggerTime).Seconds())
				activeAlert.Message = fmt.Sprintf("健康指数持续下降，当前值%.2f，累计下降%d次", healthIndex, activeAlert.DeclineCount)
				activeAlert.UpdatedAt = time.Now()
				return s.healthRepo.UpdateAlert(activeAlert)
			}
		}
	} else {
		if activeAlert.ID > 0 && healthIndex >= config.RecoveryThreshold {
			if activeAlert.RecoveryStartTime == nil {
				recoveryStart := time.Now()
				activeAlert.RecoveryStartTime = &recoveryStart
				activeAlert.CurrentHealth = healthIndex
				activeAlert.UpdatedAt = time.Now()
				return s.healthRepo.UpdateAlert(activeAlert)
			} else {
				recoveryDuration := time.Since(*activeAlert.RecoveryStartTime).Seconds()
				if recoveryDuration >= float64(config.RecoveryDuration) {
					recoveryTime := time.Now()
					activeAlert.Status = "recovered"
					activeAlert.RecoveryTime = &recoveryTime
					activeAlert.Duration = int(time.Since(activeAlert.TriggerTime).Seconds())
					activeAlert.Message = fmt.Sprintf("健康指数恢复至%.2f，已超过恢复阈值%.2f并持续%.0f秒", healthIndex, config.RecoveryThreshold, recoveryDuration)
					activeAlert.UpdatedAt = time.Now()
					return s.healthRepo.UpdateAlert(activeAlert)
				} else {
					activeAlert.CurrentHealth = healthIndex
					activeAlert.Message = fmt.Sprintf("健康指数恢复中，当前值%.2f，已持续%.0f秒（需要%.0f秒）", healthIndex, recoveryDuration, float64(config.RecoveryDuration))
					activeAlert.UpdatedAt = time.Now()
					return s.healthRepo.UpdateAlert(activeAlert)
				}
			}
		}
	}

	return nil
}

func (s *healthService) determineAlertLevel(healthIndex float64) string {
	if healthIndex < 30 {
		return "critical"
	} else if healthIndex < 50 {
		return "major"
	} else {
		return "minor"
	}
}

func (s *healthService) compressAndSaveSnapshot(snapshot *model.HealthSnapshot) error {
	config, err := s.getConfigOrDefault()
	if err != nil {
		return err
	}

	if !config.CompressionEnabled {
		return s.healthRepo.CreateSnapshot(snapshot)
	}

	latestSnapshot, err := s.healthRepo.GetLatestSnapshot(snapshot.TurbineID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if latestSnapshot.ID > 0 &&
		math.Abs(latestSnapshot.HealthIndex-snapshot.HealthIndex) < 0.01 {
		if !latestSnapshot.IsCompressed {
			latestSnapshot.IsCompressed = true
			latestSnapshot.Count = 2
			latestSnapshot.StartTime = latestSnapshot.Timestamp
		} else {
			latestSnapshot.Count++
		}
		latestSnapshot.EndTime = snapshot.Timestamp
		latestSnapshot.HealthIndex = snapshot.HealthIndex
		latestSnapshot.DataQuality = snapshot.DataQuality
		latestSnapshot.RPMScore = snapshot.RPMScore
		latestSnapshot.PowerScore = snapshot.PowerScore
		latestSnapshot.TempScore = snapshot.TempScore
		latestSnapshot.VibrationScore = snapshot.VibrationScore
		return s.healthRepo.UpdateSnapshot(latestSnapshot)
	}

	return s.healthRepo.CreateSnapshot(snapshot)
}

func (s *healthService) getTemplate(model string) (*model.HealthTemplate, error) {
	template, err := s.healthRepo.GetTemplateByModel(model)
	if err == nil && template.ID > 0 {
		return template, nil
	}

	return s.healthRepo.GetDefaultTemplate()
}

func (s *healthService) getConfigOrDefault() (*model.HealthConfig, error) {
	config, err := s.healthRepo.GetConfig()
	if err == nil && config.ID > 0 {
		return config, nil
	}

	defaultConfig := &model.HealthConfig{
		AlertThreshold:         70,
		RecoveryThreshold:      85,
		RecoveryDuration:       3600,
		SmoothingWindowMinutes: 60,
		MissingDataStrategy:    "degrade",
		MaxDataAgeMinutes:      120,
		AdjustmentExpiryHours:  24,
		CompressionEnabled:     true,
	}

	if err := s.healthRepo.CreateConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

func (s *healthService) saveCalcRecord(turbineID uint, batchID string, resultIndex float64) error {
	record := &model.CalcRecord{
		TurbineID:   turbineID,
		BatchID:     batchID,
		CalcTime:    time.Now(),
		ResultIndex: resultIndex,
		IsProcessed: true,
	}
	return s.healthRepo.CreateCalcRecord(record)
}
