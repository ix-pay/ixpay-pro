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
func (s *DepartmentService) CreateDepartment(name, description string, parentID, leaderID int64, createdBy int64, sort, status int) (*entity.Department, error) {
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
	if parentID != 0 {
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
func (s *DepartmentService) UpdateDepartment(id int64, name, description string, parentID, leaderID int64, updatedBy int64, sort, status int) (*entity.Department, error) {
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
		if parentID != 0 {
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
func (s *DepartmentService) isDescendant(children []*entity.Department, targetID int64) bool {
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
func (s *DepartmentService) DeleteDepartment(id int64) error {
	// 获取部门
	department, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取部门失败", "error", err, "id", id)
		return errors.New("部门不存在")
	}

	// 检查是否有子部门（包括禁用和启用的）
	allChildren, err := s.repo.GetChildrenByParentID(id)
	if err != nil {
		s.log.Error("获取子部门失败", "error", err, "id", id)
		return err
	}

	if len(allChildren) > 0 {
		// 检查是否有启用状态的子部门
		hasActiveChildren := false
		for _, child := range allChildren {
			if child.Status == 1 { // 1表示启用状态
				hasActiveChildren = true
				break
			}
		}

		if hasActiveChildren {
			s.log.Error("部门下有启用的子部门，无法删除", "id", id, "children_count", len(allChildren))
			return errors.New("部门下有启用的子部门，无法删除")
		}

		// 如果子部门都是禁用的，允许删除并级联删除所有子部门
		s.log.Info("部门下有禁用的子部门，将一并删除", "id", id, "children_count", len(allChildren))

		// 递归删除所有子部门
		for _, child := range allChildren {
			if err := s.DeleteDepartment(child.ID); err != nil {
				s.log.Error("删除子部门失败", "error", err, "child_id", child.ID)
				return errors.New("删除子部门失败：" + err.Error())
			}
		}
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
func (s *DepartmentService) GetDepartmentByID(id int64) (*entity.Department, error) {
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
	// ⭐ 获取所有部门数据
	tree, _, err := s.repo.List(1, 1000, map[string]interface{}{})
	if err != nil {
		s.log.Error("获取部门树失败", "error", err)
		return nil, err
	}
	return tree, nil
}

// GetDepartmentPath 获取部门路径
func (s *DepartmentService) GetDepartmentPath(id int64) ([]*entity.Department, error) {
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
func (s *DepartmentService) UpdateDepartmentLeader(id, leaderID int64, updatedBy int64) error {
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
