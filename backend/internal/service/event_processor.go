package service

import (
	"log"
	"strconv"
	"time"

	"windpower-monitor/internal/model"
	"windpower-monitor/internal/repository"
	"windpower-monitor/pkg/redis"
)

type EventProcessor interface {
	Start()
	ProcessEvent(event *model.StatusChangeEvent) error
	ProcessPendingEvents()
	CheckConsistency()
}

type eventProcessor struct {
	eventRepo     repository.EventRepository
	healthService HealthService
	turbineRepo   repository.TurbineRepository
}

func NewEventProcessor() EventProcessor {
	return &eventProcessor{
		eventRepo:     repository.NewEventRepository(),
		healthService: NewHealthService(),
		turbineRepo:   repository.NewTurbineRepository(),
	}
}

func (p *eventProcessor) Start() {
	log.Printf("【事件处理器】启动事件消费协程")

	go func() {
		for {
			p.ProcessPendingEvents()
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			p.processFailedEvents()
			time.Sleep(60 * time.Second)
		}
	}()

	go func() {
		for {
			p.cleanupOldEvents()
			time.Sleep(3600 * time.Second)
		}
	}()
}

func (p *eventProcessor) ProcessPendingEvents() {
	events, err := p.eventRepo.GetPendingEvents()
	if err != nil {
		log.Printf("【事件处理器-错误】获取待处理事件失败: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	log.Printf("【事件处理器】发现%d个待处理事件", len(events))

	for _, event := range events {
		if err := p.ProcessEvent(&event); err != nil {
			log.Printf("【事件处理器-错误】处理事件ID=%d失败: %v", event.ID, err)
		}
	}
}

func (p *eventProcessor) ProcessEvent(event *model.StatusChangeEvent) error {
	lockKey := "event:lock:" + strconv.Itoa(int(event.ID))
	token, err := redis.AcquireLock(lockKey, 30)
	if err != nil {
		return err
	}
	if token == "" {
		return nil
	}
	defer redis.ReleaseLock(lockKey, token)

	log.Printf("【事件处理器】开始处理事件ID=%d，风机%d，%s->%s",
		event.ID, event.TurbineID, event.OldStatus, event.NewStatus)

	err = p.healthService.HandleTurbineStatusChange(event.TurbineID, event.OldStatus, event.NewStatus)
	if err != nil {
		log.Printf("【事件处理器-失败】事件ID=%d处理失败: %v", event.ID, err)

		if event.RetryCount >= event.MaxRetry {
			log.Printf("【事件处理器-放弃】事件ID=%d已达最大重试次数%d，标记为失败", event.ID, event.MaxRetry)
			return p.eventRepo.MarkEventFailed(event.ID, err.Error())
		}

		event.RetryCount++
		event.ErrorMessage = err.Error()
		event.UpdatedAt = time.Now()
		log.Printf("【事件处理器-重试】事件ID=%d重试次数+1，当前=%d", event.ID, event.RetryCount)
		return p.eventRepo.UpdateEvent(event)
	}

	log.Printf("【事件处理器-成功】事件ID=%d处理成功", event.ID)
	return p.eventRepo.MarkEventProcessed(event.ID)
}

func (p *eventProcessor) processFailedEvents() {
	failedEvents, err := p.eventRepo.GetFailedEvents(24)
	if err != nil {
		log.Printf("【事件处理器-错误】获取失败事件失败: %v", err)
		return
	}

	if len(failedEvents) == 0 {
		return
	}

	log.Printf("【事件处理器】发现%d个失败事件，触发补偿处理", len(failedEvents))

	for _, event := range failedEvents {
		if err := p.compensateEvent(&event); err != nil {
			log.Printf("【事件处理器-补偿失败】事件ID=%d补偿失败: %v", event.ID, err)
		}
	}
}

func (p *eventProcessor) compensateEvent(event *model.StatusChangeEvent) error {
	log.Printf("【事件处理器-补偿】开始补偿事件ID=%d，风机%d，回滚状态从%s到%s",
		event.ID, event.TurbineID, event.NewStatus, event.OldStatus)

	turbine, err := p.turbineRepo.GetByID(event.TurbineID)
	if err != nil {
		return err
	}

	if turbine.Status != event.NewStatus {
		log.Printf("【事件处理器-补偿】风机%d当前状态%s与预期状态%s不一致，跳过补偿",
			event.TurbineID, turbine.Status, event.NewStatus)
		return p.eventRepo.MarkEventCompensated(event.ID, "auto")
	}

	turbine.Status = event.OldStatus
	if err := p.turbineRepo.Update(turbine); err != nil {
		return err
	}

	log.Printf("【事件处理器-补偿成功】风机%d状态已回滚至%s", event.TurbineID, event.OldStatus)
	return p.eventRepo.MarkEventCompensated(event.ID, "auto")
}

func (p *eventProcessor) cleanupOldEvents() {
	if err := p.eventRepo.DeleteOldEvents(168); err != nil {
		log.Printf("【事件处理器-错误】清理旧事件失败: %v", err)
	} else {
		log.Printf("【事件处理器】清理7天前的已处理事件完成")
	}
}

func (p *eventProcessor) CheckConsistency() {
	log.Printf("【一致性检查】开始执行状态与健康数据一致性检查")

	turbines, err := p.turbineRepo.GetAll()
	if err != nil {
		log.Printf("【一致性检查-错误】获取风机列表失败: %v", err)
		return
	}

	consistencyIssues := 0
	for _, turbine := range turbines {
		if turbine.Status == "running" {
			snapshot, err := p.healthService.GetHealthSnapshot(turbine.ID)
			if err != nil {
				log.Printf("【一致性检查-警告】风机%d状态为运行中，但无法获取健康快照: %v", turbine.ID, err)
				consistencyIssues++
				continue
			}

			if snapshot.ID == 0 {
				log.Printf("【一致性检查-警告】风机%d状态为运行中，但无健康快照数据", turbine.ID)
				consistencyIssues++
				continue
			}

			timeDiff := time.Since(snapshot.Timestamp)
			if timeDiff > 2*time.Hour {
				log.Printf("【一致性检查-警告】风机%d状态为运行中，但最新健康快照已过期%.0f分钟", turbine.ID, timeDiff.Minutes())
				consistencyIssues++
			}
		}
	}

	if consistencyIssues > 0 {
		log.Printf("【一致性检查-告警】发现%d个一致性问题，请及时处理", consistencyIssues)
	} else {
		log.Printf("【一致性检查-完成】所有风机状态与健康数据一致")
	}
}
