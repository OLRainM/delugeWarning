package model

import "time"

// 角色常量：系统仅两个角色，网格员兼任后台管理与复核。
const (
	RoleVillager   = "villager"   // 村民
	RoleGridworker = "gridworker" // 网格员（兼后台管理员）
)

// 预警状态机枚举。
const (
	AlertPendingReview = "pending_review" // 待人工复核（低级别可选）
	AlertTriggered     = "triggered"      // 已触发
	AlertDispatched    = "dispatched"     // 已派发
	AlertConfirmed     = "confirmed"      // 网格员已确认
	AlertHandled       = "handled"        // 已处置
	AlertArchived      = "archived"       // 已归档
	AlertCanceled      = "canceled"       // 已撤销（误报）
)

// 预警来源。
const (
	SourceSensor = "sensor" // 系统自动（传感器触发）
	SourceManual = "manual" // 人工兜底
)

// 任务状态。
const (
	TaskPending  = "pending"  // 待确认
	TaskHandling = "handling" // 处置中
	TaskFinished = "finished" // 已完成
)

type User struct {
	ID        int64     `db:"id" json:"id"`
	OpenID    string    `db:"openid" json:"openid"`
	Role      string    `db:"role" json:"role"`
	Name      string    `db:"name" json:"name"`
	Phone     string    `db:"phone" json:"phone"`
	GridID    int64     `db:"grid_id" json:"grid_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Grid struct {
	ID            int64  `db:"id" json:"id"`
	Name          string `db:"name" json:"name"`
	Village       string `db:"village" json:"village"`
	ManagerUserID int64  `db:"manager_user_id" json:"manager_user_id"`
}

type Device struct {
	ID           string     `db:"id" json:"id"`
	Type         string     `db:"type" json:"type"` // water_level / rainfall
	Name         string     `db:"name" json:"name"`
	GridID       int64      `db:"grid_id" json:"grid_id"`
	Status       string     `db:"status" json:"status"`
	LastValue    float64    `db:"last_value" json:"last_value"`
	LastReportAt *time.Time `db:"last_report_at" json:"last_report_at"`
}

type Reading struct {
	ID         int64     `db:"id" json:"id"`
	DeviceID   string    `db:"device_id" json:"device_id"`
	Value      float64   `db:"value" json:"value"`
	Unit       string    `db:"unit" json:"unit"`
	ReportedAt time.Time `db:"reported_at" json:"reported_at"`
}

type Rule struct {
	ID             int64   `db:"id" json:"id"`
	DeviceType     string  `db:"device_type" json:"device_type"`
	Metric         string  `db:"metric" json:"metric"`
	Operator       string  `db:"operator" json:"operator"` // ">" ">=" 等
	Threshold      float64 `db:"threshold" json:"threshold"`
	Level          string  `db:"level" json:"level"` // blue/yellow/orange/red
	CooldownSec    int     `db:"cooldown_sec" json:"cooldown_sec"`
	TemplateID     int64   `db:"template_id" json:"template_id"`
	ReviewRequired bool    `db:"review_required" json:"review_required"`
	Enabled        bool    `db:"enabled" json:"enabled"`
}

type Template struct {
	ID           int64  `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	DisasterType string `db:"disaster_type" json:"disaster_type"`
	ContentTpl   string `db:"content_tpl" json:"content_tpl"`
	Enabled      bool   `db:"enabled" json:"enabled"`
}

type Alert struct {
	ID           int64      `db:"id" json:"id"`
	Source       string     `db:"source" json:"source"`
	Level        string     `db:"level" json:"level"`
	DisasterType string     `db:"disaster_type" json:"disaster_type"`
	GridID       int64      `db:"grid_id" json:"grid_id"`
	DeviceID     string     `db:"device_id" json:"device_id"`
	Title        string     `db:"title" json:"title"`
	Content      string     `db:"content" json:"content"`
	TTSURL       string     `db:"tts_url" json:"tts_url"`
	Status       string     `db:"status" json:"status"`
	TriggeredBy  int64      `db:"triggered_by" json:"triggered_by"`
	ReviewedBy   int64      `db:"reviewed_by" json:"reviewed_by"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	ArchivedAt   *time.Time `db:"archived_at" json:"archived_at"`
}

type AlertLog struct {
	ID         int64     `db:"id" json:"id"`
	AlertID    int64     `db:"alert_id" json:"alert_id"`
	FromStatus string    `db:"from_status" json:"from_status"`
	ToStatus   string    `db:"to_status" json:"to_status"`
	OperatorID int64     `db:"operator_id" json:"operator_id"`
	Remark     string    `db:"remark" json:"remark"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Task struct {
	ID           int64      `db:"id" json:"id"`
	AlertID      int64      `db:"alert_id" json:"alert_id"`
	AssigneeID   int64      `db:"assignee_id" json:"assignee_id"`
	Status       string     `db:"status" json:"status"`
	HandleRemark string     `db:"handle_remark" json:"handle_remark"`
	FinishedAt   *time.Time `db:"finished_at" json:"finished_at"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type Attachment struct {
	ID        int64     `db:"id" json:"id"`
	TaskID    int64     `db:"task_id" json:"task_id"`
	Type      string    `db:"type" json:"type"` // image / video
	CosKey    string    `db:"cos_key" json:"cos_key"`
	URL       string    `db:"url" json:"url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Report struct {
	ID         int64     `db:"id" json:"id"`
	ReporterID int64     `db:"reporter_id" json:"reporter_id"`
	GridID     int64     `db:"grid_id" json:"grid_id"`
	Content    string    `db:"content" json:"content"`
	Lng        float64   `db:"lng" json:"lng"`
	Lat        float64   `db:"lat" json:"lat"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// DailyStat 对应 daily_water_stats 日聚合表。
type DailyStat struct {
	DeviceID  string    `db:"device_id" json:"device_id"`
	Metric    string    `db:"metric" json:"metric"`
	StatDate  time.Time `db:"stat_date" json:"stat_date"`
	MaxValue  float64   `db:"max_value" json:"max_value"`
	AvgValue  float64   `db:"avg_value" json:"avg_value"`
	MinValue  float64   `db:"min_value" json:"min_value"`
	SampleCnt int       `db:"sample_cnt" json:"sample_cnt"`
	Unit      string    `db:"unit" json:"unit"`
}
