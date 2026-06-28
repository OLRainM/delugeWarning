package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"delugewarning/internal/async"
	"delugewarning/internal/model"
	"delugewarning/internal/provider"
	"delugewarning/internal/repository"
	"delugewarning/internal/rule"
	"delugewarning/internal/workflow"
)

// AlertService 负责预警的生成、派发、复核、归档（工作流核心）。
type AlertService struct {
	repo   *repository.Repo
	engine *rule.Engine
	queue  *async.Queue
	tts    provider.TTSEngine
	push   provider.Pusher
}

func NewAlertService(repo *repository.Repo, engine *rule.Engine, queue *async.Queue,
	tts provider.TTSEngine, push provider.Pusher) *AlertService {
	return &AlertService{repo: repo, engine: engine, queue: queue, tts: tts, push: push}
}

// HandleReading 同步路径：规则匹配 + 防抖，命中则把慢操作入队。必须快。
func (s *AlertService) HandleReading(rd *model.Reading, dev *model.Device) {
	matched := s.engine.Match(dev.Type, rd.Value)
	if matched == nil {
		return
	}
	if !s.engine.AllowFire(dev.ID, matched.Level, matched.CooldownSec) {
		return // 冷却期内，防抖跳过
	}
	r := *matched
	// 入队：渲染文案 -> TTS -> 创建预警 -> 派发
	s.queue.Submit(func() { s.generateAlert(r, rd, dev) })
}

// generateAlert 异步执行：生成预警并按规则直发或转待审。
func (s *AlertService) generateAlert(r model.Rule, rd *model.Reading, dev *model.Device) {
	content := s.renderContent(r, rd, dev)
	title := fmt.Sprintf("%s预警", levelName(r.Level))

	ttsKey, err := s.tts.Synthesize(context.Background(), content)
	if err != nil {
		log.Printf("[alert] TTS 合成失败: %v", err)
	}

	status := model.AlertTriggered
	if r.ReviewRequired {
		status = model.AlertPendingReview
	}
	a := &model.Alert{
		Source: model.SourceSensor, Level: r.Level, DisasterType: "flood",
		GridID: dev.GridID, DeviceID: dev.ID, Title: title, Content: content,
		TTSURL: ttsKey, // 存 ObjectKey，下发时由 broadcast 接口签名
		Status: status,
	}
	if _, err := s.repo.InsertAlert(a); err != nil {
		log.Printf("[alert] 创建预警失败: %v", err)
		return
	}
	_ = s.repo.InsertAlertLog(&model.AlertLog{AlertID: a.ID, ToStatus: status, Remark: "系统自动生成"})

	if status == model.AlertTriggered {
		s.Dispatch(a) // 直发：立即派发
	}
}

// Dispatch 将预警派发给责任网格员并推送村民。
func (s *AlertService) Dispatch(a *model.Alert) {
	n, err := s.repo.UpdateAlertStatus(a.ID, model.AlertTriggered, model.AlertDispatched, nil)
	if err != nil || n == 0 {
		return
	}
	_ = s.repo.InsertAlertLog(&model.AlertLog{AlertID: a.ID,
		FromStatus: model.AlertTriggered, ToStatus: model.AlertDispatched, Remark: "派发任务"})

	// 创建网格员处置任务
	worker, err := s.repo.GridworkerByGrid(a.GridID)
	if err == nil && worker != nil {
		task := &model.Task{AlertID: a.ID, AssigneeID: worker.ID, Status: model.TaskPending}
		if _, err := s.repo.InsertTask(task); err == nil {
			s.queue.Submit(func() {
				_ = s.push.Push(context.Background(), worker.OpenID, provider.TemplateMsg{
					TemplateID: "alert_dispatch",
					Data:       map[string]string{"title": a.Title, "content": a.Content},
				})
			})
		}
	}
}

// Review 网格员复核：confirm 发布 / modify 改文案 / cancel 撤销。
func (s *AlertService) Review(alertID, operatorID int64, action, newContent string) error {
	a, err := s.repo.GetAlert(alertID)
	if err != nil {
		return err
	}
	if a.Status != model.AlertPendingReview {
		return fmt.Errorf("预警不处于待复核状态")
	}
	switch action {
	case "cancel":
		if !workflow.CanTransition(a.Status, model.AlertCanceled) {
			return fmt.Errorf("非法流转")
		}
		_, err = s.repo.UpdateAlertStatus(alertID, model.AlertPendingReview, model.AlertCanceled,
			map[string]interface{}{"reviewed_by": operatorID})
		_ = s.repo.InsertAlertLog(&model.AlertLog{AlertID: alertID, FromStatus: model.AlertPendingReview,
			ToStatus: model.AlertCanceled, OperatorID: operatorID, Remark: "复核撤销(误报)"})
		return err
	case "modify", "confirm":
		fields := map[string]interface{}{"reviewed_by": operatorID}
		if action == "modify" && newContent != "" {
			fields["content"] = newContent
		}
		_, err = s.repo.UpdateAlertStatus(alertID, model.AlertPendingReview, model.AlertTriggered, fields)
		if err != nil {
			return err
		}
		_ = s.repo.InsertAlertLog(&model.AlertLog{AlertID: alertID, FromStatus: model.AlertPendingReview,
			ToStatus: model.AlertTriggered, OperatorID: operatorID, Remark: "复核通过发布"})
		a.Status = model.AlertTriggered
		s.Dispatch(a)
		return nil
	default:
		return fmt.Errorf("未知复核动作: %s", action)
	}
}

func (s *AlertService) renderContent(r model.Rule, rd *model.Reading, dev *model.Device) string {
	tpl := "【{level}预警】{device} 监测水位已达 {value} 米，请注意防范。"
	if t, err := s.repo.GetTemplate(r.TemplateID); err == nil && t.ContentTpl != "" {
		tpl = t.ContentTpl
	}
	rep := strings.NewReplacer(
		"{level}", levelName(r.Level), "{device}", dev.Name,
		"{value}", fmt.Sprintf("%.2f", rd.Value), "{village}", "",
	)
	return rep.Replace(tpl)
}

func levelName(level string) string {
	return map[string]string{"blue": "蓝色", "yellow": "黄色", "orange": "橙色", "red": "红色"}[level]
}

// CheckTimeouts 供 cron 调用：处理派发超时未确认任务（占位，记录日志）。
func (s *AlertService) CheckTimeouts(d time.Duration) {
	tasks, err := s.repo.TimedOutDispatched(time.Now().Add(-d))
	if err != nil {
		return
	}
	for _, t := range tasks {
		log.Printf("[alert] 任务 %d 派发超时未确认，提醒重派", t.ID)
	}
}
