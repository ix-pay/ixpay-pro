package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"

	// 导入PostgreSQL驱动，即使我们不直接使用它
	_ "github.com/lib/pq"
)

// @title 微信支付API服务
// @version 1.0
// @description 基于Gin和PostgreSQL的微信支付API服务
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 127.0.0.1:8586
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header [Bearer ]
// @name Authorization
func main() {

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		// 如果无法加载配置，使用标准库log.Fatal
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志记录器
	appLoggerManager := logger.SetupMultiLogger(cfg)
	logger.SetGlobalMultiLogger(appLoggerManager)
	appLogger := appLoggerManager.GetLogger(logger.DefaultLogger)

	// 初始化应用程序
	application, err := app.InitializeApp()
	if err != nil {
		appLogger.Fatal("初始化应用程序失败", "error", err)
	}

	// 启动应用程序
	go func() {
		if err := application.Start(); err != nil {
			appLogger.Fatal("启动应用程序失败", "error", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 创建一个超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭应用程序
	if err := application.Shutdown(ctx); err != nil {
		appLogger.Fatal("关闭应用程序失败", "error", err)
	}

	appLogger.Info("应用程序已成功关闭")
}
