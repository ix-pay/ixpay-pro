package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	ListenAddr          string        `json:"listenAddr"`
	ListenPort          int           `json:"listenPort"`
	RegisterTTL         time.Duration `json:"registerTTL"`
	HealthCheckInterval time.Duration `json:"healthCheckInterval"`
	HealthCheckTimeout  time.Duration `json:"healthCheckTimeout"`
	SeedNodes           []string      `json:"seedNodes"`
	HeartbeatInterval   time.Duration `json:"heartbeatInterval"`
	RegisterAuthKey     string        `json:"registerAuthKey"`
	EnableAutoDiscovery bool          `json:"enableAutoDiscovery"`
	NodeDiscoveryPort   int           `json:"nodeDiscoveryPort"`
}

func LoadConfig(filePath string) (*Config, error) {
	config := &Config{
		ListenAddr:          "0.0.0.0",
		ListenPort:          8085,
		RegisterTTL:         time.Second * 30,
		HealthCheckInterval: time.Second * 10,
		HealthCheckTimeout:  time.Second * 5,
		SeedNodes:           []string{},
		HeartbeatInterval:   time.Second * 5,
		RegisterAuthKey:     "ihvke@2025", // 使用正确的认证密钥
		EnableAutoDiscovery: true,
		NodeDiscoveryPort:   8084,
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}
