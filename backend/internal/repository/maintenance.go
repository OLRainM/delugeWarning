package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// AggregateDaily 将"昨天及更早未聚合"的读数按本地时区聚合写入 daily_water_stats（幂等 UPSERT）。
func (r *Repo) AggregateDaily() error {
	const q = `
INSERT INTO daily_water_stats
    (device_id, metric, stat_date, max_value, avg_value, min_value, sample_cnt, unit)
SELECT r.device_id, d.type AS metric,
       (r.reported_at AT TIME ZONE 'Asia/Shanghai')::date AS stat_date,
       MAX(r.value), AVG(r.value), MIN(r.value), COUNT(*), MAX(r.unit)
FROM readings r JOIN devices d ON d.id = r.device_id
WHERE (r.reported_at AT TIME ZONE 'Asia/Shanghai')::date
        < (now() AT TIME ZONE 'Asia/Shanghai')::date
GROUP BY r.device_id, d.type, stat_date
ON CONFLICT (device_id, metric, stat_date) DO UPDATE
SET max_value = EXCLUDED.max_value, avg_value = EXCLUDED.avg_value,
    min_value = EXCLUDED.min_value, sample_cnt = EXCLUDED.sample_cnt;`
	_, err := r.sess().Exec(q)
	return err
}

// EnsureDayPartition 为指定日期创建按天分区（已存在则忽略）。
func (r *Repo) EnsureDayPartition(day time.Time) error {
	start := day.Format("2006-01-02")
	end := day.AddDate(0, 0, 1).Format("2006-01-02")
	name := "readings_" + day.Format("20060102")
	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s PARTITION OF readings
        FOR VALUES FROM ('%s') TO ('%s')`, name, start, end)
	_, err := r.sess().Exec(q)
	return err
}

// DropOldPartitions 删除早于保留期的按天分区（先聚合后清理由 cron 顺序保证）。
func (r *Repo) DropOldPartitions(retentionDays int) ([]string, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	rows, err := r.sess().Query(`
        SELECT inhrelid::regclass::text AS part
        FROM pg_inherits
        WHERE inhparent = 'readings'::regclass`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dropped []string
	var parts []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	for _, p := range parts {
		var dayStr string
		if _, err := fmt.Sscanf(p, "readings_%8s", &dayStr); err != nil {
			continue // 跳过 default 等非按天分区
		}
		day, err := time.Parse("20060102", dayStr)
		if err != nil {
			continue
		}
		if day.Before(cutoff) {
			if _, err := r.sess().Exec("DROP TABLE IF EXISTS " + p); err != nil {
				return dropped, err
			}
			dropped = append(dropped, p)
		}
	}
	return dropped, nil
}

// DB 暴露底层 *sql.DB（迁移等场景用）。
func (r *Repo) DB() *sql.DB {
	return r.conn.DB
}
