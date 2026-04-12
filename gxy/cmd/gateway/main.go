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
	// Parse command line arguments
	configPath := flag.String("config", "./config.json", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		fmt.Printf("Using default configuration...\n")
		// 使用默认配置
		cfg, err = config.LoadConfig("")
		if err != nil {
			fmt.Printf("Failed to load default config: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize logger
	logger := utils.NewLogger(utils.INFO, true)
	logger.Info("Starting gateway service...")

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
