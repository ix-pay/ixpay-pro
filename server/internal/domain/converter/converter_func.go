package converter

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
)

// LoginLogToListDTO 转换为列表 DTO（精简版）
func LoginLogToListDTO(log *entity.LoginLog) response.LoginLogListDTO {
	return response.LoginLogListDTO{
		ID:        log.ID,
		Username:  log.Username,
		IP:        log.LoginIP,
		Place:     log.LoginPlace,
		Result:    int(log.Result),
		CreatedAt: log.LoginTime.Format(time.RFC3339),
	}
}

// LoginLogToDetailDTO 转换为详情 DTO（完整版）
func LoginLogToDetailDTO(log *entity.LoginLog) response.LoginLogDetailDTO {
	return response.LoginLogDetailDTO{
		ID:        log.ID,
		UserID:    log.UserID,
		Username:  log.Username,
		IP:        log.LoginIP,
		Place:     log.LoginPlace,
		Device:    log.Device,
		Browser:   log.Browser,
		OS:        log.OS,
		Result:    int(log.Result),
		CreatedAt: log.LoginTime,
	}
}

// LoginLogToStatisticsDTO 转换为统计 DTO
func LoginLogToStatisticsDTO(log *entity.LoginLog) response.LoginLogStatisticsDTO {
	return response.LoginLogStatisticsDTO{
		Date:    log.LoginTime.Format("2006-01-02"),
		Total:   1,
		Success: int(log.Result),
		Failure: 1 - int(log.Result),
	}
}

// UserToSimpleDTO 转换为简单 DTO（列表使用）
func UserToSimpleDTO(user *entity.User) response.UserSimpleDTO {
	return response.UserSimpleDTO{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}
}

// UserToDetailDTO 转换为详情 DTO（详情使用）
func UserToDetailDTO(user *entity.User) response.UserDetailDTO {
	return response.UserDetailDTO{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Phone:        user.Phone,
		Avatar:       user.Avatar,
		Status:       user.Status,
		DepartmentID: user.DepartmentID,
		PositionID:   user.PositionID,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	}
}

// UserToSelectDTO 转换为下拉选项 DTO（选择器使用）
func UserToSelectDTO(user *entity.User) response.UserSelectDTO {
	return response.UserSelectDTO{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
	}
}

// PositionToDTO 转换为岗位 DTO（简单 Entity 也使用函数式转换器）
func PositionToDTO(position *entity.Position) response.PositionResponse {
	return response.PositionResponse{
		ID:          position.ID,
		Name:        position.Name,
		Sort:        position.Sort,
		Status:      position.Status,
		Description: position.Description,
		CreatedAt:   position.CreatedAt.Format(time.RFC3339),
	}
}

// MenuToDTO 转换为菜单 DTO
func MenuToDTO(menu *entity.Menu) response.MenuResponse {
	return response.MenuResponse{
		ID:          menu.ID,
		ParentID:    menu.ParentID,
		Path:        menu.Path,
		Name:        menu.Name,
		Component:   menu.Component,
		Title:       menu.Title,
		Icon:        menu.Icon,
		Hidden:      menu.Hidden,
		Sort:        menu.Sort,
		Status:      menu.Status,
		IsExt:       menu.IsExt,
		Redirect:    menu.Redirect,
		Permission:  menu.Permission,
		KeepAlive:   menu.KeepAlive,
		DefaultMenu: menu.DefaultMenu,
	}
}

// RoleToDTO 转换为角色 DTO
func RoleToDTO(role *entity.Role) response.RoleResponse {
	return response.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Sort:        role.Sort,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt.Format(time.RFC3339),
	}
}

// DepartmentToDTO 转换为部门 DTO
func DepartmentToDTO(dept *entity.Department) response.DepartmentResponse {
	return response.DepartmentResponse{
		ID:        dept.ID,
		Name:      dept.Name,
		ParentID:  dept.ParentID,
		Sort:      dept.Sort,
		Status:    dept.Status,
		Leader:    dept.LeaderID,
		CreatedAt: dept.CreatedAt.Format(time.RFC3339),
	}
}

// AbnormalLoginInfoToDTO 转换为异常登录信息 DTO
func AbnormalLoginInfoToDTO(info *entity.AbnormalLoginInfo) response.AbnormalLoginInfoDTO {
	return response.AbnormalLoginInfoDTO{
		IP:              info.IP,
		FailedCount:     info.FailedCount,
		LastFailedTime:  info.LastFailedTime.Format(time.RFC3339),
		Usernames:       info.Usernames,
		RiskLevel:       info.RiskLevel,
		RiskDescription: info.RiskDescription,
	}
}
