package util

import (
	"fmt"
	"strings"
)

// APIDescriber API 描述生成器
type APIDescriber struct {
	// 资源映射表：英文路径片段 -> 中文资源名
	resourceMappings map[string]string
	// 方法映射表：HTTP 方法 -> 中文动作
	methodMappings map[string]string
}

// NewAPIDescriber 创建 API 描述生成器实例
func NewAPIDescriber() *APIDescriber {
	return &APIDescriber{
		resourceMappings: buildResourceMappings(),
		methodMappings: map[string]string{
			"GET":    "查询",
			"POST":   "创建",
			"PUT":    "更新",
			"DELETE": "删除",
			"PATCH":  "部分更新",
		},
	}
}

// buildResourceMappings 构建资源映射表
func buildResourceMappings() map[string]string {
	return map[string]string{
		// 认证相关
		"auth":          "认证",
		"login":         "登录",
		"logout":        "登出",
		"register":      "注册",
		"captcha":       "验证码",
		"refresh-token": "刷新令牌",
		// 用户相关
		"user":     "用户",
		"info":     "信息",
		"password": "密码",
		"profile":  "资料",
		// 角色相关
		"role":       "角色",
		"permission": "权限",
		"menu":       "菜单",
		// 部门相关
		"dept":       "部门",
		"department": "部门",
		"position":   "职位",
		// 字典相关
		"dict": "字典",
		"item": "字典项",
		// 配置相关
		"config":  "配置",
		"setting": "设置",
		// 通知相关
		"notice":       "公告",
		"notification": "通知",
		// API 相关
		"api":   "API 路由",
		"route": "路由",
		// 按钮权限
		"btn-perm": "按钮权限",
		"button":   "按钮",
		// 日志相关
		"log":           "日志",
		"login-log":     "登录日志",
		"operation-log": "操作日志",
		"audit-log":     "审计日志",
		// 监控相关
		"monitor":     "监控",
		"online-user": "在线用户",
		"system":      "系统",
		// 任务相关
		"task":      "任务",
		"job":       "任务",
		"schedule":  "定时",
		"execution": "执行",
		// 其他
		"file":       "文件",
		"upload":     "上传",
		"download":   "下载",
		"export":     "导出",
		"import":     "导入",
		"statistics": "统计",
		"dashboard":  "仪表盘",
		"health":     "健康检查",
	}
}

// GenerateDescription 根据路径和方法生成描述
func (d *APIDescriber) GenerateDescription(path, method string) string {
	// 提取路径中的关键部分
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return "系统接口"
	}

	// 查找最有意义的资源名称
	resourceName := d.findResourceName(parts)
	if resourceName == "" {
		resourceName = "资源"
	}

	// 获取动作描述
	action := d.methodMappings[method]
	if action == "" {
		action = "操作"
	}

	return fmt.Sprintf("%s%s接口", action, resourceName)
}

// findResourceName 从路径片段中查找最合适的资源名
func (d *APIDescriber) findResourceName(parts []string) string {
	// 从后往前查找（通常资源名在路径后部）
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]

		// 跳过 ID 参数（如 :id）
		if strings.HasPrefix(part, ":") {
			continue
		}
		// 跳过版本号（如 v1, v2）
		if part == "v1" || part == "v2" {
			continue
		}
		// 跳过模块前缀（如 api, admin, wx）
		if part == "api" || part == "admin" || part == "wx" {
			continue
		}

		// 转换为小写进行匹配
		lowerPart := strings.ToLower(part)

		// 尝试匹配映射表
		if cnName, ok := d.resourceMappings[lowerPart]; ok {
			return cnName
		}

		// 如果没有映射，使用原名称（处理驼峰命名）
		return d.camelToChinese(part)
	}

	return ""
}

// camelToChinese 将驼峰命名转换为简单的中文描述（基础版本）
func (d *APIDescriber) camelToChinese(s string) string {
	// 简单的驼峰处理：大写字母前加空格
	result := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += " "
		}
		result += string(r)
	}
	return result
}

// 全局描述生成器实例
var globalDescriber = NewAPIDescriber()

// GenerateAPIDescription 全局函数：生成 API 描述
func GenerateAPIDescription(path, method string) string {
	return globalDescriber.GenerateDescription(path, method)
}
