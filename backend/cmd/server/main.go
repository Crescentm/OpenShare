package main

import (
	"fmt"
	"os"

	"github.com/openshare/backend/internal/config"
	"github.com/openshare/backend/internal/router"
	"github.com/openshare/backend/pkg/logger"
	"github.com/openshare/backend/pkg/storage"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// 初始化存储目录
	if err := storage.InitDirectories(cfg.Storage.BasePath); err != nil {
		log.Fatal("Failed to init storage directories", "error", err)
	}
	log.Info("Storage directories initialized", "path", cfg.Storage.BasePath)

	// 初始化数据库连接
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect database", "error", err)
	}
	log.Info("Database connected successfully")

	// 初始化路由
	r := router.Setup(cfg, db, log)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info("Server starting", "port", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server", "error", err)
	}
}
