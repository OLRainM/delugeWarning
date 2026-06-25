package workflow

import (
	"testing"

	"delugewarning/internal/model"
)

func TestValidTransitions(t *testing.T) {
	ok := [][2]string{
		{model.AlertPendingReview, model.AlertTriggered},
		{model.AlertPendingReview, model.AlertCanceled},
		{model.AlertTriggered, model.AlertDispatched},
		{model.AlertDispatched, model.AlertConfirmed},
		{model.AlertConfirmed, model.AlertHandled},
		{model.AlertHandled, model.AlertArchived},
	}
	for _, c := range ok {
		if !CanTransition(c[0], c[1]) {
			t.Errorf("应允许流转 %s -> %s", c[0], c[1])
		}
	}
}

func TestInvalidTransitions(t *testing.T) {
	bad := [][2]string{
		{model.AlertTriggered, model.AlertArchived},
		{model.AlertConfirmed, model.AlertDispatched},
		{model.AlertArchived, model.AlertTriggered},
		{model.AlertCanceled, model.AlertDispatched},
	}
	for _, c := range bad {
		if CanTransition(c[0], c[1]) {
			t.Errorf("不应允许流转 %s -> %s", c[0], c[1])
		}
	}
}
