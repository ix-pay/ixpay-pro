package api

import (
	"encoding/json"
	"net/http"

	"github.com/ixpay-pro/gxy/internal/discovery"
	"github.com/ixpay-pro/gxy/internal/proxy"
	"github.com/ixpay-pro/gxy/pkg/config"
	"github.com/ixpay-pro/gxy/pkg/utils"
)

type Handler struct {
	registry *discovery.Registry
	proxy    *proxy.Proxy
	config   *config.Config
	logger   *utils.Logger
}

func NewHandler(registry *discovery.Registry, proxy *proxy.Proxy, config *config.Config, logger *utils.Logger) *Handler {
	return &Handler{
		registry: registry,
		proxy:    proxy,
		config:   config,
		logger:   logger,
	}
}

func (h *Handler) RegisterService(w http.ResponseWriter, r *http.Request) {
	// 1. 验证注册凭证
	authHeader := r.Header.Get("Authorization")
	if !h.validateRegistrationAuth(authHeader) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. 解析服务注册请求
	var instance discovery.ServiceInstance
	if err := json.NewDecoder(r.Body).Decode(&instance); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 3. 注册服务实例
	h.registry.Register(&instance)
	h.logger.Info("Service registered: %s (%s:%d)", instance.Name, instance.Address, instance.Port)

	// 4. 返回注册结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Service registered successfully"})
}

func (h *Handler) DeregisterService(w http.ResponseWriter, r *http.Request) {
	// 1. 验证注册凭证
	authHeader := r.Header.Get("Authorization")
	if !h.validateRegistrationAuth(authHeader) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. 解析服务注销请求
	var req struct {
		ServiceName string `json:"service_name"`
		InstanceID  string `json:"instance_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 3. 注销服务实例
	h.registry.Deregister(req.ServiceName, req.InstanceID)
	h.logger.Info("Service deregistered: %s (instance: %s)", req.ServiceName, req.InstanceID)

	// 4. 返回注销结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Service deregistered successfully"})
}

func (h *Handler) GetServices(w http.ResponseWriter, r *http.Request) {
	// 1. 获取所有服务
	services := h.registry.GetAllServices()

	// 2. 返回服务列表
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)
}

func (h *Handler) validateRegistrationAuth(authHeader string) bool {
	// 从配置中获取预设的注册密码/密钥
	expectedAuthKey := h.config.RegisterAuthKey

	// 支持两种认证格式：
	// 1. 直接密钥匹配（向后兼容）
	// 2. Bearer认证格式 (Bearer <key>)
	if authHeader == expectedAuthKey {
		return true
	}

	// 检查Bearer格式
	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):] == expectedAuthKey
	}

	return false
}
