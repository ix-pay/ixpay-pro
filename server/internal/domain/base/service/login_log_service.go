package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// LoginLogService 登录日志服务实现
type LoginLogService struct {
	repo repo.LoginLogRepository
	log  logger.Logger
}

// NewLoginLogService 创建登录日志服务实例
func NewLoginLogService(repo repo.LoginLogRepository, log logger.Logger) *LoginLogService {
	return &LoginLogService{
		repo: repo,
		log:  log,
	}
}

// RecordLogin 记录登录日志
func (s *LoginLogService) RecordLogin(
	userID string,
	username, ip, place, device, browser, os, userAgent string,
	success bool,
	errorMsg string,
) error {
	result := entity.LoginResultSuccess
	if !success {
		result = entity.LoginResultFailed
	}

	log := &entity.LoginLog{
		UserID:     userID,
		Username:   username,
		LoginIP:    ip,
		LoginTime:  time.Now(),
		LoginPlace: place,
		Device:     device,
		Browser:    browser,
		OS:         os,
		Result:     result,
		ErrorMsg:   errorMsg,
		UserAgent:  userAgent,
	}

	if err := s.repo.Create(log); err != nil {
		s.log.Error("记录登录日志失败", "error", err, "username", username, "ip", ip, "result", result)
		return err
	}

	s.log.Info("记录登录日志成功", "user_id", userID, "username", username, "ip", ip, "result", result)
	return nil
}

// GetLoginLogByID 获取登录日志详情
func (s *LoginLogService) GetLoginLogByID(id string) (*entity.LoginLog, error) {
	log, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取登录日志失败", "error", err, "id", id)
		return nil, errors.New("登录日志不存在")
	}
	return log, nil
}

// GetLoginLogList 获取登录日志列表
func (s *LoginLogService) GetLoginLogList(page, pageSize int, filters map[string]interface{}) ([]*entity.LoginLog, int64, error) {
	s.log.Info("获取登录日志列表", "page", page, "pageSize", pageSize)
	list, total, err := s.repo.List(page, pageSize, filters)
	return list, total, err
}

// GetUserLoginLogs 获取用户登录日志列表
func (s *LoginLogService) GetUserLoginLogs(userID string, page, pageSize int) ([]*entity.LoginLog, int64, error) {
	s.log.Info("获取用户登录日志列表", "user_id", userID, "page", page, "pageSize", pageSize)
	list, total, err := s.repo.GetByUserID(userID, page, pageSize)
	return list, total, err
}

// GetStatistics 获取登录统计信息
func (s *LoginLogService) GetStatistics(startTime, endTime time.Time) (*entity.LoginStatistics, error) {
	s.log.Info("获取登录统计信息", "start_time", startTime, "end_time", endTime)
	return s.repo.GetStatistics(startTime, endTime)
}

// CheckAbnormalLogin 检测异常登录
// 检测规则：
// 1. 同一 IP 在 1 小时内失败超过 5 次 - 高风险
// 2. 同一 IP 在 1 小时内失败超过 3 次 - 中风险
// 3. 同一 IP 在 1 小时内失败超过 1 次 - 低风险
func (s *LoginLogService) CheckAbnormalLogin(ip string) (*entity.AbnormalLoginInfo, error) {
	s.log.Info("检测异常登录", "ip", ip)

	// 查询最近 1 小时内的失败登录记录
	failedLogs, err := s.repo.GetFailedByIP(ip, 1)
	if err != nil {
		s.log.Error("查询失败登录记录失败", "error", err, "ip", ip)
		return nil, err
	}

	if len(failedLogs) == 0 {
		return nil, nil // 没有异常
	}

	// 统计失败次数和尝试的用户名
	failedCount := int64(len(failedLogs))
	usernames := make(map[string]bool)
	var lastFailedTime time.Time

	for _, log := range failedLogs {
		usernames[log.Username] = true
		if log.LoginTime.After(lastFailedTime) {
			lastFailedTime = log.LoginTime
		}
	}

	usernameList := make([]string, 0, len(usernames))
	for username := range usernames {
		usernameList = append(usernameList, username)
	}

	// 判断风险等级
	riskLevel := "low"
	riskDescription := fmt.Sprintf("该 IP 在 1 小时内有 %d 次登录失败记录", failedCount)

	if failedCount >= 5 {
		riskLevel = "high"
		riskDescription = fmt.Sprintf("该 IP 在 1 小时内有 %d 次登录失败记录，可能存在暴力破解风险", failedCount)
	} else if failedCount >= 3 {
		riskLevel = "medium"
		riskDescription = fmt.Sprintf("该 IP 在 1 小时内有 %d 次登录失败记录，请密切关注", failedCount)
	}

	return &entity.AbnormalLoginInfo{
		IP:              ip,
		FailedCount:     failedCount,
		LastFailedTime:  lastFailedTime,
		Usernames:       usernameList,
		RiskLevel:       riskLevel,
		RiskDescription: riskDescription,
	}, nil
}

// GetAbnormalLogins 获取异常登录记录
// 返回所有在 1 小时内有失败记录的 IP 信息
func (s *LoginLogService) GetAbnormalLogins(page, pageSize int) ([]*entity.AbnormalLoginInfo, int64, error) {
	s.log.Info("获取异常登录记录", "page", page, "pageSize", pageSize)

	// 查询最近 1 小时内的所有失败登录记录
	since := time.Now().Add(-1 * time.Hour)
	filters := map[string]interface{}{
		"result":     entity.LoginResultFailed,
		"start_time": since,
	}

	failedLogs, _, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("查询失败登录记录失败", "error", err)
		return nil, 0, err
	}

	if len(failedLogs) == 0 {
		return []*entity.AbnormalLoginInfo{}, 0, nil
	}

	// 按 IP 分组统计
	ipMap := make(map[string]*entity.AbnormalLoginInfo)
	for _, log := range failedLogs {
		if info, exists := ipMap[log.LoginIP]; exists {
			info.FailedCount++
			if log.LoginTime.After(info.LastFailedTime) {
				info.LastFailedTime = log.LoginTime
			}
			// 添加用户名（去重）
			usernameExists := false
			for _, username := range info.Usernames {
				if username == log.Username {
					usernameExists = true
					break
				}
			}
			if !usernameExists {
				info.Usernames = append(info.Usernames, log.Username)
			}
		} else {
			ipMap[log.LoginIP] = &entity.AbnormalLoginInfo{
				IP:             log.LoginIP,
				FailedCount:    1,
				LastFailedTime: log.LoginTime,
				Usernames:      []string{log.Username},
			}
		}
	}

	// 转换为列表并设置风险等级
	result := make([]*entity.AbnormalLoginInfo, 0, len(ipMap))
	for _, info := range ipMap {
		if info.FailedCount >= 5 {
			info.RiskLevel = "high"
			info.RiskDescription = fmt.Sprintf("该 IP 在 1 小时内有 %d 次登录失败记录，可能存在暴力破解风险", info.FailedCount)
		} else if info.FailedCount >= 3 {
			info.RiskLevel = "medium"
			info.RiskDescription = fmt.Sprintf("该 IP 在 1 小时内有 %d 次登录失败记录，请密切关注", info.FailedCount)
		} else {
			info.RiskLevel = "low"
			info.RiskDescription = fmt.Sprintf("该 IP 在 1 小时内有 %d 次登录失败记录", info.FailedCount)
		}
		result = append(result, info)
	}

	// 按失败次数降序排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].FailedCount < result[j].FailedCount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	total := int64(len(result))

	// 分页
	if page > 0 && pageSize > 0 {
		start := (page - 1) * pageSize
		end := start + pageSize
		if start >= len(result) {
			return []*entity.AbnormalLoginInfo{}, total, nil
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, total, nil
}
