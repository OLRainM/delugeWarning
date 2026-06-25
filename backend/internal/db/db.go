package db

import (
	"time"

	"delugewarning/internal/config"

	"github.com/gocraft/dbr/v2"
	_ "github.com/lib/pq"
)

// New 用 dbr 打开 PostgreSQL 连接并返回会话。
func New(cfg config.DatabaseConfig) (*dbr.Connection, error) {
	conn, err := dbr.Open("postgres", cfg.DSN, nil)
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	conn.SetConnMaxLifetime(time.Hour)
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return conn, nil
}
