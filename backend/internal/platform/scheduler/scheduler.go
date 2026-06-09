package scheduler

import (
	"log"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/worker"
)

var (
	GlobalScheduler *Scheduler
	once            sync.Once
)

type Scheduler struct {
	mu        sync.RWMutex
	cron      *cron.Cron
	growthSvc service.GrowthService
	entryIDs  map[string]cron.EntryID // key: job_definition_id
	running   bool
}

func Init(growthSvc service.GrowthService) *Scheduler {
	once.Do(func() {
		// robfig/cron/v3 使用 Seconds 级别（6字段：秒 分 时 日 月 周）
		GlobalScheduler = &Scheduler{
			cron:      cron.New(cron.WithSeconds()),
			growthSvc: growthSvc,
			entryIDs:  make(map[string]cron.EntryID),
		}
	})
	return GlobalScheduler
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}
	s.loadAndRegisterJobs()
	s.cron.Start()
	s.running = true
	log.Printf("[scheduler] Cron scheduler started successfully")
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	s.cron.Stop()
	s.running = false
	log.Printf("[scheduler] Cron scheduler stopped")
}

func (s *Scheduler) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[scheduler] reloading job definitions...")

	// 停止当前 cron 实例并清空
	s.cron.Stop()
	s.cron = cron.New(cron.WithSeconds())
	s.entryIDs = make(map[string]cron.EntryID)

	s.loadAndRegisterJobs()

	if s.running {
		s.cron.Start()
	}
	log.Printf("[scheduler] Cron scheduler reloaded successfully")
}

func (s *Scheduler) loadAndRegisterJobs() {
	// 一次性获取所有 ACTIVE 的任务定义
	items, _, err := s.growthSvc.AdminListSchedulerJobDefinitions("ACTIVE", "", 1, 1000)
	if err != nil {
		log.Printf("[scheduler] failed to load job definitions: %v", err)
		return
	}

	for _, item := range items {
		def := item // copy
		cronExpr := strings.TrimSpace(def.CronExpr)
		if cronExpr == "" {
			continue
		}

		// 注册定时任务
		entryID, err := s.cron.AddFunc(cronExpr, func() {
			s.dispatchJob(def)
		})
		if err != nil {
			log.Printf("[scheduler] failed to register job %s(%s) with cron expr '%s': %v", def.DisplayName, def.JobName, cronExpr, err)
			continue
		}
		s.entryIDs[def.ID] = entryID
		log.Printf("[scheduler] registered job: %s(%s) -> cron: %s", def.DisplayName, def.JobName, cronExpr)
	}
}

func (s *Scheduler) dispatchJob(def model.SchedulerJobDefinition) {
	log.Printf("[scheduler] cron triggered job: %s", def.JobName)
	// 投递定时任务至后台执行通道，因为是定时触发，操作人设为 system
	s.enqueueAsyncJob(def.JobName, "SYSTEM_TIMER", "system")
}

func (s *Scheduler) enqueueAsyncJob(jobName string, triggerSource string, operatorID string) {
	select {
	case worker.JobQueue <- worker.JobExecutionRequest{
		JobName:       jobName,
		TriggerSource: triggerSource,
		OperatorID:    operatorID,
	}:
		log.Printf("[scheduler] job %s enqueued to async worker channel", jobName)
	default:
		log.Printf("[scheduler] job %s enqueued failed, worker queue channel is full", jobName)
	}
}
