# 批量生成 Mock 文件的脚本
# 使用 mockgen 工具为所有 Repository 接口生成 Mock 实现

Set-Location -Path "d:\g\ixpay-pro\server"

# 检查 mockgen 是否安装
$mockgen = Get-Command mockgen -ErrorAction SilentlyContinue
if (-not $mockgen) {
    Write-Host "正在安装 mockgen..." -ForegroundColor Yellow
    go install github.com/golang/mock/mockgen@latest
}

# 定义 Repository 接口和对应的 Mock 输出文件
$repositories = @(
    @{ Source = "internal/app/base/domain/repository/role_repository.go"; Destination = "internal/app/base/domain/repository/mock/role_repository_mock.go"; Interface = "RoleRepository" },
    @{ Source = "internal/app/base/domain/repository/menu_repository.go"; Destination = "internal/app/base/domain/repository/mock/menu_repository_mock.go"; Interface = "MenuRepository" },
    @{ Source = "internal/app/base/domain/repository/apis_repository.go"; Destination = "internal/app/base/domain/repository/mock/apis_repository_mock.go"; Interface = "APIRepository" },
    @{ Source = "internal/app/base/domain/repository/btn_perm_repository.go"; Destination = "internal/app/base/domain/repository/mock/btn_perm_repository_mock.go"; Interface = "BtnPermRepository" },
    @{ Source = "internal/app/base/domain/repository/department_repository.go"; Destination = "internal/app/base/domain/repository/mock/department_repository_mock.go"; Interface = "DepartmentRepository" },
    @{ Source = "internal/app/base/domain/repository/position_repository.go"; Destination = "internal/app/base/domain/repository/mock/position_repository_mock.go"; Interface = "PositionRepository" },
    @{ Source = "internal/app/base/domain/repository/config_repository.go"; Destination = "internal/app/base/domain/repository/mock/config_repository_mock.go"; Interface = "ConfigRepository" },
    @{ Source = "internal/app/base/domain/repository/dict_repository.go"; Destination = "internal/app/base/domain/repository/mock/dict_repository_mock.go"; Interface = "DictRepository" },
    @{ Source = "internal/app/base/domain/repository/login_log_repository.go"; Destination = "internal/app/base/domain/repository/mock/login_log_repository_mock.go"; Interface = "LoginLogRepository" },
    @{ Source = "internal/app/base/domain/repository/operation_log_repository.go"; Destination = "internal/app/base/domain/repository/mock/operation_log_repository_mock.go"; Interface = "OperationLogRepository" },
    @{ Source = "internal/app/base/domain/repository/task_execution_log_repository.go"; Destination = "internal/app/base/domain/repository/mock/task_execution_log_repository_mock.go"; Interface = "TaskExecutionLogRepository" },
    @{ Source = "internal/app/base/domain/repository/notice_repository.go"; Destination = "internal/app/base/domain/repository/mock/notice_repository_mock.go"; Interface = "NoticeRepository" },
    @{ Source = "internal/app/base/domain/repository/online_user_repository.go"; Destination = "internal/app/base/domain/repository/mock/online_user_repository_mock.go"; Interface = "OnlineUserRepository" },
    @{ Source = "internal/app/base/domain/repository/notice_read_record_repository.go"; Destination = "internal/app/base/domain/repository/mock/notice_read_record_repository_mock.go"; Interface = "NoticeReadRecordRepository" },
    @{ Source = "internal/app/base/domain/repository/permission_rule_repository.go"; Destination = "internal/app/base/domain/repository/mock/permission_rule_repository_mock.go"; Interface = "PermissionRuleRepository" }
)

# 确保 mock 目录存在
$mockDir = "internal/app/base/domain/repository/mock"
if (-not (Test-Path $mockDir)) {
    New-Item -ItemType Directory -Path $mockDir | Out-Null
    Write-Host "已创建 mock 目录：$mockDir" -ForegroundColor Green
}

# 遍历每个 Repository 并生成 Mock 文件
foreach ($repo in $repositories) {
    $source = $repo.Source
    $destination = $repo.Destination
    $interface = $repo.Interface
    
    Write-Host "正在生成 $interface 的 Mock 文件..." -ForegroundColor Cyan
    
    # 构建 mockgen 命令
    $cmd = "mockgen -source=$source -destination=$destination -package=mocks"
    
    # 执行命令
    Invoke-Expression $cmd
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 成功生成：$destination" -ForegroundColor Green
    } else {
        Write-Host "✗ 生成失败：$destination" -ForegroundColor Red
        Write-Host "  错误信息：$LASTEXITCODE" -ForegroundColor Red
    }
}

Write-Host "`nMock 文件生成完成！" -ForegroundColor Green
Write-Host "生成的文件位于：$mockDir" -ForegroundColor Yellow
