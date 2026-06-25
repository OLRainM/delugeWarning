package api

import (
	"encoding/json"
	"time"
)

func jsonUnmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}

// parseDate 解析 YYYY-MM-DD，失败返回默认值。
func parseDate(s string, def time.Time) time.Time {
	if s == "" {
		return def
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return def
	}
	return t
}
