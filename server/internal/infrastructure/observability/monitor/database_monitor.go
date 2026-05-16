package monitor

import (
	"context"
	"database/sql"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// DatabaseStatus 数据库状态信息
type DatabaseStatus struct {
	Connected   bool        `json:"connected"`    // 连接状态
	PoolStats   PoolStats   `json:"pool_stats"`   // 连接池统计
	QueryStats  QueryStats  `json:"query_stats"`  // SQL 查询统计
	SlowQueries []SlowQuery `json:"slow_queries"` // 慢查询列表
	ServerInfo  ServerInfo  `json:"server_info"`  // 服务器信息
	Timestamp   time.Time   `json:"timestamp"`    // 采集时间
}

// PoolStats 连接池统计信息
type PoolStats struct {
	MaxOpenConns      int   `json:"max_open_conns"`      // 最大打开连接数
	OpenConnections   int   `json:"open_connections"`    // 当前打开的连接数
	InUse             int   `json:"in_use"`              // 正在使用的连接数
	Idle              int   `json:"idle"`                // 空闲连接数
	WaitCount         int64 `json:"wait_count"`          // 等待连接数
	WaitDurationMs    int64 `json:"wait_duration_ms"`    // 等待总时长 (毫秒)
	MaxIdleClosed     int64 `json:"max_idle_closed"`     // 因超过空闲时间关闭的连接数
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"` // 因超过生命周期关闭的连接数
}

// QueryStats SQL 查询统计信息
type QueryStats struct {
	TotalQueries   int64   `json:"total_queries"`     // 总查询数
	ActiveQueries  int64   `json:"active_queries"`    // 活跃查询数
	AvgQueryTimeMs float64 `json:"avg_query_time_ms"` // 平均查询时间 (毫秒)
	MaxQueryTimeMs float64 `json:"max_query_time_ms"` // 最大查询时间 (毫秒)
	SlowQueryCount int64   `json:"slow_query_count"`  // 慢查询数量
}

// SlowQuery 慢查询信息
type SlowQuery struct {
	QueryID    int64     `json:"query_id"`    // 查询 ID
	Query      string    `json:"query"`       // SQL 查询语句
	Database   string    `json:"database"`    // 数据库名
	User       string    `json:"user"`        // 执行用户
	Client     string    `json:"client"`      // 客户端地址
	Start      time.Time `json:"start"`       // 开始时间
	DurationMs float64   `json:"duration_ms"` // 执行时长 (毫秒)
	WaitEvent  string    `json:"wait_event"`  // 等待事件
	State      string    `json:"state"`       // 查询状态
}

// ServerInfo 数据库服务器信息
type ServerInfo struct {
	Version          string `json:"version"`           // PostgreSQL 版本
	Uptime           string `json:"uptime"`            // 运行时间
	NumBackends      int64  `json:"num_backends"`      // 当前连接数
	MaxConnections   int64  `json:"max_connections"`   // 最大连接数
	DatabaseSize     int64  `json:"database_size"`     // 数据库大小 (字节)
	TransactionCount int64  `json:"transaction_count"` // 事务数
	TupReturned      int64  `json:"tup_returned"`      // 返回的行数
	TupFetched       int64  `json:"tup_fetched"`       // 获取的行数
	TupInserted      int64  `json:"tup_inserted"`      // 插入的行数
	TupUpdated       int64  `json:"tup_updated"`       // 更新的行数
	TupDeleted       int64  `json:"tup_deleted"`       // 删除的行数
}

// DatabaseMonitor 数据库监控服务
type DatabaseMonitor struct {
	db                 *database.PostgresDB
	sqlDB              *sql.DB
	ctx                context.Context
	slowQueryThreshold time.Duration // 慢查询阈值
}

// NewDatabaseMonitor 创建数据库监控服务实例
// 参数:
// - db: PostgreSQL 数据库实例
// - slowQueryThreshold: 慢查询阈值（毫秒）
func NewDatabaseMonitor(db *database.PostgresDB, slowQueryThreshold int64) *DatabaseMonitor {
	// 从 GORM 获取底层的 sql.DB 实例
	sqlDB, err := db.DB.DB()
	if err != nil {
		// 如果获取失败，使用 nil，后续操作会检查
		sqlDB = nil
	}

	return &DatabaseMonitor{
		db:                 db,
		sqlDB:              sqlDB,
		ctx:                context.Background(),
		slowQueryThreshold: time.Duration(slowQueryThreshold) * time.Millisecond,
	}
}

// GetDatabaseStatus 获取数据库状态信息
// 返回数据库连接池状态、查询统计和慢查询信息
func (m *DatabaseMonitor) GetDatabaseStatus() (*DatabaseStatus, error) {
	status := &DatabaseStatus{
		Timestamp: time.Now(),
		Connected: false,
	}

	// 检查数据库连接状态
	if err := m.checkConnection(); err != nil {
		// 连接失败，返回基本状态
		return status, nil
	}

	status.Connected = true

	// 获取连接池统计信息
	if err := m.collectPoolStats(status); err != nil {
		return nil, err
	}

	// 获取查询统计信息
	if err := m.collectQueryStats(status); err != nil {
		return nil, err
	}

	// 获取慢查询信息
	if err := m.collectSlowQueries(status); err != nil {
		// 慢查询收集失败不影响整体状态
		status.SlowQueries = []SlowQuery{}
	}

	// 获取服务器信息
	if err := m.collectServerInfo(status); err != nil {
		return nil, err
	}

	return status, nil
}

// GetSlowQueries 获取慢查询日志
// 参数:
// - limit: 返回数量限制
// - minDuration: 最小执行时长（毫秒）
// 返回:
// - []SlowQuery: 慢查询列表
// - error: 错误信息
func (m *DatabaseMonitor) GetSlowQueries(limit int, minDuration int64) ([]SlowQuery, error) {
	if err := m.checkConnection(); err != nil {
		return nil, err
	}

	// 查询 pg_stat_statements 获取慢查询
	query := `
		SELECT 
			queryid,
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			max_exec_time,
			rows
		FROM pg_stat_statements
		WHERE mean_exec_time > $1
		ORDER BY mean_exec_time DESC
		LIMIT $2
	`

	rows, err := m.db.Raw(query, minDuration, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slowQueries []SlowQuery
	for rows.Next() {
		var sq SlowQuery
		var totalExecTime, meanExecTime, maxExecTime float64
		var calls, rowsAffected int64

		if err := rows.Scan(&sq.QueryID, &sq.Query, &calls, &totalExecTime, &meanExecTime, &maxExecTime, &rowsAffected); err != nil {
			continue
		}

		sq.DurationMs = meanExecTime
		slowQueries = append(slowQueries, sq)
	}

	return slowQueries, nil
}

// checkConnection 检查数据库连接状态
func (m *DatabaseMonitor) checkConnection() error {
	if m.sqlDB == nil {
		return nil
	}
	return m.sqlDB.PingContext(m.ctx)
}

// collectPoolStats 收集连接池统计信息
func (m *DatabaseMonitor) collectPoolStats(status *DatabaseStatus) error {
	if m.sqlDB == nil {
		return nil
	}

	stats := m.sqlDB.Stats()

	status.PoolStats.MaxOpenConns = stats.MaxOpenConnections
	status.PoolStats.OpenConnections = stats.OpenConnections
	status.PoolStats.InUse = stats.InUse
	status.PoolStats.Idle = stats.Idle
	status.PoolStats.WaitCount = stats.WaitCount
	status.PoolStats.WaitDurationMs = stats.WaitDuration.Milliseconds()
	status.PoolStats.MaxIdleClosed = stats.MaxIdleClosed
	status.PoolStats.MaxLifetimeClosed = stats.MaxLifetimeClosed

	return nil
}

// collectQueryStats 收集查询统计信息
func (m *DatabaseMonitor) collectQueryStats(status *DatabaseStatus) error {
	// 查询当前活跃查询数
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE state = 'active') as active,
			COALESCE(AVG(EXTRACT(EPOCH FROM (now() - query_start))) * 1000, 0) as avg_time,
			COALESCE(MAX(EXTRACT(EPOCH FROM (now() - query_start))) * 1000, 0) as max_time
		FROM pg_stat_activity
		WHERE state IS NOT NULL
	`

	type QueryStatsResult struct {
		Total   int64
		Active  int64
		AvgTime float64
		MaxTime float64
	}

	var result QueryStatsResult
	err := m.db.Raw(query).Scan(&result).Error
	if err != nil {
		return err
	}

	totalQueries := result.Total
	activeQueries := result.Active
	avgTime := result.AvgTime
	maxTime := result.MaxTime

	status.QueryStats.TotalQueries = totalQueries
	status.QueryStats.ActiveQueries = activeQueries
	status.QueryStats.AvgQueryTimeMs = avgTime
	status.QueryStats.MaxQueryTimeMs = maxTime

	// 查询慢查询数量
	slowQueryCount, err := m.getSlowQueryCount()
	if err != nil {
		return err
	}
	status.QueryStats.SlowQueryCount = slowQueryCount

	return nil
}

// collectSlowQueries 收集慢查询信息
func (m *DatabaseMonitor) collectSlowQueries(status *DatabaseStatus) error {
	// 查询当前正在执行的慢查询
	query := `
		SELECT 
			pid,
			query,
			datname,
			usename,
			client_addr,
			query_start,
			EXTRACT(EPOCH FROM (now() - query_start)) * 1000 as duration_ms,
			wait_event_type || '.' || wait_event as wait_event,
			state
		FROM pg_stat_activity
		WHERE state != 'idle'
			AND EXTRACT(EPOCH FROM (now() - query_start)) * 1000 > $1
		ORDER BY duration_ms DESC
		LIMIT 20
	`

	rows, err := m.db.Raw(query, m.slowQueryThreshold.Milliseconds()).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var slowQueries []SlowQuery
	for rows.Next() {
		var sq SlowQuery
		var clientAddr sql.NullString
		var waitEvent sql.NullString
		var startTime time.Time

		if err := rows.Scan(&sq.QueryID, &sq.Query, &sq.Database, &sq.User, &clientAddr, &startTime, &sq.DurationMs, &waitEvent, &sq.State); err != nil {
			continue
		}

		sq.Start = startTime
		if clientAddr.Valid {
			sq.Client = clientAddr.String
		}
		if waitEvent.Valid {
			sq.WaitEvent = waitEvent.String
		}

		slowQueries = append(slowQueries, sq)
	}

	status.SlowQueries = slowQueries
	return nil
}

// collectServerInfo 收集服务器信息
func (m *DatabaseMonitor) collectServerInfo(status *DatabaseStatus) error {
	// 获取 PostgreSQL 版本
	versionQuery := `SELECT version()`
	var version string
	if err := m.db.Raw(versionQuery).Scan(&version).Error; err != nil {
		return err
	}
	status.ServerInfo.Version = version

	// 获取运行时间
	uptimeQuery := `
		SELECT date_trunc('second', now() - pg_postmaster_start_time())::text
		FROM pg_postmaster_start_time()
	`
	var uptime string
	if err := m.db.Raw(uptimeQuery).Scan(&uptime).Error; err != nil {
		return err
	}
	status.ServerInfo.Uptime = uptime

	// 获取连接数信息
	connQuery := `
		SELECT 
			COUNT(*) as num_backends,
			(SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max_connections
		FROM pg_stat_activity
	`
	type ConnResult struct {
		NumBackends    int64
		MaxConnections int64
	}
	var connResult ConnResult
	if err := m.db.Raw(connQuery).Scan(&connResult).Error; err != nil {
		return err
	}
	status.ServerInfo.NumBackends = connResult.NumBackends
	status.ServerInfo.MaxConnections = connResult.MaxConnections

	// 获取数据库大小
	dbNameQuery := `SELECT current_database()`
	var dbName string
	if err := m.db.Raw(dbNameQuery).Scan(&dbName).Error; err != nil {
		return err
	}

	sizeQuery := `SELECT pg_database_size($1)`
	if err := m.db.Raw(sizeQuery, dbName).Scan(&status.ServerInfo.DatabaseSize).Error; err != nil {
		return err
	}

	// 获取统计信息
	statsQuery := `
		SELECT 
			SUM(xact_commit) as transaction_count,
			SUM(tup_returned) as tup_returned,
			SUM(tup_fetched) as tup_fetched,
			SUM(tup_inserted) as tup_inserted,
			SUM(tup_updated) as tup_updated,
			SUM(tup_deleted) as tup_deleted
		FROM pg_stat_database
		WHERE datname = $1
	`
	type StatsResult struct {
		TransactionCount int64
		TupReturned      int64
		TupFetched       int64
		TupInserted      int64
		TupUpdated       int64
		TupDeleted       int64
	}
	var statsResult StatsResult
	if err := m.db.Raw(statsQuery, dbName).Scan(&statsResult).Error; err != nil {
		return err
	}
	status.ServerInfo.TransactionCount = statsResult.TransactionCount
	status.ServerInfo.TupReturned = statsResult.TupReturned
	status.ServerInfo.TupFetched = statsResult.TupFetched
	status.ServerInfo.TupInserted = statsResult.TupInserted
	status.ServerInfo.TupUpdated = statsResult.TupUpdated
	status.ServerInfo.TupDeleted = statsResult.TupDeleted

	return nil
}

// getSlowQueryCount 获取慢查询数量
func (m *DatabaseMonitor) getSlowQueryCount() (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM pg_stat_activity
		WHERE state != 'idle'
			AND EXTRACT(EPOCH FROM (now() - query_start)) * 1000 > $1
	`

	var count int64
	err := m.db.Raw(query, m.slowQueryThreshold.Milliseconds()).Scan(&count).Error
	return count, err
}
