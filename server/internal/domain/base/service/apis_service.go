package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// APIService API路由服务
type APIService struct {
	repo repo.APIRepository
	log  logger.Logger
}

// NewAPIService 创建API路由服务实例
func NewAPIService(repo repo.APIRepository, log logger.Logger) *APIService {
	return &APIService{
		repo: repo,
		log:  log,
	}
}

// GetRoutes 获取所有 API 路由列表
func (s *APIService) GetRoutes() ([]*entity.API, error) {
	s.log.Info("获取所有 API 路由列表")
	return s.repo.GetAllRoutes()
}

// GetRouteByID 根据 ID 获取 API 路由
func (s *APIService) GetRouteByID(id int64) (*entity.API, error) {
	s.log.Info("根据 ID 获取 API 路由", "id", id)
	route, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取 API 路由失败", "error", err, "id", id)
		return nil, err
	}
	return route, nil
}

// GetRouteByPathAndMethod 根据路径和方法获取 API 路由
func (s *APIService) GetRouteByPathAndMethod(path, method string) (*entity.API, error) {
	s.log.Info("根据路径和方法获取 API 路由", "path", path, "method", method)
	return s.repo.GetByPathAndMethod(path, method)
}

// InitRoutes 初始化路由信息
func (s *APIService) InitRoutes() error {
	s.log.Info("初始化 API 路由")
	// 这里可以实现初始化路由的逻辑
	// 例如从配置文件或路由注册中获取路由信息并保存到数据库
	return nil
}

// BatchUpdateRoutes 批量更新路由信息
func (s *APIService) BatchUpdateRoutes(routes []*entity.API) error {
	s.log.Info("批量更新 API 路由", "count", len(routes))
	return s.repo.BatchSave(routes)
}

// Create 创建 API 路由
func (s *APIService) Create(route *entity.API, createdBy int64) error {
	s.log.Info("创建 API 路由", "path", route.Path, "method", route.Method)
	return s.repo.Create(route)
}

// Update 更新 API 路由
func (s *APIService) Update(route *entity.API, updatedBy int64) error {
	s.log.Info("更新 API 路由", "id", route.ID, "path", route.Path, "method", route.Method)
	route.UpdatedBy = updatedBy
	return s.repo.Update(route)
}

// Delete 删除 API 路由
func (s *APIService) Delete(id int64) error {
	s.log.Info("删除 API 路由", "id", id)
	return s.repo.Delete(id)
}

// GetList 获取 API 路由列表，支持分页和过滤
func (s *APIService) GetList(page, pageSize int, filters map[string]interface{}) ([]*entity.API, int, error) {
	s.log.Info("获取 API 路由列表", "page", page, "pageSize", pageSize, "filters", filters)
	list, total64, err := s.repo.List(page, pageSize, filters)
	return list, int(total64), err
}

// CheckRouteAuth 检查路由是否需要认证
func (s *APIService) CheckRouteAuth(path, method string) (bool, error) {
	s.log.Info("检查路由认证要求", "path", path, "method", method)
	route, err := s.repo.GetByPathAndMethod(path, method)
	if err != nil {
		s.log.Error("获取路由失败", "error", err, "path", path, "method", method)
		// 如果路由不存在，默认返回需要认证
		return true, nil
	}
	return route.AuthRequired, nil
}

// CreateAPIRoute 创建 API 路由
func (s *APIService) CreateAPIRoute(route *entity.API, createdBy int64) error {
	s.log.Info("创建 API 路由", "path", route.Path, "method", route.Method)

	// 检查路由是否已存在
	existingRoute, err := s.repo.GetByPathAndMethod(route.Path, route.Method)
	if err == nil && existingRoute != nil {
		s.log.Error("API 路由已存在", "path", route.Path, "method", route.Method)
		return errors.New("API 路由已存在")
	}

	// 设置责任主体字段
	route.CreatedBy = createdBy
	route.UpdatedBy = createdBy

	// 创建路由
	if err := s.repo.Create(route); err != nil {
		s.log.Error("创建 API 路由失败", "error", err)
		return err
	}

	s.log.Info("API 路由创建成功", "id", route.ID)
	return nil
}

// UpdateAPIRoute 更新 API 路由
func (s *APIService) UpdateAPIRoute(route *entity.API, updatedBy int64) error {
	s.log.Info("更新 API 路由", "id", route.ID)

	// 检查路由是否存在
	existingRoute, err := s.repo.GetByID(route.ID)
	if err != nil {
		s.log.Error("获取 API 路由失败", "error", err, "id", route.ID)
		return err
	}

	// 检查路径和方法是否已被其他路由使用
	if existingRoute.Path != route.Path || existingRoute.Method != route.Method {
		checkRoute, err := s.repo.GetByPathAndMethod(route.Path, route.Method)
		if err == nil && checkRoute != nil && checkRoute.ID != route.ID {
			s.log.Error("API 路由路径和方法组合已存在", "path", route.Path, "method", route.Method)
			return errors.New("API 路由路径和方法组合已存在")
		}
	}

	// 设置更新信息
	route.UpdatedBy = updatedBy

	// 更新路由
	if err := s.repo.Update(route); err != nil {
		s.log.Error("更新 API 路由失败", "error", err)
		return err
	}

	s.log.Info("API 路由更新成功", "id", route.ID)
	return nil
}

// DeleteAPIRoute 删除 API 路由
func (s *APIService) DeleteAPIRoute(id int64) error {
	s.log.Info("删除 API 路由", "id", id)

	// 检查路由是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取 API 路由失败", "error", err, "id", id)
		return err
	}

	// 删除路由
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除 API 路由失败", "error", err)
		return err
	}

	s.log.Info("API 路由删除成功", "id", id)
	return nil
}

// GetAPIRouteList 获取 API 路由列表，支持分页和过滤
func (s *APIService) GetAPIRouteList(page, pageSize int, filters map[string]interface{}) ([]*entity.API, int64, error) {
	s.log.Info("获取 API 路由列表", "page", page, "pageSize", pageSize)

	routes, total, err := s.repo.List(int(page), int(pageSize), filters)
	if err != nil {
		s.log.Error("获取 API 路由列表失败", "error", err)
		return nil, 0, err
	}

	s.log.Info("API 路由列表获取成功", "count", len(routes), "total", total)
	return routes, total, nil
}
