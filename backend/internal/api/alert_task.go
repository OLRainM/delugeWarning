package api

import (
	"net/http"
	"strconv"
	"time"

	"delugewarning/internal/middleware"
	"delugewarning/internal/model"
	"delugewarning/internal/provider"

	"github.com/gin-gonic/gin"
)

// ---------- 预警（网格员复核为主） ----------

func (h *Handler) listAlerts(c *gin.Context) {
	list, err := h.repo.ListAlerts(c.Query("status"), 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *Handler) getAlert(c *gin.Context) {
	id := parseID(c.Param("id"))
	a, err := h.repo.GetAlert(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预警不存在"})
		return
	}
	logs, _ := h.repo.ListAlertLogs(id)
	c.JSON(http.StatusOK, gin.H{"alert": a, "logs": logs})
}

type reviewReq struct {
	Action  string `json:"action"` // confirm / modify / cancel
	Content string `json:"content"`
}

func (h *Handler) reviewAlert(c *gin.Context) {
	id := parseID(c.Param("id"))
	var req reviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.alert.Review(id, middleware.UserIDOf(c), req.Action, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type manualAlertReq struct {
	GridID  int64  `json:"grid_id"`
	Level   string `json:"level"`
	Content string `json:"content"`
}

func (h *Handler) createManualAlert(c *gin.Context) {
	var req manualAlertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	a, err := h.alert.CreateManual(middleware.UserIDOf(c), req.GridID, req.Level, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *Handler) archiveAlert(c *gin.Context) {
	id := parseID(c.Param("id"))
	if err := h.alert.Archive(id, middleware.UserIDOf(c)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- 网格员任务 ----------

func (h *Handler) listTasks(c *gin.Context) {
	list, err := h.task.ListMine(middleware.UserIDOf(c), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *Handler) confirmTask(c *gin.Context) {
	if err := h.task.Confirm(parseID(c.Param("id")), middleware.UserIDOf(c)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type handleReq struct {
	Remark      string             `json:"remark"`
	Attachments []model.Attachment `json:"attachments"`
}

func (h *Handler) handleTask(c *gin.Context) {
	var req handleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.task.Handle(parseID(c.Param("id")), middleware.UserIDOf(c), req.Remark, req.Attachments); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- COS 预签名直传 ----------

type presignReq struct {
	Key string `json:"key"`
}

func (h *Handler) presignUpload(c *gin.Context) {
	var req presignReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 key"})
		return
	}
	uploadURL, accessURL, err := h.storage.PresignPut(c, req.Key, 10*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"upload_url": uploadURL, "access_url": accessURL})
}

func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

var _ = provider.TemplateMsg{}
