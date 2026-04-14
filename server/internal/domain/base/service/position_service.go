package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// PositionService 岗位服务实现
type PositionService struct {
	repo repo.PositionRepository
	log  logger.Logger
}

// NewPositionService 创建岗位服务实例
func NewPositionService(repo repo.PositionRepository, log logger.Logger) *PositionService {
	return &PositionService{
		repo: repo,
		log:  log,
	}
}

// CreatePosition 创建岗位
func (s *PositionService) CreatePosition(name, description string, createdBy int64, sort, status int) (*entity.Position, error) {
	// 参数验证
	if name == "" {
		return nil, errors.New("岗位名称不能为空")
	}

	// 检查岗位名称是否已存在
	existingPosition, err := s.repo.GetByName(name)
	if err == nil && existingPosition != nil {
		s.log.Error("岗位名称已存在", "name", name)
		return nil, errors.New("岗位名称已存在")
	}

	// 创建岗位
	position := &entity.Position{
		Name:        name,
		Sort:        sort,
		Status:      status,
		Description: description,
		CreatedBy:   createdBy,
	}

	if err := s.repo.Create(position); err != nil {
		s.log.Error("创建岗位失败", "error", err, "name", name)
		return nil, err
	}

	s.log.Info("创建岗位成功", "id", position.ID, "name", name)
	return position, nil
}

// UpdatePosition 更新岗位
func (s *PositionService) UpdatePosition(id int64, name, description string, updatedBy int64, sort, status int) (*entity.Position, error) {
	// 获取岗位
	position, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取岗位失败", "error", err, "id", id)
		return nil, errors.New("岗位不存在")
	}

	// 检查岗位名称是否已存在（排除当前岗位）
	if name != position.Name {
		existingPosition, err := s.repo.GetByName(name)
		if err == nil && existingPosition != nil {
			s.log.Error("岗位名称已存在", "name", name, "existing_id", existingPosition.ID)
			return nil, errors.New("岗位名称已存在")
		}
	}

	// 更新岗位信息
	position.Name = name
	position.Description = description
	position.Sort = sort
	position.Status = status
	position.UpdatedBy = updatedBy

	if err := s.repo.Update(position); err != nil {
		s.log.Error("更新岗位失败", "error", err, "id", id)
		return nil, err
	}

	s.log.Info("更新岗位成功", "id", id, "name", name)
	return position, nil
}

// DeletePosition 删除岗位
func (s *PositionService) DeletePosition(id int64) error {
	// 获取岗位
	position, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取岗位失败", "error", err, "id", id)
		return errors.New("岗位不存在")
	}

	// 删除岗位
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除岗位失败", "error", err, "id", id)
		return err
	}

	s.log.Info("删除岗位成功", "id", id, "name", position.Name)
	return nil
}

// GetPositionByID 获取岗位详情
func (s *PositionService) GetPositionByID(id int64) (*entity.Position, error) {
	position, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取岗位失败", "error", err, "id", id)
		return nil, errors.New("岗位不存在")
	}
	return position, nil
}

// GetPositionList 获取岗位列表
func (s *PositionService) GetPositionList(page, pageSize int, filters map[string]interface{}) ([]*entity.Position, int64, error) {
	positions, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取岗位列表失败", "error", err)
		return nil, 0, err
	}
	return positions, total, nil
}

// GetAllPositions 获取所有岗位
func (s *PositionService) GetAllPositions() ([]*entity.Position, error) {
	positions, err := s.repo.GetAll()
	if err != nil {
		s.log.Error("获取所有岗位失败", "error", err)
		return nil, err
	}
	return positions, nil
}
