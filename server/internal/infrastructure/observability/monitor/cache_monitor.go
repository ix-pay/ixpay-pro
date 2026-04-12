package monitor

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheStatus 缓存状态信息
type CacheStatus struct {
	Connected    bool          `json:"connected"`     // 连接状态
	RedisInfo    RedisInfo     `json:"redis_info"`    // Redis 服务器信息
	KeyStats     KeyStats      `json:"key_stats"`     // 键统计
	CommandStats []CommandStat `json:"command_stats"` // 命令统计
	HitRate      float64       `json:"hit_rate"`      // 缓存命中率
	Timestamp    time.Time     `json:"timestamp"`     // 采集时间
}

// RedisInfo Redis 服务器信息
type RedisInfo struct {
	Version          string `json:"version"`           // Redis 版本
	Mode             string `json:"mode"`              // 运行模式（standalone/cluster/sentinel）
	ConnectedClients int64  `json:"connected_clients"` // 已连接客户端数
	UsedMemoryMB     uint64 `json:"used_memory_mb"`    // 已使用内存 (MB)
	MaxMemoryMB      uint64 `json:"max_memory_mb"`     // 最大内存限制 (MB)
	UptimeSeconds    int64  `json:"uptime_seconds"`    // 运行时间 (秒)
	TotalCommands    int64  `json:"total_commands"`    // 总命令数
	TotalConnections int64  `json:"total_connections"` // 总连接数
	ExpiredKeys      int64  `json:"expired_keys"`      // 过期键数
	EvictedKeys      int64  `json:"evicted_keys"`      // 被驱逐键数
	KeyspaceHits     int64  `json:"keyspace_hits"`     // 命中次数
	KeyspaceMisses   int64  `json:"keyspace_misses"`   // 未命中次数
}

// KeyStats 键统计信息
type KeyStats struct {
	TotalKeys   int64            `json:"total_keys"`   // 总键数
	KeysByDB    map[string]int64 `json:"keys_by_db"`   // 按数据库分组的键数
	ExpiredKeys int64            `json:"expired_keys"` // 过期键数
	EvictedKeys int64            `json:"evicted_keys"` // 被驱逐键数
}

// CommandStat 命令统计信息
type CommandStat struct {
	Command   string `json:"command"`     // 命令名称
	Calls     int64  `json:"calls"`       // 调用次数
	UseTimeUs int64  `json:"use_time_us"` // 耗时 (微秒)
	AvgTimeUs int64  `json:"avg_time_us"` // 平均耗时 (微秒)
}

// CacheMonitor 缓存监控服务
type CacheMonitor struct {
	redisClient *redis.Client
	ctx         context.Context
}

// NewCacheMonitor 创建缓存监控服务实例
// 参数:
// - redisClient: Redis 客户端实例
func NewCacheMonitor(redisClient *redis.Client) *CacheMonitor {
	return &CacheMonitor{
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

// GetCacheStatus 获取缓存状态信息
// 返回 Redis 缓存的使用情况，包括连接状态、键统计、命中率等
func (m *CacheMonitor) GetCacheStatus() (*CacheStatus, error) {
	status := &CacheStatus{
		Timestamp: time.Now(),
		Connected: false,
	}

	// 检查 Redis 连接状态
	if err := m.checkConnection(); err != nil {
		// 连接失败，返回基本状态
		return status, nil
	}

	status.Connected = true

	// 获取 Redis 服务器信息
	if err := m.collectRedisInfo(status); err != nil {
		return nil, err
	}

	// 获取键统计信息
	if err := m.collectKeyStats(status); err != nil {
		return nil, err
	}

	// 获取命令统计信息
	if err := m.collectCommandStats(status); err != nil {
		return nil, err
	}

	// 计算缓存命中率
	m.calculateHitRate(status)

	return status, nil
}

// GetRedisKeys 获取 Redis 键列表（支持模式匹配）
// 参数:
// - pattern: 键名模式（如 "user:*"）
// - limit: 返回数量限制
// 返回:
// - []string: 键名列表
// - int64: 匹配的总键数
// - error: 错误信息
func (m *CacheMonitor) GetRedisKeys(pattern string, limit int64) ([]string, int64, error) {
	if err := m.checkConnection(); err != nil {
		return nil, 0, err
	}

	// 如果模式为空，使用默认模式
	if pattern == "" {
		pattern = "*"
	}

	// 使用 SCAN 命令代替 KEYS 命令，避免阻塞 Redis
	var keys []string
	var cursor uint64
	count := int64(0)

	for {
		// 扫描键
		result, nextCursor, err := m.redisClient.Scan(m.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, 0, err
		}

		keys = append(keys, result...)
		count += int64(len(result))

		// 检查是否达到限制
		if limit > 0 && count >= limit {
			keys = keys[:limit]
			break
		}

		// 检查是否扫描完成
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	totalCount := count
	if limit > 0 && totalCount > limit {
		totalCount = limit
	}

	return keys, totalCount, nil
}

// checkConnection 检查 Redis 连接状态
func (m *CacheMonitor) checkConnection() error {
	_, err := m.redisClient.Ping(m.ctx).Result()
	return err
}

// collectRedisInfo 收集 Redis 服务器信息
func (m *CacheMonitor) collectRedisInfo(status *CacheStatus) error {
	info, err := m.redisClient.Info(m.ctx).Result()
	if err != nil {
		return err
	}

	// 解析 Redis INFO 命令返回的信息
	infoMap := parseRedisInfo(info)

	// 提取关键信息
	status.RedisInfo.Version = infoMap["redis_version"]
	status.RedisInfo.Mode = "standalone" // 默认为单机模式
	status.RedisInfo.ConnectedClients = parseInt64(infoMap["connected_clients"])
	status.RedisInfo.UsedMemoryMB = parseUint64(infoMap["used_memory"]) / 1024 / 1024
	status.RedisInfo.MaxMemoryMB = parseUint64(infoMap["maxmemory"]) / 1024 / 1024
	status.RedisInfo.UptimeSeconds = parseInt64(infoMap["uptime_in_seconds"])
	status.RedisInfo.TotalCommands = parseInt64(infoMap["total_commands_processed"])
	status.RedisInfo.TotalConnections = parseInt64(infoMap["total_connections_received"])
	status.RedisInfo.ExpiredKeys = parseInt64(infoMap["expired_keys"])
	status.RedisInfo.EvictedKeys = parseInt64(infoMap["evicted_keys"])
	status.RedisInfo.KeyspaceHits = parseInt64(infoMap["keyspace_hits"])
	status.RedisInfo.KeyspaceMisses = parseInt64(infoMap["keyspace_misses"])

	return nil
}

// collectKeyStats 收集键统计信息
func (m *CacheMonitor) collectKeyStats(status *CacheStatus) error {
	// 获取所有数据库的键数统计
	dbStats, err := m.redisClient.DBSize(m.ctx).Result()
	if err != nil {
		return err
	}

	status.KeyStats.TotalKeys = dbStats
	status.KeyStats.KeysByDB = make(map[string]int64)
	status.KeyStats.KeysByDB["db0"] = dbStats
	status.KeyStats.ExpiredKeys = status.RedisInfo.ExpiredKeys
	status.KeyStats.EvictedKeys = status.RedisInfo.EvictedKeys

	return nil
}

// collectCommandStats 收集命令统计信息
func (m *CacheMonitor) collectCommandStats(status *CacheStatus) error {
	// 获取命令统计
	cmdStats, err := m.redisClient.Info(m.ctx, "commandstats").Result()
	if err != nil {
		return err
	}

	// 解析命令统计信息
	lines := splitLines(cmdStats)
	for _, line := range lines {
		if len(line) == 0 || line[0] != '#' {
			continue
		}

		// 解析命令行：cmdstat_get:calls=10,usec=100,usec_per_call=10.00
		if len(line) > 9 && line[:9] == "cmdstat_" {
			stat := parseCommandStat(line)
			if stat != nil {
				status.CommandStats = append(status.CommandStats, *stat)
			}
		}
	}

	return nil
}

// calculateHitRate 计算缓存命中率
func (m *CacheMonitor) calculateHitRate(status *CacheStatus) {
	hits := status.RedisInfo.KeyspaceHits
	misses := status.RedisInfo.KeyspaceMisses

	if hits+misses > 0 {
		status.HitRate = float64(hits) / float64(hits+misses) * 100
	} else {
		status.HitRate = 0
	}
}

// parseRedisInfo 解析 Redis INFO 命令返回的信息
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := splitLines(info)

	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := splitByColon(line)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// parseCommandStat 解析命令统计信息
func parseCommandStat(line string) *CommandStat {
	// 格式：cmdstat_get:calls=10,usec=100,usec_per_call=10.00
	parts := splitByColon(line)
	if len(parts) < 2 {
		return nil
	}

	// 提取命令名称
	cmdName := parts[0][9:] // 去掉 "cmdstat_" 前缀

	// 解析统计信息
	stats := make(map[string]string)
	statParts := splitByComma(parts[1])
	for _, statPart := range statParts {
		kv := splitByEquals(statPart)
		if len(kv) == 2 {
			stats[kv[0]] = kv[1]
		}
	}

	calls := parseInt64(stats["calls"])
	useTime := parseInt64(stats["usec"])
	avgTime := int64(0)
	if calls > 0 {
		avgTime = useTime / calls
	}

	return &CommandStat{
		Command:   cmdName,
		Calls:     calls,
		UseTimeUs: useTime,
		AvgTimeUs: avgTime,
	}
}

// 辅助函数
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			if i > start {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitByColon(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func splitByComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func splitByEquals(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func parseInt64(s string) int64 {
	var result int64
	_, err := scanString(s, "%d", &result)
	if err != nil {
		return 0
	}
	return result
}

func parseUint64(s string) uint64 {
	var result uint64
	_, err := scanString(s, "%d", &result)
	if err != nil {
		return 0
	}
	return result
}

// scanString 简单的字符串解析函数
func scanString(s string, format string, dest interface{}) (int, error) {
	// 这里使用一个简单的实现，实际项目中可以使用 fmt.Sscanf
	switch d := dest.(type) {
	case *int64:
		var val int64
		n, err := scanInt64(s, &val)
		if err == nil {
			*d = val
		}
		return n, err
	case *uint64:
		var val uint64
		n, err := scanUint64(s, &val)
		if err == nil {
			*d = val
		}
		return n, err
	default:
		return 0, nil
	}
}

// scanInt64 解析 int64
func scanInt64(s string, dest *int64) (int, error) {
	var result int64 = 0
	var sign int64 = 1
	var i int

	// 跳过前导空格
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}

	// 处理符号
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}

	// 解析数字
	var parsed bool
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		result = result*10 + int64(s[i]-'0')
		i++
		parsed = true
	}

	if !parsed {
		return 0, nil
	}

	*dest = result * sign
	return i, nil
}

// scanUint64 解析 uint64
func scanUint64(s string, dest *uint64) (int, error) {
	var result uint64 = 0
	var i int

	// 跳过前导空格
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}

	// 解析数字
	var parsed bool
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		result = result*10 + uint64(s[i]-'0')
		i++
		parsed = true
	}

	if !parsed {
		return 0, nil
	}

	*dest = result
	return i, nil
}
