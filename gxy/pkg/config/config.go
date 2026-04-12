package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	ListenAddr          string        `json:"listen_addr"`
	ListenPort          int           `json:"listen_port"`
	RegisterTTL         time.Duration `json:"register_ttl"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	SeedNodes           []string      `json:"seed_nodes"`
	HeartbeatInterval   time.Duration `json:"heartbeat_interval"`
	RegisterAuthKey     string        `json:"register_auth_key"`
	EnableAutoDiscovery bool          `json:"enable_auto_discovery"`
	NodeDiscoveryPort   int           `json:"node_discovery_port"`
}

func LoadConfig(filePath string) (*Config, error) {
	config := &Config{
		ListenAddr:          "0.0.0.0",
		ListenPort:          8385,
		RegisterTTL:         time.Second * 30,
		HealthCheckInterval: time.Second * 10,
		HealthCheckTimeout:  time.Second * 5,
		SeedNodes:           []string{},
		HeartbeatInterval:   time.Second * 5,
		RegisterAuthKey:     "ihvke@2025", // 使用正确的认证密钥
		EnableAutoDiscovery: true,
		NodeDiscoveryPort:   8384,
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
