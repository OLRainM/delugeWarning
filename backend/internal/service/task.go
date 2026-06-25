package service

import (
	"fmt"
	"time"

	"delugewarning/internal/model"
	"delugewarning/internal/repository"
	"delugewarning/internal/workflow"
)

func nowPtr() *time.Time { t := time.Now(); return &t }

// TaskService 网格员任务处置。
type TaskService struct {
	repo *repository.Repo
}

func NewTaskService(repo *repository.Repo) *TaskService {
	return &TaskService{repo: repo}
}

// Confirm 网格员确认接收任务，并推进预警 dispatched->confirmed。
func (s *TaskService) Confirm(taskID, userID int64) error {
	t, err := s.repo.GetTask(taskID)
	if err != nil {
		return err
	}
	if t.AssigneeID != userID {
		return fmt.Errorf("无权操作他人任务")
	}
	n, err := s.repo.UpdateTaskStatus(taskID, model.TaskPending, model.TaskHandling, nil)
	if err != nil || n == 0 {
		return fmt.Errorf("任务状态不可确认")
	}
	s.advanceAlert(t.AlertID, model.AlertDispatched, model.AlertConfirmed, userID, "网格员确认接收")
	return nil
}

// Handle 提交处置结果（说明 + 附件已通过 attachments 入库），推进 confirmed->handled。
func (s *TaskService) Handle(taskID, userID int64, remark string, atts []model.Attachment) error {
	t, err := s.repo.GetTask(taskID)
	if err != nil {
		return err
	}
	if t.AssigneeID != userID {
		return fmt.Errorf("无权操作他人任务")
	}
	n, err := s.repo.UpdateTaskStatus(taskID, model.TaskHandling, model.TaskFinished,
		map[string]interface{}{"handle_remark": remark, "finished_at": nowPtr()})
	if err != nil || n == 0 {
		return fmt.Errorf("任务状态不可处置")
	}
	for i := range atts {
		atts[i].TaskID = taskID
		_ = s.repo.InsertAttachment(&atts[i])
	}
	s.advanceAlert(t.AlertID, model.AlertConfirmed, model.AlertHandled, userID, "现场处置完成")
	return nil
}

// advanceAlert 在合法的前提下推进预警状态并记日志。
func (s *TaskService) advanceAlert(alertID int64, from, to string, userID int64, remark string) {
	if !workflow.CanTransition(from, to) {
		return
	}
	n, err := s.repo.UpdateAlertStatus(alertID, from, to, nil)
	if err != nil || n == 0 {
		return
	}
	_ = s.repo.InsertAlertLog(&model.AlertLog{AlertID: alertID, FromStatus: from,
		ToStatus: to, OperatorID: userID, Remark: remark})
}

// ListMine 我的任务列表。
func (s *TaskService) ListMine(userID int64, status string) ([]model.Task, error) {
	return s.repo.ListTasksByAssignee(userID, status)
}
