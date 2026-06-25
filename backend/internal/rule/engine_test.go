package rule

import (
	"testing"

	"delugewarning/internal/model"
)

func sampleRules() []model.Rule {
	return []model.Rule{
		{DeviceType: "water_level", Operator: ">=", Threshold: 2.5, Level: "blue", CooldownSec: 600},
		{DeviceType: "water_level", Operator: ">=", Threshold: 3.0, Level: "yellow", CooldownSec: 600},
		{DeviceType: "water_level", Operator: ">=", Threshold: 3.5, Level: "orange", CooldownSec: 300},
		{DeviceType: "water_level", Operator: ">=", Threshold: 4.0, Level: "red", CooldownSec: 180},
	}
}

func TestMatchReturnsHighestLevel(t *testing.T) {
	e := NewEngine()
	e.Load(sampleRules())

	cases := []struct {
		value float64
		want  string // "" 表示未命中
	}{
		{2.0, ""},
		{2.6, "blue"},
		{3.2, "yellow"},
		{3.7, "orange"},
		{4.5, "red"},
	}
	for _, c := range cases {
		r := e.Match("water_level", c.value)
		if c.want == "" {
			if r != nil {
				t.Errorf("value=%.1f 期望未命中，却命中 %s", c.value, r.Level)
			}
			continue
		}
		if r == nil || r.Level != c.want {
			got := "nil"
			if r != nil {
				got = r.Level
			}
			t.Errorf("value=%.1f 期望 %s，实际 %s", c.value, c.want, got)
		}
	}
}

func TestAllowFireDebounce(t *testing.T) {
	e := NewEngine()
	if !e.AllowFire("dev-1", "orange", 600) {
		t.Fatal("首次应允许触发")
	}
	if e.AllowFire("dev-1", "orange", 600) {
		t.Fatal("冷却期内不应再次触发")
	}
	if !e.AllowFire("dev-1", "red", 600) {
		t.Fatal("不同级别应独立计冷却")
	}
}

func TestLevelRank(t *testing.T) {
	if LevelRank("red") <= LevelRank("orange") {
		t.Fatal("red 应高于 orange")
	}
}
