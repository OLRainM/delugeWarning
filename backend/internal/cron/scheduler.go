package cron

import (
	"log"
	"time"

	"delugewarning/internal/repository"
	"delugewarning/internal/service"

	"github.com/robfig/cron/v3"
)

// Scheduler 封装进程内定时任务。
type Scheduler struct {
	c             *cron.Cron
	repo          *repository.Repo
	alert         *service.AlertService
	retentionDays int
}

func New(repo *repository.Repo, alert *service.AlertService, retentionDays int) *Scheduler {
	return &Scheduler{c: cron.New(), repo: repo, alert: alert, retentionDays: retentionDays}
}

// Start 注册并启动所有定时任务。
func (s *Scheduler) Start() {
	// 预建未来 3 天分区（每天 00:05）
	_, _ = s.c.AddFunc("5 0 * * *", s.ensurePartitions)
	// 日聚合（每天 00:30）
	_, _ = s.c.AddFunc("30 0 * * *", s.aggregate)
	// 清理过期分区（每天 01:00，聚合之后）
	_, _ = s.c.AddFunc("0 1 * * *", s.dropOld)
	// 超时重派检查（每分钟）
	_, _ = s.c.AddFunc("* * * * *", func() { s.alert.CheckTimeouts(10 * time.Minute) })

	s.ensurePartitions() // 启动即保证当天分区存在
	s.c.Start()
	log.Println("[cron] 定时任务已启动")
}

func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
}

func (s *Scheduler) ensurePartitions() {
	now := time.Now()
	for i := 0; i <= 3; i++ {
		day := now.AddDate(0, 0, i)
		if err := s.repo.EnsureDayPartition(day); err != nil {
			log.Printf("[cron] 创建分区失败 %s: %v", day.Format("2006-01-02"), err)
		}
	}
}

func (s *Scheduler) aggregate() {
	if err := s.repo.AggregateDaily(); err != nil {
		log.Printf("[cron] 日聚合失败: %v", err)
		return
	}
	log.Println("[cron] 日聚合完成")
}

func (s *Scheduler) dropOld() {
	dropped, err := s.repo.DropOldPartitions(s.retentionDays)
	if err != nil {
		log.Printf("[cron] 清理过期分区失败: %v", err)
		return
	}
	if len(dropped) > 0 {
		log.Printf("[cron] 已清理过期分区: %v", dropped)
	}
}
