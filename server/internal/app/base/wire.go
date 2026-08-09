package base

import (
	"github.com/google/wire"
	baseapi "github.com/ix-pay/ixpay-pro/internal/app/base/api"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/domain/converter"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/monitor"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	repository "github.com/ix-pay/ixpay-pro/internal/persistence/base"
)

// 定义模块服务提供者集合
var ProviderSetBaseMonitor = wire.NewSet(
	// 监控层
	monitor.SetupSystemMonitor,
	monitor.SetupCacheMonitor,
	monitor.SetupDatabaseMonitor,
)

var ProviderSetBaseRepo = wire.NewSet(
	// 仓库层
	repository.NewAPIRepository,
	repository.NewMenuRepository,
	repository.NewLoginLogRepository,
	repository.NewRoleRepository,
	repository.NewUserRepository,
	repository.NewUserSettingRepository,
	repository.NewConfigRepository,
	repository.NewDictRepository,
	repository.NewDictItemRepository,
	repository.NewOperationLogRepository,
	repository.NewPermissionRuleRepository,
	repository.NewTaskExecutionLogRepository,
	repository.NewTaskRepository,
	repository.NewDepartmentRepository,
	repository.NewPositionRepository,
	repository.NewNoticeRepository,
	repository.NewNoticeReadRecordRepository,
	repository.NewOnlineUserRepository,
	repository.NewPermissionLogRepository,
	repository.NewNodeRepository,
)

var ProviderSetBaseService = wire.NewSet(
	// 服务层
	service.NewLoginLogService,
	service.NewAPIService,
	service.NewMenuService,
	service.NewRoleService,
	service.NewRolePermissionService,
	service.NewUserService,
	service.NewConfigService,
	service.NewDictService,
	service.NewDictItemService,
	service.NewOperationLogService,
	service.NewPermissionService,
	service.NewTaskExecutionLogService,
	service.NewTaskService,
	service.NewDepartmentService,
	service.NewPositionService,
	service.NewNoticeService,
	service.NewNoticeReadRecordService,
	service.NewOnlineUserService,
	service.NewPermissionLogService,
	service.NewNodeService,
)

var ProviderSetBaseConverter = wire.NewSet(
	// 转换器层
	converter.NewEntityConverter,
)

var ProviderSetBaseController = wire.NewSet(
	// 控制器层
	baseapi.NewAPIController,
	baseapi.NewMenuController,
	baseapi.NewLoginLogController,
	baseapi.NewRoleController,
	baseapi.NewUserController,
	baseapi.NewAuthController,
	baseapi.NewConfigController,
	baseapi.NewDictController,
	baseapi.NewOperationLogController,
	task.NewTaskFactory,
	baseapi.NewTaskController,
	baseapi.NewDepartmentController,
	baseapi.NewPositionController,
	baseapi.NewNoticeController,
	baseapi.NewOnlineUserController,
	baseapi.NewMonitorController,
	baseapi.NewPermissionLogController,
	baseapi.NewNodeController,
	baseapi.NewGatewayServiceHandler,
)
var ProviderSetBaseApp = wire.NewSet(
	// 应用层
	NewAppBase,
)
