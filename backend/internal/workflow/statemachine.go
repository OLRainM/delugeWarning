package workflow

import "delugewarning/internal/model"

// transitions 定义合法状态流转表（配置化状态机，替代重量级 BPMN）。
var transitions = map[string][]string{
	model.AlertPendingReview: {model.AlertTriggered, model.AlertCanceled},
	model.AlertTriggered:     {model.AlertDispatched, model.AlertCanceled},
	model.AlertDispatched:    {model.AlertConfirmed, model.AlertDispatched, model.AlertCanceled},
	model.AlertConfirmed:     {model.AlertHandled},
	model.AlertHandled:       {model.AlertArchived},
}

// CanTransition 判断从 from 到 to 是否为合法流转。
func CanTransition(from, to string) bool {
	for _, t := range transitions[from] {
		if t == to {
			return true
		}
	}
	return false
}
