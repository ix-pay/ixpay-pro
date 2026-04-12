// monitor 包提供系统监控功能
// 包括系统资源监控、缓存监控和数据库监控
package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemStatus 系统状态信息
type SystemStatus struct {
	CPU       CPUStatus    `json:"cpu"`       // CPU 使用状态
	Memory    MemoryStatus `json:"memory"`    // 内存使用状态
	Disk      DiskStatus   `json:"disk"`      // 磁盘使用状态
	Load      LoadStatus   `json:"load"`      // 系统负载状态
	Timestamp time.Time    `json:"timestamp"` // 采集时间
}

// CPUStatus CPU 使用状态
type CPUStatus struct {
	UsagePercent float64   `json:"usage_percent"` // CPU 使用率百分比
	Cores        int       `json:"cores"`         // CPU 核心数
	PerCPUUsage  []float64 `json:"per_cpu_usage"` // 每个 CPU 核心的使用率
}

// MemoryStatus 内存使用状态
type MemoryStatus struct {
	TotalMB      uint64  `json:"total_mb"`      // 总内存 (MB)
	UsedMB       uint64  `json:"used_mb"`       // 已使用内存 (MB)
	FreeMB       uint64  `json:"free_mb"`       // 空闲内存 (MB)
	UsagePercent float64 `json:"usage_percent"` // 内存使用率百分比
	AvailableMB  uint64  `json:"available_mb"`  // 可用内存 (MB)
	UsedPercent  float64 `json:"used_percent"`  // 已使用百分比（与 usage_percent 相同）
	BuffersMB    uint64  `json:"buffers_mb"`    // 缓冲区内存 (MB)
	CachedMB     uint64  `json:"cached_mb"`     // 缓存内存 (MB)
}

// DiskStatus 磁盘使用状态
type DiskStatus struct {
	TotalGB      uint64  `json:"total_gb"`       // 总磁盘空间 (GB)
	UsedGB       uint64  `json:"used_gb"`        // 已使用磁盘空间 (GB)
	FreeGB       uint64  `json:"free_gb"`        // 空闲磁盘空间 (GB)
	UsagePercent float64 `json:"usage_percent"`  // 磁盘使用率百分比
	ReadBytes    uint64  `json:"read_bytes"`     // 读取字节数
	WriteBytes   uint64  `json:"write_bytes"`    // 写入字节数
	ReadCount    uint64  `json:"read_count"`     // 读取次数
	WriteCount   uint64  `json:"write_count"`    // 写入次数
	IoReadBytes  uint64  `json:"io_read_bytes"`  // IO 读取字节数
	IoWriteBytes uint64  `json:"io_write_bytes"` // IO 写入字节数
}

// LoadStatus 系统负载状态
type LoadStatus struct {
	Load1  float64 `json:"load1"`  // 1 分钟平均负载
	Load5  float64 `json:"load5"`  // 5 分钟平均负载
	Load15 float64 `json:"load15"` // 15 分钟平均负载
}

// SystemMonitor 系统监控服务
type SystemMonitor struct{}

// NewSystemMonitor 创建系统监控服务实例
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{}
}

// GetSystemStatus 获取系统状态信息
// 返回系统资源使用情况，包括 CPU、内存、磁盘和负载
func (m *SystemMonitor) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{
		Timestamp: time.Now(),
	}

	// 获取 CPU 使用率
	if err := m.collectCPUStatus(status); err != nil {
		return nil, err
	}

	// 获取内存使用情况
	if err := m.collectMemoryStatus(status); err != nil {
		return nil, err
	}

	// 获取磁盘使用情况
	if err := m.collectDiskStatus(status); err != nil {
		return nil, err
	}

	// 获取系统负载
	if err := m.collectLoadStatus(status); err != nil {
		return nil, err
	}

	return status, nil
}

// collectCPUStatus 收集 CPU 状态信息
func (m *SystemMonitor) collectCPUStatus(status *SystemStatus) error {
	// 获取 CPU 使用率百分比（1 秒内的平均值）
	percent, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil {
		return err
	}

	if len(percent) > 0 {
		status.CPU.UsagePercent = percent[0]
	}

	// 获取 CPU 核心数
	cores, err := cpu.Counts(true)
	if err != nil {
		return err
	}
	status.CPU.Cores = cores

	// 获取每个 CPU 核心的使用率
	perCPU, err := cpu.Percent(100*time.Millisecond, true)
	if err != nil {
		return err
	}
	status.CPU.PerCPUUsage = perCPU

	return nil
}

// collectMemoryStatus 收集内存状态信息
func (m *SystemMonitor) collectMemoryStatus(status *SystemStatus) error {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	status.Memory.TotalMB = vmStat.Total / 1024 / 1024
	status.Memory.UsedMB = vmStat.Used / 1024 / 1024
	status.Memory.FreeMB = vmStat.Free / 1024 / 1024
	status.Memory.UsagePercent = vmStat.UsedPercent
	status.Memory.UsedPercent = vmStat.UsedPercent
	status.Memory.AvailableMB = vmStat.Available / 1024 / 1024
	status.Memory.BuffersMB = vmStat.Buffers / 1024 / 1024
	status.Memory.CachedMB = vmStat.Cached / 1024 / 1024

	return nil
}

// collectDiskStatus 收集磁盘状态信息
func (m *SystemMonitor) collectDiskStatus(status *SystemStatus) error {
	// 获取根分区使用情况
	partitionStat, err := disk.Usage("/")
	if err != nil {
		// Windows 系统使用 C 盘
		partitionStat, err = disk.Usage("C:/")
		if err != nil {
			return err
		}
	}

	status.Disk.TotalGB = partitionStat.Total / 1024 / 1024 / 1024
	status.Disk.UsedGB = partitionStat.Used / 1024 / 1024 / 1024
	status.Disk.FreeGB = partitionStat.Free / 1024 / 1024 / 1024
	status.Disk.UsagePercent = partitionStat.UsedPercent

	// 获取磁盘 IO 统计
	ioStat, err := disk.IOCounters()
	if err != nil {
		return err
	}

	// 累加所有磁盘的 IO 统计
	for _, io := range ioStat {
		status.Disk.ReadBytes += io.ReadBytes
		status.Disk.WriteBytes += io.WriteBytes
		status.Disk.ReadCount += io.ReadCount
		status.Disk.WriteCount += io.WriteCount
		status.Disk.IoReadBytes += io.ReadBytes
		status.Disk.IoWriteBytes += io.WriteBytes
	}

	return nil
}

// collectLoadStatus 收集系统负载信息
func (m *SystemMonitor) collectLoadStatus(status *SystemStatus) error {
	loadStat, err := load.Avg()
	if err != nil {
		return err
	}

	status.Load.Load1 = loadStat.Load1
	status.Load.Load5 = loadStat.Load5
	status.Load.Load15 = loadStat.Load15

	return nil
}
