package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ixpay-pro/gxy/internal/api"
	"github.com/ixpay-pro/gxy/internal/cluster"
	"github.com/ixpay-pro/gxy/internal/discovery"
	"github.com/ixpay-pro/gxy/internal/loadbalance"
	"github.com/ixpay-pro/gxy/internal/proxy"
	"github.com/ixpay-pro/gxy/pkg/config"
	"github.com/ixpay-pro/gxy/pkg/utils"
)

func main() {
	// 初始化日志
	logPath := flag.String("log", "", "日志文件路径（留空则输出到终端）")
	configPath := flag.String("config", "./config.json", "配置文件路径")
	flag.Parse()

	var logger *utils.Logger
	if *logPath != "" {
		var err error
		logger, err = utils.NewLoggerWithFile(utils.INFO, false, *logPath)
		if err != nil {
			fmt.Printf("初始化日志文件失败: %v，将输出到终端\n", err)
			logger = utils.NewLogger(utils.INFO, false)
		}
	} else {
		logger = utils.NewLogger(utils.INFO, true)
	}
	defer logger.Close()
	logger.Info("Starting gateway service...")

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Warn("加载配置文件失败: %v，使用默认配置", err)
		cfg, err = config.LoadConfig("")
		if err != nil {
			logger.Fatal("加载默认配置失败: %v", err)
			os.Exit(1)
		}
	}

	// Initialize components
	registry := discovery.NewRegistry()
	balancer := loadbalance.NewRoundRobinBalancer(100) // 连接数阈值设为100
	proxy := proxy.NewProxy(registry, balancer, logger)
	handler := api.NewHandler(registry, proxy, cfg, logger)
	router := api.NewRouter(handler, proxy)
	clusterSync := cluster.NewClusterSync(registry, cfg, logger)
	healthChecker := discovery.NewHealthChecker(registry, cfg.HealthCheckInterval, cfg.HealthCheckTimeout, logger)

	// Setup routes
	router.SetupRoutes()

	// Start services
	healthChecker.Start()
	clusterSync.Start()

	// Start HTTP server
	serverAddr := fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.ListenPort)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      nil, // Uses default http.Handler
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Gateway started and listening on %s", serverAddr)
	logger.Info("Register auth key: %s", cfg.RegisterAuthKey)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
