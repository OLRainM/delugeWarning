package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"delugewarning/internal/auth"
	"delugewarning/internal/config"
	"delugewarning/internal/model"
	"delugewarning/internal/provider"
	"delugewarning/internal/repository"
	"delugewarning/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler 持有所有依赖，挂载到 Gin 路由。
type Handler struct {
	cfg     *config.Config
	repo    *repository.Repo
	jwt     *auth.Manager
	device  *service.DeviceService
	alert   *service.AlertService
	task    *service.TaskService
	storage provider.Storage
}

func NewHandler(cfg *config.Config, repo *repository.Repo, jwt *auth.Manager,
	dev *service.DeviceService, alert *service.AlertService, task *service.TaskService,
	storage provider.Storage) *Handler {
	return &Handler{cfg: cfg, repo: repo, jwt: jwt, device: dev, alert: alert, task: task, storage: storage}
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
}

// ---------- 鉴权 ----------

type wxLoginReq struct {
	Code string `json:"code"`
	Role string `json:"role"` // 仅首次注册时生效：villager / gridworker
}

// wxLogin 用 code 换取用户身份并签发 JWT。
// 说明：本地联调用 code 直接当作 openid；接入微信后替换为 code2session。
func (h *Handler) wxLogin(c *gin.Context) {
	var req wxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 code"})
		return
	}
	openid := h.resolveOpenID(req.Code)
	user, err := h.repo.UpsertUserByOpenID(openid, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}
	token, err := h.jwt.Generate(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "签发 token 失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// resolveOpenID 本地模式直接用 code，生产应调用微信 code2session。
func (h *Handler) resolveOpenID(code string) string {
	return "openid-" + code
}

// ---------- 设备/读数 ----------

type readingReq struct {
	DeviceID   string    `json:"device_id"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	ReportedAt time.Time `json:"reported_at"`
}

// ingestReading 设备上报读数（同步落库 + 规则入队），可选 HMAC 签名校验。
func (h *Handler) ingestReading(c *gin.Context) {
	deviceID := c.Param("id")
	body, _ := io.ReadAll(c.Request.Body)
	if !h.verifySignature(c, body) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "签名校验失败"})
		return
	}
	var req readingReq
	if err := jsonUnmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体解析失败"})
		return
	}
	if err := h.device.Ingest(deviceID, req.Value, req.Unit, req.ReportedAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) verifySignature(c *gin.Context, body []byte) bool {
	if h.cfg.DeviceSecret == "" {
		return true // 未配置密钥则跳过（便于本地联调）
	}
	sig := c.GetHeader("X-Signature")
	mac := hmac.New(sha256.New, []byte(h.cfg.DeviceSecret))
	mac.Write(body)
	return hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil))))
}

func (h *Handler) listDevices(c *gin.Context) {
	list, err := h.device.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *Handler) latestReading(c *gin.Context) {
	rd, err := h.device.Latest(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "暂无读数"})
		return
	}
	c.JSON(http.StatusOK, rd)
}

func (h *Handler) deviceTrend(c *gin.Context) {
	from := parseDate(c.Query("from"), time.Now().AddDate(0, 0, -30))
	to := parseDate(c.Query("to"), time.Now())
	list, err := h.device.Trend(c.Param("id"), c.Query("metric"), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

var _ = model.RoleVillager
