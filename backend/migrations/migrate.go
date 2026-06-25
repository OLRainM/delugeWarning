package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed *.up.sql
var upFiles embed.FS

// Up 运行所有尚未应用的 up 迁移，按文件名顺序执行，幂等。
func Up(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version TEXT PRIMARY KEY,
        applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("创建迁移表失败: %w", err)
	}

	entries, err := upFiles.ReadDir(".")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, ".up.sql")
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		content, err := upFiles.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("应用迁移 %s 失败: %w", name, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations(version) VALUES($1)`, version); err != nil {
			return err
		}
		fmt.Printf("[migrate] 已应用 %s\n", name)
	}
	return nil
}
