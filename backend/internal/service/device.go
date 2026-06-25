package service

import (
	"time"

	"delugewarning/internal/model"
	"delugewarning/internal/repository"
)

// DeviceService 处理读数接入（同步落库，慢逻辑交给 AlertService 入队）。
type DeviceService struct {
	repo  *repository.Repo
	alert *AlertService
}

func NewDeviceService(repo *repository.Repo, alert *AlertService) *DeviceService {
	return &DeviceService{repo: repo, alert: alert}
}

// Ingest 同步路径：落库 + 触发规则匹配（命中则入队），必须快。
func (s *DeviceService) Ingest(deviceID string, value float64, unit string, reportedAt time.Time) error {
	dev, err := s.repo.GetDevice(deviceID)
	if err != nil {
		return err
	}
	if reportedAt.IsZero() {
		reportedAt = time.Now()
	}
	if unit == "" {
		unit = "m"
	}
	rd := &model.Reading{DeviceID: deviceID, Value: value, Unit: unit, ReportedAt: reportedAt}
	if _, err := s.repo.InsertReading(rd); err != nil {
		return err
	}
	// 规则匹配 + 入队（不阻塞）
	s.alert.HandleReading(rd, dev)
	return nil
}

func (s *DeviceService) List() ([]model.Device, error) {
	return s.repo.ListDevices()
}

func (s *DeviceService) Latest(deviceID string) (*model.Reading, error) {
	return s.repo.LatestReading(deviceID)
}

func (s *DeviceService) Trend(deviceID, metric string, from, to time.Time) ([]model.DailyStat, error) {
	if metric == "" {
		metric = "water_level"
	}
	return s.repo.TrendStats(deviceID, metric, from, to)
}
