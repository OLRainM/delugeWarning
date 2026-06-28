package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 是后端全局配置，来源于 YAML 文件 + 环境变量覆盖。
type Config struct {
	Server                ServerConfig   `yaml:"server"`
	Database              DatabaseConfig `yaml:"database"`
	JWT                   JWTConfig      `yaml:"jwt"`
	WeChat                WeChatConfig   `yaml:"wechat"`
	ReadingsRetentionDays int            `yaml:"readings_retention_days"`
	AsyncWorkers          int            `yaml:"async_workers"`
	DeviceSecret          string         `yaml:"device_secret"`
	TTS                   TTSConfig      `yaml:"tts"`
	Storage               StorageConfig  `yaml:"storage"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type WeChatConfig struct {
	AppID  string `yaml:"appid"`
	Secret string `yaml:"secret"`
}

type TTSConfig struct {
	Provider  string `yaml:"provider"`
	Voice     string `yaml:"voice"`
	PythonBin string `yaml:"python_bin"`
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
}

type StorageConfig struct {
	Provider  string `yaml:"provider"`
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	BaseURL   string `yaml:"base_url"`
}

// Load 读取 YAML 配置文件并应用环境变量覆盖。
func Load(path string) (*Config, error) {
	cfg := defaultConfig()
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	applyEnvOverrides(cfg)
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server:                ServerConfig{Addr: ":8080", Mode: "debug"},
		Database:              DatabaseConfig{MaxOpenConns: 20, MaxIdleConns: 5},
		JWT:                   JWTConfig{Secret: "change-me-in-prod", ExpireHours: 168},
		ReadingsRetentionDays: 30,
		AsyncWorkers:          4,
		TTS:                   TTSConfig{Provider: "mock", Region: "ap-guangzhou"},
		Storage:               StorageConfig{Provider: "mock", Region: "ap-guangzhou"},
	}
}

// applyEnvOverrides 用环境变量覆盖敏感/部署相关配置。
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SERVER_ADDR"); v != "" {
		cfg.Server.Addr = v
	}
	if v := os.Getenv("PG_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("WX_APPID"); v != "" {
		cfg.WeChat.AppID = v
	}
	if v := os.Getenv("WX_SECRET"); v != "" {
		cfg.WeChat.Secret = v
	}
	if v := os.Getenv("DEVICE_SECRET"); v != "" {
		cfg.DeviceSecret = v
	}
	if v := os.Getenv("ASYNC_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.AsyncWorkers = n
		}
	}
	if v := os.Getenv("READINGS_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ReadingsRetentionDays = n
		}
	}
	if v := os.Getenv("COS_SECRET_ID"); v != "" {
		cfg.Storage.SecretID = v
	}
	if v := os.Getenv("COS_SECRET_KEY"); v != "" {
		cfg.Storage.SecretKey = v
	}
}
