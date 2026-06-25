package repository

import (
	"time"

	"delugewarning/internal/model"

	"github.com/gocraft/dbr/v2"
)

// Repo 聚合所有数据访问，基于 dbr 会话。
type Repo struct {
	conn *dbr.Connection
}

func New(conn *dbr.Connection) *Repo {
	return &Repo{conn: conn}
}

func (r *Repo) sess() *dbr.Session {
	return r.conn.NewSession(nil)
}

// ---------- 设备与读数 ----------

// InsertReading 插入一条读数并更新设备最新值。
func (r *Repo) InsertReading(rd *model.Reading) (int64, error) {
	sess := r.sess()
	var id int64
	err := sess.InsertInto("readings").
		Columns("device_id", "value", "unit", "reported_at").
		Record(rd).
		Returning("id").
		Load(&id)
	if err != nil {
		return 0, err
	}
	_, _ = sess.Update("devices").
		Set("last_value", rd.Value).
		Set("last_report_at", rd.ReportedAt).
		Where("id = ?", rd.DeviceID).
		Exec()
	return id, nil
}

func (r *Repo) GetDevice(id string) (*model.Device, error) {
	var d model.Device
	err := r.sess().Select("*").From("devices").Where("id = ?", id).LoadOne(&d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repo) ListDevices() ([]model.Device, error) {
	var list []model.Device
	_, err := r.sess().Select("*").From("devices").OrderBy("id").Load(&list)
	return list, err
}

// LatestReading 取设备最新一条读数。
func (r *Repo) LatestReading(deviceID string) (*model.Reading, error) {
	var rd model.Reading
	err := r.sess().Select("*").From("readings").
		Where("device_id = ?", deviceID).
		OrderDir("reported_at", false).
		Limit(1).LoadOne(&rd)
	if err != nil {
		return nil, err
	}
	return &rd, nil
}

// ---------- 规则与模板 ----------

func (r *Repo) ListEnabledRules() ([]model.Rule, error) {
	var list []model.Rule
	_, err := r.sess().Select("*").From("rules").
		Where("enabled = ?", true).
		OrderDir("threshold", false).Load(&list)
	return list, err
}

func (r *Repo) GetTemplate(id int64) (*model.Template, error) {
	var t model.Template
	err := r.sess().Select("*").From("templates").Where("id = ?", id).LoadOne(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ---------- 日聚合趋势 ----------

// TrendStats 查询某设备某时间段的日聚合数据。
func (r *Repo) TrendStats(deviceID, metric string, from, to time.Time) ([]model.DailyStat, error) {
	var list []model.DailyStat
	_, err := r.sess().Select("*").From("daily_water_stats").
		Where("device_id = ?", deviceID).
		Where("metric = ?", metric).
		Where("stat_date >= ?", from).
		Where("stat_date <= ?", to).
		OrderBy("stat_date").Load(&list)
	return list, err
}
