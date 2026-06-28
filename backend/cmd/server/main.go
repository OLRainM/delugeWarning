package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"delugewarning/internal/api"
	"delugewarning/internal/async"
	"delugewarning/internal/auth"
	"delugewarning/internal/config"
	"delugewarning/internal/cron"
	"delugewarning/internal/db"
	"delugewarning/internal/provider"
	"delugewarning/internal/repository"
	"delugewarning/internal/rule"
	"delugewarning/internal/service"
	migrate "delugewarning/migrations"

	"github.com/gin-gonic/gin"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	conn, err := db.New(cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer conn.Close()

	// 自动迁移
	if err := migrate.Up(conn.DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	repo := repository.New(conn)

	// 规则引擎：启动加载常驻内存
	engine := rule.NewEngine()
	if rules, err := repo.ListEnabledRules(); err == nil {
		engine.Load(rules)
		log.Printf("[rule] 已加载 %d 条规则", len(rules))
	}

	// 异步任务队列
	queue := async.New(cfg.AsyncWorkers, 2048)
	defer queue.Close()

	// 外部能力（mock 便于本地联调，可切腾讯云）
	tts := buildTTS(cfg)
	storage := buildStorage(cfg)
	pusher := buildPusher(cfg)

	// 服务装配
	alertSvc := service.NewAlertService(repo, engine, queue, tts, pusher)
	deviceSvc := service.NewDeviceService(repo, alertSvc)
	taskSvc := service.NewTaskService(repo)
	jwtMgr := auth.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	// 定时任务
	sched := cron.New(repo, alertSvc, cfg.ReadingsRetentionDays)
	sched.Start()
	defer sched.Stop()

	// HTTP
	gin.SetMode(ginMode(cfg.Server.Mode))
	r := gin.Default()
	h := api.NewHandler(cfg, repo, jwtMgr, deviceSvc, alertSvc, taskSvc, storage)
	h.Register(r)

	go func() {
		log.Printf("[server] 监听 %s", cfg.Server.Addr)
		if err := r.Run(cfg.Server.Addr); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[server] 正在关闭...")
}

func buildTTS(cfg *config.Config) provider.TTSEngine {
	// TODO: provider == "tencent" 时返回腾讯云实现
	return provider.MockTTS{}
}

func buildStorage(cfg *config.Config) provider.Storage {
	if cfg.Storage.Provider == "cos" {
		s, err := provider.NewCOSStorage(cfg.Storage.SecretID, cfg.Storage.SecretKey, cfg.Storage.BaseURL)
		if err != nil {
			log.Fatalf("初始化 COS 存储失败: %v", err)
		}
		log.Printf("[storage] 使用腾讯云 COS，桶=%s，域名=%s", cfg.Storage.Bucket, cfg.Storage.BaseURL)
		return s
	}
	log.Println("[storage] 使用 mock 存储（本地联调）")
	return provider.MockStorage{}
}

func buildPusher(cfg *config.Config) provider.Pusher {
	return provider.MockPusher{}
}

func ginMode(mode string) string {
	if mode == "release" {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}
