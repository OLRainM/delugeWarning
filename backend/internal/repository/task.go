package repository

import (
	"time"

	"delugewarning/internal/model"
)

// ---------- 任务 ----------

func (r *Repo) InsertTask(t *model.Task) (int64, error) {
	t.CreatedAt = time.Now()
	var id int64
	err := r.sess().InsertInto("tasks").
		Columns("alert_id", "assignee_id", "status", "handle_remark", "created_at").
		Record(t).Returning("id").Load(&id)
	t.ID = id
	return id, err
}

func (r *Repo) GetTask(id int64) (*model.Task, error) {
	var t model.Task
	err := r.sess().Select("*").From("tasks").Where("id = ?", id).LoadOne(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repo) ListTasksByAssignee(assigneeID int64, status string) ([]model.Task, error) {
	q := r.sess().Select("*").From("tasks").
		Where("assignee_id = ?", assigneeID).OrderDir("created_at", false)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var list []model.Task
	_, err := q.Load(&list)
	return list, err
}

// UpdateTaskStatus 带 from 校验更新任务状态。
func (r *Repo) UpdateTaskStatus(id int64, from, to string, fields map[string]interface{}) (int64, error) {
	stmt := r.sess().Update("tasks").Set("status", to).Where("id = ?", id).Where("status = ?", from)
	for k, v := range fields {
		stmt = stmt.Set(k, v)
	}
	res, err := stmt.Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// TimedOutDispatched 返回派发后超时仍未确认的任务（用于自动重派提醒）。
func (r *Repo) TimedOutDispatched(before time.Time) ([]model.Task, error) {
	var list []model.Task
	_, err := r.sess().Select("*").From("tasks").
		Where("status = ?", model.TaskPending).
		Where("created_at < ?", before).Load(&list)
	return list, err
}

func (r *Repo) InsertAttachment(a *model.Attachment) error {
	a.CreatedAt = time.Now()
	_, err := r.sess().InsertInto("attachments").
		Columns("task_id", "type", "cos_key", "url", "created_at").
		Record(a).Exec()
	return err
}

func (r *Repo) ListAttachments(taskID int64) ([]model.Attachment, error) {
	var list []model.Attachment
	_, err := r.sess().Select("*").From("attachments").
		Where("task_id = ?", taskID).OrderBy("id").Load(&list)
	return list, err
}

// ---------- 村民隐患上报 ----------

func (r *Repo) InsertReport(rp *model.Report) (int64, error) {
	rp.CreatedAt = time.Now()
	var id int64
	err := r.sess().InsertInto("reports").
		Columns("reporter_id", "grid_id", "content", "lng", "lat", "status", "created_at").
		Record(rp).Returning("id").Load(&id)
	rp.ID = id
	return id, err
}

func (r *Repo) ListReportsByGrid(gridID int64, limit int) ([]model.Report, error) {
	var list []model.Report
	_, err := r.sess().Select("*").From("reports").
		Where("grid_id = ?", gridID).
		OrderDir("created_at", false).Limit(uint64(limit)).Load(&list)
	return list, err
}
