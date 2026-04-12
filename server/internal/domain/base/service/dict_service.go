package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// DictService 实现字典领域服务接口
type DictService struct {
	repo repo.DictRepository
	log  logger.Logger
}

// NewDictService 创建字典服务实例
func NewDictService(repo repo.DictRepository, log logger.Logger) *DictService {
	return &DictService{
		repo: repo,
		log:  log,
	}
}

// GetDictByID 根据 ID 获取字典
func (s *DictService) GetDictByID(id string) (*entity.Dict, error) {
	dict, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取字典失败", "id", id, "error", err)
		return nil, errors.New("字典不存在")
	}
	return dict, nil
}

// GetDictByCode 根据字典编码获取字典
func (s *DictService) GetDictByCode(dictCode string) (*entity.Dict, error) {
	dict, err := s.repo.GetByCode(dictCode)
	if err != nil {
		s.log.Error("获取字典失败", "dict_code", dictCode, "error", err)
		return nil, errors.New("字典不存在")
	}
	return dict, nil
}

// CreateDict 创建字典
func (s *DictService) CreateDict(dictName, dictCode, description string, status int, createdBy string) (*entity.Dict, error) {
	// 检查字典编码是否已存在
	existingDict, err := s.repo.GetByCode(dictCode)
	if err == nil {
		if existingDict.ID != "" {
			s.log.Error("字典编码已存在", "dict_code", dictCode)
			return nil, errors.New("字典编码已存在")
		}
	}

	// 创建字典
	dict := &entity.Dict{
		DictName:    dictName,
		DictCode:    dictCode,
		Description: description,
		Status:      status,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	// 保存字典
	if err := s.repo.Create(dict); err != nil {
		s.log.Error("创建字典失败", "error", err)
		return nil, errors.New("创建字典失败")
	}

	s.log.Info("创建字典成功", "dict_code", dictCode)
	return dict, nil
}

// UpdateDict 更新字典
func (s *DictService) UpdateDict(id string, dictName, dictCode, description string, status int, updatedBy string) error {
	// 获取字典
	dict, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取字典失败", "id", id, "error", err)
		return errors.New("字典不存在")
	}

	// 检查字典编码是否已被其他字典使用
	if dict.DictCode != dictCode {
		existingDict, err := s.repo.GetByCode(dictCode)
		if err == nil && existingDict.ID != id {
			s.log.Error("字典编码已存在", "dict_code", dictCode)
			return errors.New("字典编码已被使用")
		}
	}

	// 更新字典
	dict.DictName = dictName
	dict.DictCode = dictCode
	dict.Description = description
	dict.Status = status
	dict.UpdatedBy = updatedBy

	// 保存更新
	if err := s.repo.Update(dict); err != nil {
		s.log.Error("更新字典失败", "error", err)
		return errors.New("更新字典失败")
	}

	s.log.Info("更新字典成功", "dict_code", dictCode)
	return nil
}

// DeleteDict 删除字典
func (s *DictService) DeleteDict(id string) error {
	// 获取字典
	_, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取字典失败", "id", id, "error", err)
		return errors.New("字典不存在")
	}

	// 删除字典
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除字典失败", "error", err)
		return errors.New("删除字典失败")
	}

	s.log.Info("删除字典成功", "id", id)
	return nil
}

// GetDictList 获取字典列表
func (s *DictService) GetDictList(page, pageSize int64, filters map[string]interface{}) ([]*entity.Dict, int64, error) {
	dicts, total, err := s.repo.List(int(page), int(pageSize), filters)
	if err != nil {
		s.log.Error("获取字典列表失败", "error", err)
		return nil, 0, errors.New("获取字典列表失败")
	}
	return dicts, total, nil
}

// GetAllActiveDicts 获取所有启用的字典
func (s *DictService) GetAllActiveDicts() ([]*entity.Dict, error) {
	dicts, err := s.repo.GetAllActive()
	if err != nil {
		s.log.Error("获取启用的字典列表失败", "error", err)
		return nil, errors.New("获取启用的字典列表失败")
	}
	return dicts, nil
}

// DictItemService 实现字典项领域服务接口
type DictItemService struct {
	repo     repo.DictItemRepository
	dictRepo repo.DictRepository
	log      logger.Logger
}

// NewDictItemService 创建字典项服务实例
func NewDictItemService(repo repo.DictItemRepository, dictRepo repo.DictRepository, log logger.Logger) *DictItemService {
	return &DictItemService{
		repo:     repo,
		dictRepo: dictRepo,
		log:      log,
	}
}

// GetDictItemByID 根据 ID 获取字典项
func (s *DictItemService) GetDictItemByID(id string) (*entity.DictItem, error) {
	dictItem, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取字典项失败", "id", id, "error", err)
		return nil, errors.New("字典项不存在")
	}
	return dictItem, nil
}

// GetDictItemsByDictID 根据字典 ID 获取字典项列表
func (s *DictItemService) GetDictItemsByDictID(dictID string) ([]*entity.DictItem, error) {
	// 检查字典是否存在
	_, err := s.dictRepo.GetByID(dictID)
	if err != nil {
		s.log.Error("获取字典失败", "dict_id", dictID, "error", err)
		return nil, errors.New("字典不存在")
	}

	dictItems, err := s.repo.GetByDictID(dictID)
	if err != nil {
		s.log.Error("获取字典项列表失败", "dict_id", dictID, "error", err)
		return nil, errors.New("获取字典项列表失败")
	}
	return dictItems, nil
}

// CreateDictItem 创建字典项
func (s *DictItemService) CreateDictItem(dictID string, itemKey, itemValue, description string, sort, status int, createdBy string) (*entity.DictItem, error) {
	// 检查字典是否存在
	_, err := s.dictRepo.GetByID(dictID)
	if err != nil {
		s.log.Error("获取字典失败", "dict_id", dictID, "error", err)
		return nil, errors.New("字典不存在")
	}

	// 检查字典项键是否已存在
	dictItems, err := s.repo.GetByDictID(dictID)
	if err == nil {
		for _, item := range dictItems {
			if item.ItemKey == itemKey {
				s.log.Error("字典项键已存在", "dict_id", dictID, "item_key", itemKey)
				return nil, errors.New("字典项键已存在")
			}
		}
	}

	// 创建字典项
	dictItem := &entity.DictItem{
		DictID:      dictID,
		ItemKey:     itemKey,
		ItemValue:   itemValue,
		Sort:        sort,
		Description: description,
		Status:      status,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	// 保存字典项
	if err := s.repo.Create(dictItem); err != nil {
		s.log.Error("创建字典项失败", "error", err)
		return nil, errors.New("创建字典项失败")
	}

	s.log.Info("创建字典项成功", "dict_id", dictID, "item_key", itemKey)
	return dictItem, nil
}

// UpdateDictItem 更新字典项
func (s *DictItemService) UpdateDictItem(id string, dictID string, itemKey, itemValue, description string, sort, status int, updatedBy string) error {
	// 获取字典项
	dictItem, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取字典项失败", "id", id, "error", err)
		return errors.New("字典项不存在")
	}

	// 检查字典是否存在
	_, err = s.dictRepo.GetByID(dictID)
	if err != nil {
		s.log.Error("获取字典失败", "dict_id", dictID, "error", err)
		return errors.New("字典不存在")
	}

	// 检查字典项键是否已被其他字典项使用
	if dictItem.ItemKey != itemKey || dictItem.DictID != dictID {
		dictItems, err := s.repo.GetByDictID(dictID)
		if err == nil {
			for _, item := range dictItems {
				if item.ID != id && item.ItemKey == itemKey {
					s.log.Error("字典项键已存在", "dict_id", dictID, "item_key", itemKey)
					return errors.New("字典项键已被使用")
				}
			}
		}
	}

	// 更新字典项
	dictItem.DictID = dictID
	dictItem.ItemKey = itemKey
	dictItem.ItemValue = itemValue
	dictItem.Sort = sort
	dictItem.Description = description
	dictItem.Status = status
	dictItem.UpdatedBy = updatedBy

	// 保存更新
	if err := s.repo.Update(dictItem); err != nil {
		s.log.Error("更新字典项失败", "error", err)
		return errors.New("更新字典项失败")
	}

	s.log.Info("更新字典项成功", "id", id)
	return nil
}

// DeleteDictItem 删除字典项
func (s *DictItemService) DeleteDictItem(id string) error {
	// 获取字典项
	_, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取字典项失败", "id", id, "error", err)
		return errors.New("字典项不存在")
	}

	// 删除字典项
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除字典项失败", "error", err)
		return errors.New("删除字典项失败")
	}

	s.log.Info("删除字典项成功", "id", id)
	return nil
}

// GetDictItemList 获取字典项列表
func (s *DictItemService) GetDictItemList(page, pageSize int, filters map[string]interface{}) ([]*entity.DictItem, int64, error) {
	dictItems, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取字典项列表失败", "error", err)
		return nil, 0, errors.New("获取字典项列表失败")
	}
	return dictItems, total, nil
}

// GetActiveDictItemsByDictID 获取指定字典的所有启用项
func (s *DictItemService) GetActiveDictItemsByDictID(dictID string) ([]*entity.DictItem, error) {
	// 检查字典是否存在
	_, err := s.dictRepo.GetByID(dictID)
	if err != nil {
		s.log.Error("获取字典失败", "dict_id", dictID, "error", err)
		return nil, errors.New("字典不存在")
	}

	dictItems, err := s.repo.GetActiveByDictID(dictID)
	if err != nil {
		s.log.Error("获取启用的字典项列表失败", "dict_id", dictID, "error", err)
		return nil, errors.New("获取启用的字典项列表失败")
	}
	return dictItems, nil
}
