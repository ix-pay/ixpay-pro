package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// BtnPermService 按钮权限服务实现
type BtnPermService struct {
	btnPermRepo repo.BtnPermRepository
	logger      logger.Logger
}

// NewBtnPermService 创建按钮权限服务实例
func NewBtnPermService(btnPermRepo repo.BtnPermRepository, logger logger.Logger) *BtnPermService {
	return &BtnPermService{
		btnPermRepo: btnPermRepo,
		logger:      logger,
	}
}

// CreateBtnPerm 创建按钮权限
func (s *BtnPermService) CreateBtnPerm(btnPerm *entity.BtnPerm, createdBy string) error {
	// 验证参数
	if btnPerm == nil {
		return errors.New("无效的参数")
	}
	if btnPerm.MenuID == "" {
		return errors.New("菜单 ID 不能为空")
	}
	if btnPerm.Name == "" {
		return errors.New("按钮名称不能为空")
	}
	if btnPerm.Code == "" {
		return errors.New("按钮编码不能为空")
	}

	// 检查权限编码是否已存在
	existingBtnPerm, err := s.btnPermRepo.GetByCode(btnPerm.Code)
	if err != nil {
		s.logger.Error("检查按钮编码失败", "error", err, "code", btnPerm.Code)
		return errors.New("检查按钮编码失败")
	}
	if existingBtnPerm != nil {
		return errors.New("按钮编码已存在")
	}

	// 设置创建者
	btnPerm.CreatedBy = createdBy
	btnPerm.UpdatedBy = createdBy

	// 创建按钮权限
	err = s.btnPermRepo.Create(btnPerm)
	if err != nil {
		s.logger.Error("创建按钮权限失败", "error", err, "btn_perm", btnPerm)
		return err
	}

	s.logger.Info("创建按钮权限成功", "btn_perm_id", btnPerm.ID, "btn_perm_name", btnPerm.Name)
	return nil
}

// UpdateBtnPerm 更新按钮权限
func (s *BtnPermService) UpdateBtnPerm(btnPerm *entity.BtnPerm, updatedBy string) error {
	// 验证参数
	if btnPerm == nil || btnPerm.ID == "" {
		return errors.New("无效的参数")
	}
	if btnPerm.MenuID == "" {
		return errors.New("菜单 ID 不能为空")
	}
	if btnPerm.Name == "" {
		return errors.New("按钮名称不能为空")
	}
	if btnPerm.Code == "" {
		return errors.New("按钮编码不能为空")
	}

	// 检查按钮权限是否存在
	existingBtnPerm, err := s.btnPermRepo.GetByID(btnPerm.ID)
	if err != nil {
		s.logger.Error("获取按钮权限失败", "error", err, "btn_perm_id", btnPerm.ID)
		return err
	}

	// 检查权限编码是否与其他记录冲突
	if btnPerm.Code != existingBtnPerm.Code {
		codeBtnPerm, checkErr := s.btnPermRepo.GetByCode(btnPerm.Code)
		if checkErr == nil && codeBtnPerm != nil && codeBtnPerm.ID != btnPerm.ID {
			return errors.New("按钮编码已存在")
		}
	}

	// 更新字段
	existingBtnPerm.Name = btnPerm.Name
	existingBtnPerm.Description = btnPerm.Description
	existingBtnPerm.Code = btnPerm.Code
	existingBtnPerm.Status = btnPerm.Status
	existingBtnPerm.MenuID = btnPerm.MenuID
	existingBtnPerm.UpdatedBy = updatedBy

	// 保存更新
	err = s.btnPermRepo.Update(existingBtnPerm)
	if err != nil {
		s.logger.Error("更新按钮权限失败", "error", err, "btn_perm_id", btnPerm.ID)
		return err
	}

	s.logger.Info("更新按钮权限成功", "btn_perm_id", btnPerm.ID)
	return nil
}

// DeleteBtnPerm 删除按钮权限
func (s *BtnPermService) DeleteBtnPerm(id string) error {
	// 验证参数
	if id == "" {
		return errors.New("无效的参数")
	}

	// 检查按钮权限是否存在
	_, err := s.btnPermRepo.GetByID(id)
	if err != nil {
		s.logger.Error("获取按钮权限失败", "error", err, "btn_perm_id", id)
		return err
	}

	// 删除按钮权限
	err = s.btnPermRepo.Delete(id)
	if err != nil {
		s.logger.Error("删除按钮权限失败", "error", err, "btn_perm_id", id)
		return err
	}

	s.logger.Info("删除按钮权限成功", "btn_perm_id", id)
	return nil
}

// GetBtnPermByID 获取按钮权限详情
func (s *BtnPermService) GetBtnPermByID(id string) (*entity.BtnPerm, error) {
	// 验证参数
	if id == "" {
		return nil, errors.New("无效的参数")
	}

	// 获取按钮权限
	btnPerm, err := s.btnPermRepo.GetByID(id)
	if err != nil {
		s.logger.Error("获取按钮权限失败", "error", err, "btn_perm_id", id)
		return nil, err
	}

	return btnPerm, nil
}

// GetBtnPermList 获取按钮权限列表，支持分页和过滤
func (s *BtnPermService) GetBtnPermList(page, pageSize int, filters map[string]interface{}) ([]*entity.BtnPerm, int64, error) {
	// 验证分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取按钮权限列表
	btnPerms, total, err := s.btnPermRepo.List(page, pageSize, filters)
	if err != nil {
		s.logger.Error("获取按钮权限列表失败", "error", err, "page", page, "pageSize", pageSize)
		return nil, 0, err
	}

	return btnPerms, total, nil
}

// GetBtnPermsByMenu 获取菜单下的所有按钮权限
func (s *BtnPermService) GetBtnPermsByMenu(menuID string) ([]*entity.BtnPerm, error) {
	// 验证参数
	if menuID == "" {
		return nil, errors.New("无效的参数")
	}

	// 获取菜单下的按钮权限
	btnPerms, err := s.btnPermRepo.GetBtnPermsByMenu(menuID)
	if err != nil {
		s.logger.Error("获取菜单下的按钮权限失败", "error", err, "menu_id", menuID)
		return nil, err
	}

	return btnPerms, nil
}

// AssignAPIToBtnPerm 分配 API 路由到按钮权限
func (s *BtnPermService) AssignAPIToBtnPerm(btnPermID, routeID string) error {
	// 验证参数
	if btnPermID == "" || routeID == "" {
		return errors.New("无效的参数")
	}

	// 检查按钮权限是否存在
	_, err := s.btnPermRepo.GetByID(btnPermID)
	if err != nil {
		s.logger.Error("获取按钮权限失败", "error", err, "btn_perm_id", btnPermID)
		return err
	}

	// 分配 API 路由
	err = s.btnPermRepo.AddAPIToBtnPerm(btnPermID, routeID)
	if err != nil {
		s.logger.Error("分配 API 路由到按钮权限失败", "error", err, "btn_perm_id", btnPermID, "route_id", routeID)
		return err
	}

	s.logger.Info("分配 API 路由到按钮权限成功", "btn_perm_id", btnPermID, "route_id", routeID)
	return nil
}

// RevokeAPIFromBtnPerm 撤销 API 路由从按钮权限
func (s *BtnPermService) RevokeAPIFromBtnPerm(btnPermID, routeID string) error {
	// 验证参数
	if btnPermID == "" || routeID == "" {
		return errors.New("无效的参数")
	}

	// 撤销 API 路由
	err := s.btnPermRepo.RemoveAPIFromBtnPerm(btnPermID, routeID)
	if err != nil {
		s.logger.Error("撤销 API 路由从按钮权限失败", "error", err, "btn_perm_id", btnPermID, "route_id", routeID)
		return err
	}

	s.logger.Info("撤销 API 路由从按钮权限成功", "btn_perm_id", btnPermID, "route_id", routeID)
	return nil
}

// GetAPIsForBtnPerm 获取按钮权限关联的 API 路由
func (s *BtnPermService) GetAPIsForBtnPerm(btnPermID string) ([]*entity.API, error) {
	// 验证参数
	if btnPermID == "" {
		return nil, errors.New("无效的参数")
	}

	// 获取关联的 API 路由
	routes, err := s.btnPermRepo.GetAPIsByBtnPerm(btnPermID)
	if err != nil {
		s.logger.Error("获取按钮权限关联的 API 路由失败", "error", err, "btn_perm_id", btnPermID)
		return nil, err
	}

	return routes, nil
}

// GetBtnPermsForAPI 获取关联指定 API 路由的所有按钮权限
func (s *BtnPermService) GetBtnPermsForAPI(routeID string) ([]*entity.BtnPerm, error) {
	// 验证参数
	if routeID == "" {
		return nil, errors.New("无效的参数")
	}

	// 获取关联的按钮权限
	btnPerms, err := s.btnPermRepo.GetBtnPermsByAPI(routeID)
	if err != nil {
		s.logger.Error("获取 API 路由关联的按钮权限失败", "error", err, "route_id", routeID)
		return nil, err
	}

	return btnPerms, nil
}
