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

// writeJSON 统一响应格式 {code, data, msg}
func writeJSON(w http.ResponseWriter, httpStatus, code int, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]any{"code": code, "data": data, "msg": msg})
}

func (h *Handler) RegisterService(w http.ResponseWriter, r *http.Request) {
	// 1. 验证注册凭证
	authHeader := r.Header.Get("Authorization")
	if !h.validateRegistrationAuth(authHeader) {
		writeJSON(w, http.StatusUnauthorized, http.StatusUnauthorized, nil, "未授权：注册凭证无效")
		return
	}

	// 2. 解析服务注册请求
	var instance discovery.ServiceInstance
	if err := json.NewDecoder(r.Body).Decode(&instance); err != nil {
		writeJSON(w, http.StatusBadRequest, http.StatusBadRequest, nil, "解析服务注册请求失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	// 3. 注册服务实例（检查是否首次注册）
	isNew := true
	existingInstances := h.registry.GetInstances(instance.Name)
	for _, existing := range existingInstances {
		if existing.ID == instance.ID {
			isNew = false
			break
		}
	}

	h.registry.Register(&instance)
	if isNew {
		h.logger.Info("服务已注册：%s (%s:%d)", instance.Name, instance.Address, instance.Port)
	}

	// 4. 返回注册结果
	writeJSON(w, http.StatusOK, 0, nil, "服务注册成功")
}

func (h *Handler) DeregisterService(w http.ResponseWriter, r *http.Request) {
	// 1. 验证注册凭证
	authHeader := r.Header.Get("Authorization")
	if !h.validateRegistrationAuth(authHeader) {
		writeJSON(w, http.StatusUnauthorized, http.StatusUnauthorized, nil, "未授权：注册凭证无效")
		return
	}

	// 2. 解析服务注销请求
	var req struct {
		ServiceName string `json:"serviceName"`
		InstanceID  string `json:"instanceID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, http.StatusBadRequest, nil, "解析服务注销请求失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	// 3. 注销服务实例
	deregistered := h.registry.Deregister(req.ServiceName, req.InstanceID)
	if deregistered {
		h.logger.Debug("服务已注销：%s (实例 ID：%s)", req.ServiceName, req.InstanceID)
	}

	// 4. 返回注销结果
	writeJSON(w, http.StatusOK, 0, nil, "服务注销成功")
}

func (h *Handler) GetServices(w http.ResponseWriter, r *http.Request) {
	// 1. 获取所有服务
	services := h.registry.GetAllServices()

	// 2. 返回服务列表
	// 注意：该接口为服务间内部协议，后端按结构化类型直接解析裸数据，不包装 {code, data, msg}
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
