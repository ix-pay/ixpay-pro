// monitor 包提供系统监控功能
// 包括系统资源监控、缓存监控和数据库监控
package monitor

import (
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// truncateFloat 向下保留2位小数
func truncateFloat(val float64) float64 {
	return math.Trunc(val*100) / 100
}

// SystemStatus 系统状态信息
type SystemStatus struct {
	OSInfo    OSInfo        `json:"os_info"`   // 操作系统信息
	CPU       CPUStatus     `json:"cpu"`       // CPU 使用状态
	Memory    MemoryStatus  `json:"memory"`    // 内存使用状态
	Disk      []DiskStatus  `json:"disk"`      // 磁盘使用状态（支持多磁盘）
	Network   NetworkStatus `json:"network"`   // 网络使用状态
	Load      LoadStatus    `json:"load"`      // 系统负载状态
	Processes ProcessStatus `json:"processes"` // 进程信息
	Timestamp time.Time     `json:"timestamp"` // 采集时间
}

// OSInfo 操作系统信息
type OSInfo struct {
	OS       string `json:"os"`       // 操作系统类型：windows/linux/darwin
	Hostname string `json:"hostname"` // 主机名
	Platform string `json:"platform"` // 平台名称
	Kernel   string `json:"kernel"`   // 内核版本
	Uptime   uint64 `json:"uptime"`   // 系统运行时间（秒）
}

// CPUStatus CPU 使用状态
type CPUStatus struct {
	UsagePercent float64   `json:"usage_percent"` // CPU 使用率百分比
	Cores        int       `json:"cores"`         // CPU 核心数
	PerCPUUsage  []float64 `json:"per_cpu_usage"` // 每个 CPU 核心的使用率
	Frequency    float64   `json:"frequency"`     // CPU 频率（MHz）
}

// MemoryStatus 内存使用状态
type MemoryStatus struct {
	Total        uint64  `json:"total"`         // 总内存（字节）
	Used         uint64  `json:"used"`          // 已使用内存（字节）
	Free         uint64  `json:"free"`          // 空闲内存（字节）
	Available    uint64  `json:"available"`     // 可用内存（字节）
	UsagePercent float64 `json:"usage_percent"` // 内存使用率百分比
	Buffers      uint64  `json:"buffers"`       // 缓冲区内存（字节）
	Cached       uint64  `json:"cached"`        // 缓存内存（字节）
	SwapTotal    uint64  `json:"swap_total"`    // Swap 总空间（字节）
	SwapUsed     uint64  `json:"swap_used"`     // Swap 已使用（字节）
	SwapFree     uint64  `json:"swap_free"`     // Swap 空闲（字节）
}

// DiskStatus 磁盘使用状态
type DiskStatus struct {
	Device       string  `json:"device"`        // 设备名称
	Mountpoint   string  `json:"mountpoint"`    // 挂载点
	Fstype       string  `json:"fstype"`        // 文件系统类型
	Total        uint64  `json:"total"`         // 总磁盘空间（字节）
	Used         uint64  `json:"used"`          // 已使用磁盘空间（字节）
	Free         uint64  `json:"free"`          // 空闲磁盘空间（字节）
	UsagePercent float64 `json:"usage_percent"` // 磁盘使用率百分比
}

// NetworkStatus 网络使用状态
type NetworkStatus struct {
	BytesSent   uint64 `json:"bytes_sent"`   // 发送字节数
	BytesRecv   uint64 `json:"bytes_recv"`   // 接收字节数
	PacketsSent uint64 `json:"packets_sent"` // 发送数据包数
	PacketsRecv uint64 `json:"packets_recv"` // 接收数据包数
	Errin       uint64 `json:"errin"`        // 接收错误数
	Errout      uint64 `json:"errout"`       // 发送错误数
	Dropin      uint64 `json:"dropin"`       // 接收丢弃数
	Dropout     uint64 `json:"dropout"`      // 发送丢弃数
}

// LoadStatus 系统负载状态
type LoadStatus struct {
	Load1       float64 `json:"load1"`        // 1 分钟平均负载（或 CPU 使用率）
	Load5       float64 `json:"load5"`        // 5 分钟平均负载（或 CPU 使用率）
	Load15      float64 `json:"load15"`       // 15 分钟平均负载（或 CPU 使用率）
	IsSimulated bool    `json:"is_simulated"` // 是否为模拟值（Windows）
	CPUPercent  float64 `json:"cpu_percent"`  // 当前 CPU 使用率（Windows 时有效）
}

// ProcessStatus 进程信息
type ProcessStatus struct {
	Total     int           `json:"total"`      // 进程总数
	TopCPU    []ProcessInfo `json:"top_cpu"`    // CPU 使用率最高的进程
	TopMemory []ProcessInfo `json:"top_memory"` // 内存使用率最高的进程
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID        int32   `json:"pid"`         // 进程 ID
	Name       string  `json:"name"`        // 进程名称
	CPUPercent float64 `json:"cpu_percent"` // CPU 使用率
	MemPercent float64 `json:"mem_percent"` // 内存使用率
}

// SystemMonitor 系统监控服务
type SystemMonitor struct {
	lastCPUTimes []cpu.TimesStat
	lastCPUTime  time.Time
}

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

	// 获取操作系统信息
	if err := m.collectOSInfo(status); err != nil {
		// OS 信息获取失败不影响整体
		status.OSInfo.OS = runtime.GOOS
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
		// 磁盘获取失败不影响整体
		status.Disk = []DiskStatus{}
	}

	// 获取网络使用情况
	if err := m.collectNetworkStatus(status); err != nil {
		// 网络获取失败不影响整体
		status.Network = NetworkStatus{}
	}

	// 获取系统负载
	if err := m.collectLoadStatus(status); err != nil {
		// 负载获取失败不影响整体
		status.Load = LoadStatus{IsSimulated: true}
	}

	// 获取进程信息
	if err := m.collectProcessStatus(status); err != nil {
		// 进程获取失败不影响整体
		status.Processes = ProcessStatus{Total: 0}
	}

	return status, nil
}

// collectOSInfo 收集操作系统信息
func (m *SystemMonitor) collectOSInfo(status *SystemStatus) error {
	info, err := host.Info()
	if err != nil {
		return err
	}

	status.OSInfo.OS = info.Platform
	status.OSInfo.Hostname = info.Hostname
	status.OSInfo.Platform = info.Platform + " " + info.PlatformVersion
	status.OSInfo.Kernel = info.KernelVersion
	status.OSInfo.Uptime = info.Uptime

	return nil
}

// collectCPUStatus 收集 CPU 状态信息
func (m *SystemMonitor) collectCPUStatus(status *SystemStatus) error {
	// 获取 CPU 使用率百分比（500ms 内的平均值，提高准确性）
	percent, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return err
	}

	if len(percent) > 0 {
		status.CPU.UsagePercent = truncateFloat(percent[0])
	}

	// 获取 CPU 核心数
	cores, err := cpu.Counts(true)
	if err != nil {
		return err
	}
	status.CPU.Cores = cores

	// 获取每个 CPU 核心的使用率
	perCPU, err := cpu.Percent(0, true)
	if err == nil && len(perCPU) > 0 {
		// 截断每个核心的使用率到2位小数
		truncated := make([]float64, len(perCPU))
		for i, v := range perCPU {
			truncated[i] = truncateFloat(v)
		}
		status.CPU.PerCPUUsage = truncated
	}

	// 获取 CPU 频率（实时频率，不是基准频率）
	frequency, err := cpu.Info()
	if err == nil && len(frequency) > 0 {
		// 使用 cpu.Percent 来获取实际运行频率
		// gopsutil 的 cpu.Info 返回的是最大/基准频率
		// 这里我们使用 MHz 字段作为参考
		status.CPU.Frequency = truncateFloat(frequency[0].Mhz)
	}

	return nil
}

// collectMemoryStatus 收集内存状态信息
func (m *SystemMonitor) collectMemoryStatus(status *SystemStatus) error {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	status.Memory.Total = vmStat.Total
	status.Memory.Used = vmStat.Used
	status.Memory.Free = vmStat.Free
	status.Memory.UsagePercent = truncateFloat(vmStat.UsedPercent)
	status.Memory.Available = vmStat.Available
	status.Memory.Buffers = vmStat.Buffers
	status.Memory.Cached = vmStat.Cached

	// 获取 Swap 信息（Linux 特有，Windows 会返回空）
	swapStat, err := mem.SwapMemory()
	if err == nil && swapStat != nil {
		status.Memory.SwapTotal = swapStat.Total
		status.Memory.SwapUsed = swapStat.Used
		status.Memory.SwapFree = swapStat.Free
	}

	return nil
}

// collectDiskStatus 收集磁盘状态信息
func (m *SystemMonitor) collectDiskStatus(status *SystemStatus) error {
	// 获取所有分区使用情况
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	var diskStatuses []DiskStatus
	for _, partition := range partitions {
		// 跳过特殊文件系统
		if shouldSkipFilesystem(partition.Fstype, runtime.GOOS) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskStatuses = append(diskStatuses, DiskStatus{
			Device:       partition.Device,
			Mountpoint:   partition.Mountpoint,
			Fstype:       partition.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: truncateFloat(usage.UsedPercent),
		})
	}

	status.Disk = diskStatuses
	return nil
}

// shouldSkipFilesystem 判断是否应该跳过该文件系统
func shouldSkipFilesystem(fstype string, goos string) bool {
	// Windows 跳过特殊文件系统
	if goos == "windows" {
		skipList := []string{"procfs", "sysfs", "tmpfs", "devtmpfs", "cgroup", "cgroup2"}
		for _, skip := range skipList {
			if fstype == skip {
				return true
			}
		}
	}

	// Linux 跳过特殊文件系统
	if goos == "linux" {
		skipList := []string{"procfs", "sysfs", "tmpfs", "devtmpfs", "cgroup", "cgroup2", "devpts", "securityfs", "pstore", "bpf", "tracefs", "debugfs", "fusectl", "configfs", "hugetlbfs", "mqueue", "autofs"}
		for _, skip := range skipList {
			if fstype == skip {
				return true
			}
		}
	}

	return false
}

// collectNetworkStatus 收集网络状态信息
func (m *SystemMonitor) collectNetworkStatus(status *SystemStatus) error {
	// 获取所有网络接口的 IO 统计
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	var total NetworkStatus
	for _, io := range ioCounters {
		// 跳过回环接口
		if io.Name == "lo" || io.Name == "Loopback" {
			continue
		}

		total.BytesSent += io.BytesSent
		total.BytesRecv += io.BytesRecv
		total.PacketsSent += io.PacketsSent
		total.PacketsRecv += io.PacketsRecv
		total.Errin += io.Errin
		total.Errout += io.Errout
		total.Dropin += io.Dropin
		total.Dropout += io.Dropout
	}

	status.Network = total
	return nil
}

// collectLoadStatus 收集系统负载信息
func (m *SystemMonitor) collectLoadStatus(status *SystemStatus) error {
	// Windows 系统不支持 load.Avg()，使用 CPU 使用率模拟
	if runtime.GOOS == "windows" {
		// 使用当前 CPU 使用率作为负载指标
		status.Load.Load1 = truncateFloat(status.CPU.UsagePercent)
		status.Load.Load5 = truncateFloat(status.CPU.UsagePercent)
		status.Load.Load15 = truncateFloat(status.CPU.UsagePercent)
		status.Load.IsSimulated = true
		status.Load.CPUPercent = truncateFloat(status.CPU.UsagePercent)
		return nil
	}

	// Linux/Darwin 系统使用原生 load average
	loadStat, err := load.Avg()
	if err != nil {
		// 如果获取失败，使用 CPU 使用率模拟
		status.Load.Load1 = truncateFloat(status.CPU.UsagePercent)
		status.Load.Load5 = truncateFloat(status.CPU.UsagePercent)
		status.Load.Load15 = truncateFloat(status.CPU.UsagePercent)
		status.Load.IsSimulated = true
		status.Load.CPUPercent = truncateFloat(status.CPU.UsagePercent)
		return nil
	}

	status.Load.Load1 = truncateFloat(loadStat.Load1)
	status.Load.Load5 = truncateFloat(loadStat.Load5)
	status.Load.Load15 = truncateFloat(loadStat.Load15)
	status.Load.IsSimulated = false
	status.Load.CPUPercent = 0

	return nil
}

// procInfo 内部进程信息结构
type procInfo struct {
	pid        int32
	name       string
	cpuPercent float64
	memPercent float64
}

// collectProcessStatus 收集进程状态信息
func (m *SystemMonitor) collectProcessStatus(status *SystemStatus) error {
	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	status.Processes.Total = len(processes)

	var procInfos []procInfo
	for _, p := range processes {
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()

		procInfos = append(procInfos, procInfo{
			pid:        p.Pid,
			name:       name,
			cpuPercent: cpuPercent,
			memPercent: float64(memPercent),
		})
	}

	// 按 CPU 使用率排序
	status.Processes.TopCPU = getTopProcesses(procInfos, 5, func(a, b procInfo) bool {
		return a.cpuPercent > b.cpuPercent
	})

	// 按内存使用率排序
	status.Processes.TopMemory = getTopProcesses(procInfos, 5, func(a, b procInfo) bool {
		return a.memPercent > b.memPercent
	})

	return nil
}

// getTopProcesses 获取 Top N 进程
func getTopProcesses(procs []procInfo, n int, less func(a, b procInfo) bool) []ProcessInfo {
	// 简单冒泡排序取前 N 个
	for i := 0; i < n && i < len(procs); i++ {
		for j := i + 1; j < len(procs); j++ {
			if less(procs[j], procs[i]) {
				procs[i], procs[j] = procs[j], procs[i]
			}
		}
	}

	count := n
	if len(procs) < n {
		count = len(procs)
	}

	var result []ProcessInfo
	for i := 0; i < count; i++ {
		result = append(result, ProcessInfo{
			PID:        procs[i].pid,
			Name:       procs[i].name,
			CPUPercent: truncateFloat(procs[i].cpuPercent),
			MemPercent: truncateFloat(procs[i].memPercent),
		})
	}

	return result
}
