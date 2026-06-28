package api

import (
	"net/http"

	"delugewarning/internal/middleware"
	"delugewarning/internal/model"

	"github.com/gin-gonic/gin"
)

// ---------- 村民端 ----------

func (h *Handler) profile(c *gin.Context) {
	u, err := h.repo.GetUser(middleware.UserIDOf(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// updateName 更新当前用户姓名（网格员"我的"页填写）。
func (h *Handler) updateName(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name 不能为空"})
		return
	}
	if err := h.repo.UpdateUserName(middleware.UserIDOf(c), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// villageAlerts 村民查看本网格生效中的预警。
func (h *Handler) villageAlerts(c *gin.Context) {
	u, err := h.repo.GetUser(middleware.UserIDOf(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	list, err := h.repo.ActiveAlertsByGrid(u.GridID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// broadcast 获取预警广播文本与 TTS 音频地址。
func (h *Handler) broadcast(c *gin.Context) {
	a, err := h.repo.GetAlert(parseID(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预警不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"title": a.Title, "content": a.Content, "tts_url": a.TTSURL})
}

type reportReq struct {
	Content string  `json:"content"`
	Lng     float64 `json:"lng"`
	Lat     float64 `json:"lat"`
}

// submitReport 村民提交隐患上报。
func (h *Handler) submitReport(c *gin.Context) {
	var req reportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	u, err := h.repo.GetUser(middleware.UserIDOf(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	rp := &model.Report{ReporterID: u.ID, GridID: u.GridID, Content: req.Content,
		Lng: req.Lng, Lat: req.Lat, Status: "open"}
	if _, err := h.repo.InsertReport(rp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rp)
}

// guides 避险指引（静态内容，按灾种返回）。
func (h *Handler) guides(c *gin.Context) {
	dt := c.DefaultQuery("disaster_type", "flood")
	guides := map[string][]string{
		"flood": {
			"立即向高处转移，避免靠近河道、低洼地带。",
			"切断低洼处电源，防止触电。",
			"听从网格员安排，前往就近避难点。",
		},
	}
	c.JSON(http.StatusOK, gin.H{"disaster_type": dt, "tips": guides[dt]})
}
