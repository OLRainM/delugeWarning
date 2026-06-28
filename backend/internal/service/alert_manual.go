package service

import (
	"context"
	"fmt"
	"log"

	"delugewarning/internal/model"
)

// CreateManual 网格员人工兜底发布预警。
func (s *AlertService) CreateManual(operatorID, gridID int64, level, content string) (*model.Alert, error) {
	if content == "" {
		return nil, fmt.Errorf("预警内容不能为空")
	}
	ttsKey, err := s.tts.Synthesize(context.Background(), content)
	if err != nil {
		log.Printf("[alert] 人工预警 TTS 失败: %v", err)
	}
	a := &model.Alert{
		Source: model.SourceManual, Level: level, DisasterType: "flood",
		GridID: gridID, Title: levelName(level) + "预警", Content: content,
		TTSURL: ttsKey, // 存 ObjectKey，下发时由 broadcast 接口签名
		Status: model.AlertTriggered, TriggeredBy: operatorID,
	}
	if _, err := s.repo.InsertAlert(a); err != nil {
		return nil, err
	}
	_ = s.repo.InsertAlertLog(&model.AlertLog{AlertID: a.ID, ToStatus: model.AlertTriggered,
		OperatorID: operatorID, Remark: "人工兜底发布"})
	s.Dispatch(a)
	return a, nil
}

// Archive 归档已处置的预警。
func (s *AlertService) Archive(alertID, operatorID int64) error {
	a, err := s.repo.GetAlert(alertID)
	if err != nil {
		return err
	}
	if a.Status != model.AlertHandled {
		return fmt.Errorf("仅已处置预警可归档，当前状态: %s", a.Status)
	}
	n, err := s.repo.UpdateAlertStatus(alertID, model.AlertHandled, model.AlertArchived,
		map[string]interface{}{"archived_at": nowPtr()})
	if err != nil || n == 0 {
		return fmt.Errorf("归档失败")
	}
	return s.repo.InsertAlertLog(&model.AlertLog{AlertID: alertID, FromStatus: model.AlertHandled,
		ToStatus: model.AlertArchived, OperatorID: operatorID, Remark: "归档"})
}
