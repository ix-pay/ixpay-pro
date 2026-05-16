package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// ConfigService 实现配置领域服务接口
type ConfigService struct {
	repo repo.ConfigRepository
	log  logger.Logger
}

// NewConfigService 创建配置服务实例
func NewConfigService(repo repo.ConfigRepository, log logger.Logger) *ConfigService {
	return &ConfigService{
		repo: repo,
		log:  log,
	}
}

// GetConfigByKey 根据配置键获取配置
func (s *ConfigService) GetConfigByKey(configKey string) (*entity.Config, error) {
	config, err := s.repo.GetByKey(configKey)
	if err != nil {
		s.log.Error("获取配置失败", "config_key", configKey, "error", err)
		return nil, errors.New("配置不存在")
	}
	return config, nil
}

// GetConfigByID 根据 ID 获取配置
func (s *ConfigService) GetConfigByID(id int64) (*entity.Config, error) {
	config, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取配置失败", "id", id, "error", err)
		return nil, errors.New("配置不存在")
	}
	return config, nil
}

// CreateConfig 创建配置
func (s *ConfigService) CreateConfig(configKey, configValue string, configType int, description string, status int, createdBy int64) (*entity.Config, error) {
	// 检查配置键是否已存在
	existingConfig, err := s.repo.GetByKey(configKey)
	if err == nil {
		if existingConfig.ID != 0 {
			s.log.Error("配置键已存在", "config_key", configKey)
			return nil, errors.New("配置键已存在")
		}
	}

	// 创建配置
	config := &entity.Config{
		ConfigKey:   configKey,
		ConfigValue: configValue,
		ConfigType:  configType,
		Description: description,
		Status:      status,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	// 保存配置
	if err := s.repo.Create(config); err != nil {
		s.log.Error("创建配置失败", "error", err)
		return nil, errors.New("创建配置失败")
	}

	s.log.Info("创建配置成功", "config_key", configKey)
	return config, nil
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(id int64, configKey, configValue string, configType int, description string, status int, updatedBy int64) error {
	// 获取配置
	config, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取配置失败", "id", id, "error", err)
		return errors.New("配置不存在")
	}

	// 检查配置键是否已被其他配置使用
	if config.ConfigKey != configKey {
		existingConfig, err := s.repo.GetByKey(configKey)
		if err == nil && existingConfig.ID != id {
			s.log.Error("配置键已存在", "config_key", configKey)
			return errors.New("配置键已被使用")
		}
	}

	// 更新配置
	config.ConfigKey = configKey
	config.ConfigValue = configValue
	config.ConfigType = configType
	config.Description = description
	config.Status = status
	config.UpdatedBy = updatedBy

	// 保存更新
	if err := s.repo.Update(config); err != nil {
		s.log.Error("更新配置失败", "error", err)
		return errors.New("更新配置失败")
	}

	s.log.Info("更新配置成功", "config_key", configKey)
	return nil
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(id int64) error {
	// 获取配置
	_, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取配置失败", "id", id, "error", err)
		return errors.New("配置不存在")
	}

	// 删除配置
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除配置失败", "error", err)
		return errors.New("删除配置失败")
	}

	s.log.Info("删除配置成功", "id", id)
	return nil
}

// GetConfigList 获取配置列表
func (s *ConfigService) GetConfigList(page, pageSize int, filters map[string]interface{}) ([]*entity.Config, int64, error) {
	configs, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取配置列表失败", "error", err)
		return nil, 0, errors.New("获取配置列表失败")
	}
	return configs, total, nil
}

// GetAllActiveConfigs 获取所有启用的配置
func (s *ConfigService) GetAllActiveConfigs() ([]*entity.Config, error) {
	configs, err := s.repo.GetAllActive()
	if err != nil {
		s.log.Error("获取启用的配置列表失败", "error", err)
		return nil, errors.New("获取启用的配置列表失败")
	}
	return configs, nil
}
