package repository

import (
	"time"

	"delugewarning/internal/model"

	"github.com/gocraft/dbr/v2"
)

// ---------- 用户 ----------

// UpsertUserByOpenID 按 openid 查找用户，不存在则创建（默认村民）。
func (r *Repo) UpsertUserByOpenID(openid, role string) (*model.User, error) {
	sess := r.sess()
	var u model.User
	err := sess.Select("*").From("users").Where("openid = ?", openid).LoadOne(&u)
	if err == nil {
		return &u, nil
	}
	if err != dbr.ErrNotFound {
		return nil, err
	}
	if role == "" {
		role = model.RoleVillager
	}
	u = model.User{OpenID: openid, Role: role, CreatedAt: time.Now()}
	err = sess.InsertInto("users").
		Columns("openid", "role", "name", "phone", "grid_id", "created_at").
		Record(&u).Returning("id").Load(&u.ID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetUser(id int64) (*model.User, error) {
	var u model.User
	err := r.sess().Select("*").From("users").Where("id = ?", id).LoadOne(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GridworkerByGrid 返回某网格的网格员（用于任务派发）。
func (r *Repo) GridworkerByGrid(gridID int64) (*model.User, error) {
	var u model.User
	err := r.sess().Select("*").From("users").
		Where("role = ?", model.RoleGridworker).
		Where("grid_id = ?", gridID).
		Limit(1).LoadOne(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---------- 预警 ----------

func (r *Repo) InsertAlert(a *model.Alert) (int64, error) {
	a.CreatedAt = time.Now()
	var id int64
	err := r.sess().InsertInto("alerts").
		Columns("source", "level", "disaster_type", "grid_id", "device_id",
			"title", "content", "tts_url", "status", "triggered_by", "reviewed_by", "created_at").
		Record(a).Returning("id").Load(&id)
	a.ID = id
	return id, err
}

func (r *Repo) GetAlert(id int64) (*model.Alert, error) {
	var a model.Alert
	err := r.sess().Select("*").From("alerts").Where("id = ?", id).LoadOne(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repo) ListAlerts(status string, limit int) ([]model.Alert, error) {
	q := r.sess().Select("*").From("alerts").OrderDir("created_at", false)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if limit > 0 {
		q = q.Limit(uint64(limit))
	}
	var list []model.Alert
	_, err := q.Load(&list)
	return list, err
}

// UpdateAlertStatus 带 from 校验的状态更新（幂等并发安全），返回受影响行数。
func (r *Repo) UpdateAlertStatus(id int64, from, to string, fields map[string]interface{}) (int64, error) {
	stmt := r.sess().Update("alerts").Set("status", to).Where("id = ?", id).Where("status = ?", from)
	for k, v := range fields {
		stmt = stmt.Set(k, v)
	}
	res, err := stmt.Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repo) InsertAlertLog(l *model.AlertLog) error {
	l.CreatedAt = time.Now()
	_, err := r.sess().InsertInto("alert_logs").
		Columns("alert_id", "from_status", "to_status", "operator_id", "remark", "created_at").
		Record(l).Exec()
	return err
}

func (r *Repo) ListAlertLogs(alertID int64) ([]model.AlertLog, error) {
	var list []model.AlertLog
	_, err := r.sess().Select("*").From("alert_logs").
		Where("alert_id = ?", alertID).OrderBy("created_at").Load(&list)
	return list, err
}

// ActiveAlertsByGrid 返回某网格生效中的预警（供村民端查看）。
func (r *Repo) ActiveAlertsByGrid(gridID int64, limit int) ([]model.Alert, error) {
	var list []model.Alert
	_, err := r.sess().Select("*").From("alerts").
		Where("grid_id = ?", gridID).
		Where("status NOT IN ?", []string{model.AlertArchived, model.AlertCanceled}).
		OrderDir("created_at", false).Limit(uint64(limit)).Load(&list)
	return list, err
}
