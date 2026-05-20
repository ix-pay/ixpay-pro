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
		logger, err = utils.NewLoggerWithFile(utils.INFO, false, *logPath, "网关")
		if err != nil {
			fmt.Printf("初始化日志文件失败：%v，将输出到终端\n", err)
			logger = utils.NewLogger(utils.INFO, false, "网关")
		}
	} else {
		logger = utils.NewLogger(utils.INFO, true, "网关")
	}
	defer logger.Close()
	logger.Info("正在启动网关服务...")

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Warn("加载配置文件失败：%v，使用默认配置", err)
		cfg, err = config.LoadConfig("")
		if err != nil {
			logger.Fatal("加载默认配置失败：%v", err)
			os.Exit(1)
		}
	}

	// 初始化组件
	registry := discovery.NewRegistry()
	balancer := loadbalance.NewRoundRobinBalancer(100) // 连接数阈值设为 100
	proxy := proxy.NewProxy(registry, balancer, logger)
	handler := api.NewHandler(registry, proxy, cfg, logger)
	router := api.NewRouter(handler, proxy)
	clusterSync := cluster.NewClusterSync(registry, cfg, logger)
	healthChecker := discovery.NewHealthChecker(registry, cfg.HealthCheckInterval, cfg.HealthCheckTimeout, logger)

	// 设置路由
	router.SetupRoutes()

	// 启动服务
	healthChecker.Start()
	clusterSync.Start()
	serverAddr := fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.ListenPort)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      nil, // Uses default http.Handler
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("网关已启动，监听地址：%s", serverAddr)
	logger.Info("注册认证密钥：%s", cfg.RegisterAuthKey)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("启动服务器失败：%v", err)
	}
}
