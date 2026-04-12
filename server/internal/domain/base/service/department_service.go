package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// DepartmentService 部门服务实现
type DepartmentService struct {
	repo repo.DepartmentRepository
	log  logger.Logger
}

// NewDepartmentService 创建部门服务实例
func NewDepartmentService(repo repo.DepartmentRepository, log logger.Logger) *DepartmentService {
	return &DepartmentService{
		repo: repo,
		log:  log,
	}
}

// CreateDepartment 创建部门
func (s *DepartmentService) CreateDepartment(name, description, parentID, leaderID string, createdBy string, sort, status int) (*entity.Department, error) {
	// 检查部门名称是否已存在（在同一父部门下）
	existingDepts, err := s.repo.GetChildrenByParentID(parentID)
	if err != nil {
		s.log.Error("检查部门名称失败", "error", err, "parent_id", parentID)
		return nil, err
	}

	for _, dept := range existingDepts {
		if dept.Name == name {
			s.log.Error("部门名称已存在", "name", name, "parent_id", parentID)
			return nil, errors.New("部门名称已存在")
		}
	}

	// 如果指定了父部门，验证父部门是否存在
	if parentID != "" {
		_, err := s.repo.GetByID(parentID)
		if err != nil {
			s.log.Error("父部门不存在", "error", err, "parent_id", parentID)
			return nil, errors.New("父部门不存在")
		}
	}

	// 如果指定了部门负责人，验证负责人是否存在（这里简化处理，实际应该调用用户服务验证）
	// 如果需要严格验证，可以注入用户服务进行验证

	// 创建部门
	department := &entity.Department{
		Name:        name,
		ParentID:    parentID,
		LeaderID:    leaderID,
		Sort:        sort,
		Status:      status,
		Description: description,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	if err := s.repo.Create(department); err != nil {
		s.log.Error("创建部门失败", "error", err, "name", name)
		return nil, err
	}

	s.log.Info("创建部门成功", "id", department.ID, "name", name, "parent_id", parentID)
	return department, nil
}

// UpdateDepartment 更新部门
func (s *DepartmentService) UpdateDepartment(id, name, description, parentID, leaderID string, updatedBy string, sort, status int) (*entity.Department, error) {
	// 获取部门
	department, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取部门失败", "error", err, "id", id)
		return nil, errors.New("部门不存在")
	}

	// 检查部门名称是否已存在（在同一父部门下，排除当前部门）
	if parentID == department.ParentID {
		// 父部门未变化，只检查同级部门
		existingDepts, err := s.repo.GetChildrenByParentID(parentID)
		if err != nil {
			s.log.Error("检查部门名称失败", "error", err, "parent_id", parentID)
			return nil, err
		}

		for _, dept := range existingDepts {
			if dept.Name == name && dept.ID != id {
				s.log.Error("部门名称已存在", "name", name, "parent_id", parentID)
				return nil, errors.New("部门名称已存在")
			}
		}
	} else {
		// 父部门发生变化，检查新父部门下是否有重名部门
		existingDepts, err := s.repo.GetChildrenByParentID(parentID)
		if err != nil {
			s.log.Error("检查部门名称失败", "error", err, "parent_id", parentID)
			return nil, err
		}

		for _, dept := range existingDepts {
			if dept.Name == name {
				s.log.Error("部门名称已存在", "name", name, "parent_id", parentID)
				return nil, errors.New("部门名称已存在")
			}
		}

		// 验证新父部门是否存在
		if parentID != "" {
			_, err := s.repo.GetByID(parentID)
			if err != nil {
				s.log.Error("新父部门不存在", "error", err, "parent_id", parentID)
				return nil, errors.New("新父部门不存在")
			}
		}

		// 不能将部门移动到其子部门下（避免循环）
		children, err := s.repo.GetChildrenByParentID(id)
		if err != nil {
			s.log.Error("获取子部门失败", "error", err, "id", id)
			return nil, err
		}

		// 递归检查所有子孙部门
		if s.isDescendant(children, parentID) {
			s.log.Error("不能将部门移动到其子部门下", "id", id, "parent_id", parentID)
			return nil, errors.New("不能将部门移动到其子部门下")
		}
	}

	// 更新部门信息
	department.Name = name
	department.ParentID = parentID
	department.LeaderID = leaderID
	department.Sort = sort
	department.Status = status
	department.Description = description
	department.UpdatedBy = updatedBy

	if err := s.repo.Update(department); err != nil {
		s.log.Error("更新部门失败", "error", err, "id", id)
		return nil, err
	}

	s.log.Info("更新部门成功", "id", id, "name", name, "parent_id", parentID)
	return department, nil
}

// isDescendant 递归检查是否是子孙部门
func (s *DepartmentService) isDescendant(children []*entity.Department, targetID string) bool {
	for _, child := range children {
		if child.ID == targetID {
			return true
		}
		// 递归检查子部门的子部门
		grandChildren, err := s.repo.GetChildrenByParentID(child.ID)
		if err == nil && len(grandChildren) > 0 {
			if s.isDescendant(grandChildren, targetID) {
				return true
			}
		}
	}
	return false
}

// DeleteDepartment 删除部门
func (s *DepartmentService) DeleteDepartment(id string) error {
	// 获取部门
	department, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取部门失败", "error", err, "id", id)
		return errors.New("部门不存在")
	}

	// 检查是否有子部门
	children, err := s.repo.GetChildrenByParentID(id)
	if err != nil {
		s.log.Error("获取子部门失败", "error", err, "id", id)
		return err
	}

	if len(children) > 0 {
		s.log.Error("部门下有子部门，无法删除", "id", id, "children_count", len(children))
		return errors.New("部门下有子部门，无法删除")
	}

	// 删除部门
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除部门失败", "error", err, "id", id)
		return err
	}

	s.log.Info("删除部门成功", "id", id, "name", department.Name)
	return nil
}

// GetDepartmentByID 获取部门详情
func (s *DepartmentService) GetDepartmentByID(id string) (*entity.Department, error) {
	// ⭐ 优化：使用 Preload 加载子部门、父部门、负责人
	department, err := s.repo.GetByID(id,
		repo.DepartmentRelationChildren,
		repo.DepartmentRelationParent,
		repo.DepartmentRelationLeader)
	if err != nil {
		s.log.Error("获取部门失败", "error", err, "id", id)
		return nil, errors.New("部门不存在")
	}
	return department, nil
}

// GetDepartmentList 获取部门列表
func (s *DepartmentService) GetDepartmentList(page, pageSize int, filters map[string]interface{}) ([]*entity.Department, int64, error) {
	departments, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取部门列表失败", "error", err)
		return nil, 0, err
	}
	return departments, total, nil
}

// GetDepartmentTree 获取部门树形结构
func (s *DepartmentService) GetDepartmentTree() ([]*entity.Department, error) {
	// ⭐ 优化：使用 Preload 一次性加载所有子部门和负责人
	// 先获取所有顶级部门（parent_id = 0）
	tree, _, err := s.repo.List(1, 1000, map[string]interface{}{"parent_id = ?": 0})
	if err != nil {
		s.log.Error("获取部门树失败", "error", err)
		return nil, err
	}
	return tree, nil
}

// GetDepartmentPath 获取部门路径
func (s *DepartmentService) GetDepartmentPath(id string) ([]*entity.Department, error) {
	// ⭐ 优化：使用 Preload 加载父部门
	department, err := s.repo.GetByID(id, repo.DepartmentRelationParent)
	if err != nil {
		s.log.Error("获取部门失败", "error", err, "id", id)
		return nil, err
	}

	// 构建部门路径（从根到当前部门）
	var path []*entity.Department
	current := department
	for current != nil {
		path = append([]*entity.Department{current}, path...)
		current = current.Parent
	}

	return path, nil
}

// UpdateDepartmentLeader 更新部门负责人
func (s *DepartmentService) UpdateDepartmentLeader(id, leaderID string, updatedBy string) error {
	// 获取部门
	department, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取部门失败", "error", err, "id", id)
		return errors.New("部门不存在")
	}

	// 如果指定了负责人，验证负责人是否存在（这里简化处理）
	// 如果需要严格验证，可以注入用户服务进行验证

	// 更新负责人
	department.LeaderID = leaderID
	department.UpdatedBy = updatedBy

	if err := s.repo.Update(department); err != nil {
		s.log.Error("更新部门负责人失败", "error", err, "id", id, "leader_id", leaderID)
		return err
	}

	s.log.Info("更新部门负责人成功", "id", id, "leader_id", leaderID, "updated_by", updatedBy)
	return nil
}
