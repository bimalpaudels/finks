package monitor

import (
	"context"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// MetricsService handles server metrics collection
type MetricsService struct{}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

// GetMetrics collects comprehensive system metrics
func (ms *MetricsService) GetMetrics() (*ServerMetrics, error) {
	ctx := context.Background()
	timestamp := time.Now()

	// Collect all metrics concurrently for better performance
	cpuMetrics := ms.getCPUMetrics(ctx)
	memoryMetrics := ms.getMemoryMetrics(ctx)
	diskMetrics := ms.getDiskMetrics(ctx)
	networkMetrics := ms.getNetworkMetrics(ctx)
	processMetrics := ms.getProcessMetrics(ctx)
	loadMetrics := ms.getLoadMetrics(ctx)
	systemMetrics := ms.getSystemMetrics(ctx)

	return &ServerMetrics{
		Timestamp: timestamp,
		CPU:       cpuMetrics,
		Memory:    memoryMetrics,
		Disk:      diskMetrics,
		Network:   networkMetrics,
		Processes: processMetrics,
		Load:      loadMetrics,
		System:    systemMetrics,
	}, nil
}

func (ms *MetricsService) getCPUMetrics(ctx context.Context) CPUMetrics {
	// Get overall CPU usage
	cpuPercent, _ := cpu.PercentWithContext(ctx, time.Second, false)
	var usage float64
	if len(cpuPercent) > 0 {
		usage = cpuPercent[0]
	}

	// Get per-core CPU usage
	perCore, _ := cpu.PercentWithContext(ctx, time.Second, true)

	// Get detailed CPU times
	cpuTimes, _ := cpu.TimesWithContext(ctx, false)
	var user, system, idle, iowait float64
	if len(cpuTimes) > 0 {
		total := cpuTimes[0].User + cpuTimes[0].System + cpuTimes[0].Idle + cpuTimes[0].Iowait
		if total > 0 {
			user = (cpuTimes[0].User / total) * 100
			system = (cpuTimes[0].System / total) * 100
			idle = (cpuTimes[0].Idle / total) * 100
			iowait = (cpuTimes[0].Iowait / total) * 100
		}
	}

	// Get load averages
	loadAvg, _ := load.AvgWithContext(ctx)
	var load1, load5, load15 float64
	if loadAvg != nil {
		load1 = loadAvg.Load1
		load5 = loadAvg.Load5
		load15 = loadAvg.Load15
	}

	return CPUMetrics{
		Usage:     usage,
		User:      user,
		System:    system,
		Idle:      idle,
		IOWait:    iowait,
		PerCore:   perCore,
		LoadAvg1:  load1,
		LoadAvg5:  load5,
		LoadAvg15: load15,
	}
}

func (ms *MetricsService) getMemoryMetrics(ctx context.Context) MemoryMetrics {
	memStat, _ := mem.VirtualMemoryWithContext(ctx)
	swapStat, _ := mem.SwapMemoryWithContext(ctx)

	var memory MemoryMetrics
	if memStat != nil {
		memory.Total = memStat.Total
		memory.Available = memStat.Available
		memory.Used = memStat.Used
		memory.UsedPercent = memStat.UsedPercent
		memory.Cached = memStat.Cached
		memory.Buffers = memStat.Buffers
	}

	if swapStat != nil {
		memory.SwapTotal = swapStat.Total
		memory.SwapUsed = swapStat.Used
		memory.SwapPercent = swapStat.UsedPercent
	}

	return memory
}

func (ms *MetricsService) getDiskMetrics(ctx context.Context) DiskMetrics {
	diskStat, _ := disk.UsageWithContext(ctx, "/")
	
	var diskMetrics DiskMetrics
	if diskStat != nil {
		diskMetrics.Total = diskStat.Total
		diskMetrics.Used = diskStat.Used
		diskMetrics.Free = diskStat.Free
		diskMetrics.UsedPercent = diskStat.UsedPercent
		diskMetrics.InodesTotal = diskStat.InodesTotal
		diskMetrics.InodesUsed = diskStat.InodesUsed
	}

	// Get disk I/O stats
	ioStats, _ := disk.IOCountersWithContext(ctx)
	var readIOPS, writeIOPS uint64
	var readMBps, writeMBps float64
	
	for _, io := range ioStats {
		readIOPS += io.ReadCount
		writeIOPS += io.WriteCount
		readMBps += float64(io.ReadBytes) / 1024 / 1024
		writeMBps += float64(io.WriteBytes) / 1024 / 1024
	}

	diskMetrics.ReadIOPS = readIOPS
	diskMetrics.WriteIOPS = writeIOPS
	diskMetrics.ReadMBps = readMBps
	diskMetrics.WriteMBps = writeMBps

	return diskMetrics
}

func (ms *MetricsService) getNetworkMetrics(ctx context.Context) NetworkMetrics {
	netStats, _ := net.IOCountersWithContext(ctx, false)
	connStats, _ := net.ConnectionsWithContext(ctx, "inet")

	var network NetworkMetrics
	if len(netStats) > 0 {
		network.BytesSent = netStats[0].BytesSent
		network.BytesRecv = netStats[0].BytesRecv
		network.PacketsSent = netStats[0].PacketsSent
		network.PacketsRecv = netStats[0].PacketsRecv
		network.Errin = netStats[0].Errin
		network.Errout = netStats[0].Errout
		network.Dropin = netStats[0].Dropin
		network.Dropout = netStats[0].Dropout
		
		// Calculate approximate throughput (this is cumulative, would need delta for real-time)
		network.ThroughputIn = float64(netStats[0].BytesRecv) / 1024 / 1024
		network.ThroughputOut = float64(netStats[0].BytesSent) / 1024 / 1024
	}

	network.Connections = uint64(len(connStats))

	return network
}

func (ms *MetricsService) getProcessMetrics(ctx context.Context) ProcessMetrics {
	processes, _ := process.ProcessesWithContext(ctx)
	
	var totalProcs, runningProcs, sleepingProcs, zombieProcs uint64
	var processInfos []ProcessInfo

	for _, p := range processes {
		totalProcs++
		
		status, _ := p.StatusWithContext(ctx)
		if len(status) > 0 {
			switch status[0] {
			case "R":
				runningProcs++
			case "S":
				sleepingProcs++
			case "Z":
				zombieProcs++
			}
		}

		// Get detailed process info for top processes
		name, _ := p.NameWithContext(ctx)
		cpuPercent, _ := p.CPUPercentWithContext(ctx)
		memInfo, _ := p.MemoryInfoWithContext(ctx)
		memPercent, _ := p.MemoryPercentWithContext(ctx)

		var memUsage float64
		if memInfo != nil {
			memUsage = float64(memInfo.RSS) / 1024 / 1024 // Convert to MB
		}

		if cpuPercent > 0 || memUsage > 0 {
			processInfos = append(processInfos, ProcessInfo{
				PID:        p.Pid,
				Name:       name,
				CPUUsage:   cpuPercent,
				MemUsage:   memUsage,
				MemPercent: float64(memPercent),
			})
		}
	}

	// Sort by CPU usage and get top 5
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].CPUUsage > processInfos[j].CPUUsage
	})
	topCPU := make([]ProcessInfo, 0, 5)
	for i := 0; i < len(processInfos) && i < 5; i++ {
		topCPU = append(topCPU, processInfos[i])
	}

	// Sort by memory usage and get top 5
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].MemUsage > processInfos[j].MemUsage
	})
	topMemory := make([]ProcessInfo, 0, 5)
	for i := 0; i < len(processInfos) && i < 5; i++ {
		topMemory = append(topMemory, processInfos[i])
	}

	return ProcessMetrics{
		Total:     totalProcs,
		Running:   runningProcs,
		Sleeping:  sleepingProcs,
		Zombie:    zombieProcs,
		TopCPU:    topCPU,
		TopMemory: topMemory,
	}
}

func (ms *MetricsService) getLoadMetrics(ctx context.Context) LoadMetrics {
	loadAvg, _ := load.AvgWithContext(ctx)
	
	var load1, load5, load15 float64
	if loadAvg != nil {
		load1 = loadAvg.Load1
		load5 = loadAvg.Load5
		load15 = loadAvg.Load15
	}

	return LoadMetrics{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}
}

func (ms *MetricsService) getSystemMetrics(ctx context.Context) SystemMetrics {
	hostInfo, _ := host.InfoWithContext(ctx)
	
	var system SystemMetrics
	system.NumCPU = runtime.NumCPU()
	
	if hostInfo != nil {
		system.Uptime = float64(hostInfo.Uptime)
		system.BootTime = hostInfo.BootTime
		system.Hostname = hostInfo.Hostname
		system.Platform = hostInfo.Platform
		system.KernelVersion = hostInfo.KernelVersion
	}

	// Try to get CPU temperature (may not be available on all systems)
	temps, _ := host.SensorsTemperaturesWithContext(ctx)
	if len(temps) > 0 {
		system.Temperature = temps[0].Temperature
	}

	return system
}