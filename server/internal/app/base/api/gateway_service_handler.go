package baseapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

type GatewayServiceController struct {
	logger logger.Logger
}

func NewGatewayServiceHandler(log logger.Logger) *GatewayServiceController {
	return &GatewayServiceController{
		logger: log,
	}
}

// GatewayServiceInfo 网关服务信息
type GatewayServiceInfo struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Address           string            `json:"address"`
	Port              int               `json:"port"`
	Metadata          map[string]string `json:"metadata"`
	LastSeen          string            `json:"lastSeen"`
	ActiveConnections int               `json:"activeConnections"`
	Status            string            `json:"status"`
}

// GetGatewayServices 获取网关注册的服务列表
// @Summary 获取网关服务列表
// @Description 从网关获取所有已注册的服务实例
// @Tags 网关管理
// @Accept json
// @Produce json
// @Success 200 {object} baseRes.Response{data=[]GatewayServiceInfo}
// @Router /gateway/services [get]
func (h *GatewayServiceController) GetGatewayServices(c *gin.Context) {
	gatewayURL := "http://127.0.0.1:8385"

	url := fmt.Sprintf("%s/api/services", gatewayURL)

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		h.logger.Error("获取网关服务列表失败", "error", err)
		baseRes.FailWithMessage("获取网关服务列表失败: "+err.Error(), c)
		return
	}
	defer resp.Body.Close()

	var services map[string][]GatewayServiceInfo
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		h.logger.Error("解析网关服务列表失败", "error", err)
		baseRes.FailWithMessage("解析网关服务列表失败", c)
		return
	}

	result := make([]GatewayServiceInfo, 0)
	for _, instances := range services {
		for _, instance := range instances {
			if instance.Metadata == nil {
				instance.Metadata = make(map[string]string)
			}

			instance.Status = "healthy"
			result = append(result, instance)
		}
	}

	if result == nil {
		result = make([]GatewayServiceInfo, 0)
	}

	baseRes.OkWithDetailed(result, "获取网关服务列表成功", c)
}
