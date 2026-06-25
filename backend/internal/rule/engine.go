package rule

import (
	"sort"
	"sync"
	"time"

	"delugewarning/internal/model"
)

// levelRank 用于比较预警级别高低。
var levelRank = map[string]int{"blue": 1, "yellow": 2, "orange": 3, "red": 4}

// Engine 是常驻内存的规则引擎，负责阈值匹配、防抖、升级判定。
type Engine struct {
	mu       sync.RWMutex
	rules    []model.Rule          // 按 threshold 降序，便于取命中的最高级
	cooldown map[string]time.Time  // key: deviceID|level -> 上次触发时间
}

func NewEngine() *Engine {
	return &Engine{cooldown: make(map[string]time.Time)}
}

// Load 用最新规则刷新内存（启动时与规则变更时调用）。
func (e *Engine) Load(rules []model.Rule) {
	sorted := make([]model.Rule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Threshold > sorted[j].Threshold })
	e.mu.Lock()
	e.rules = sorted
	e.mu.Unlock()
}

// Match 对一条读数评估，返回命中的最高级规则（未命中返回 nil）。
func (e *Engine) Match(deviceType string, value float64) *model.Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for i := range e.rules {
		r := e.rules[i]
		if r.DeviceType != deviceType {
			continue
		}
		if compare(value, r.Operator, r.Threshold) {
			return &r // rules 已按阈值降序，第一个命中即最高级
		}
	}
	return nil
}

// AllowFire 防抖判定：冷却期内同设备同级别不重复触发。命中返回 true 并记录时间。
func (e *Engine) AllowFire(deviceID, level string, cooldownSec int) bool {
	key := deviceID + "|" + level
	now := time.Now()
	e.mu.Lock()
	defer e.mu.Unlock()
	if last, ok := e.cooldown[key]; ok {
		if now.Sub(last) < time.Duration(cooldownSec)*time.Second {
			return false
		}
	}
	e.cooldown[key] = now
	return true
}

// LevelRank 暴露级别排序，供升级判定使用。
func LevelRank(level string) int { return levelRank[level] }

func compare(v float64, op string, threshold float64) bool {
	switch op {
	case ">":
		return v > threshold
	case ">=":
		return v >= threshold
	case "<":
		return v < threshold
	case "<=":
		return v <= threshold
	case "==":
		return v == threshold
	default:
		return v >= threshold
	}
}
