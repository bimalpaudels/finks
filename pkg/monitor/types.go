package monitor

import "time"

// ServerMetrics represents comprehensive system metrics
type ServerMetrics struct {
	Timestamp time.Time    `json:"timestamp"`
	CPU       CPUMetrics   `json:"cpu"`
	Memory    MemoryMetrics `json:"memory"`
	Disk      DiskMetrics  `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	Processes ProcessMetrics `json:"processes"`
	Load      LoadMetrics  `json:"load"`
	System    SystemMetrics `json:"system"`
}

// CPUMetrics represents CPU usage and breakdown
type CPUMetrics struct {
	Usage    float64   `json:"usage"`      // Overall CPU usage percentage
	User     float64   `json:"user"`       // User time percentage
	System   float64   `json:"system"`     // System time percentage
	Idle     float64   `json:"idle"`       // Idle time percentage
	IOWait   float64   `json:"iowait"`     // IO wait time percentage
	PerCore  []float64 `json:"per_core"`   // Per-core usage percentages
	LoadAvg1 float64   `json:"load_avg_1"` // 1-minute load average
	LoadAvg5 float64   `json:"load_avg_5"` // 5-minute load average
	LoadAvg15 float64  `json:"load_avg_15"` // 15-minute load average
}

// MemoryMetrics represents memory usage details
type MemoryMetrics struct {
	Total      uint64  `json:"total"`       // Total memory in bytes
	Available  uint64  `json:"available"`   // Available memory in bytes
	Used       uint64  `json:"used"`        // Used memory in bytes
	UsedPercent float64 `json:"used_percent"` // Used memory percentage
	Cached     uint64  `json:"cached"`      // Cached memory in bytes
	Buffers    uint64  `json:"buffers"`     // Buffer memory in bytes
	SwapTotal  uint64  `json:"swap_total"`  // Total swap in bytes
	SwapUsed   uint64  `json:"swap_used"`   // Used swap in bytes
	SwapPercent float64 `json:"swap_percent"` // Swap usage percentage
}

// DiskMetrics represents disk usage and I/O statistics
type DiskMetrics struct {
	Total       uint64  `json:"total"`        // Total disk space in bytes
	Used        uint64  `json:"used"`         // Used disk space in bytes
	Free        uint64  `json:"free"`         // Free disk space in bytes
	UsedPercent float64 `json:"used_percent"` // Used disk percentage
	InodesTotal uint64  `json:"inodes_total"` // Total inodes
	InodesUsed  uint64  `json:"inodes_used"`  // Used inodes
	ReadIOPS    uint64  `json:"read_iops"`    // Read IOPS
	WriteIOPS   uint64  `json:"write_iops"`   // Write IOPS
	ReadMBps    float64 `json:"read_mbps"`    // Read MB/s
	WriteMBps   float64 `json:"write_mbps"`   // Write MB/s
}

// NetworkMetrics represents network statistics
type NetworkMetrics struct {
	BytesSent      uint64  `json:"bytes_sent"`       // Total bytes sent
	BytesRecv      uint64  `json:"bytes_recv"`       // Total bytes received
	PacketsSent    uint64  `json:"packets_sent"`     // Total packets sent
	PacketsRecv    uint64  `json:"packets_recv"`     // Total packets received
	Errin          uint64  `json:"errin"`            // Input errors
	Errout         uint64  `json:"errout"`           // Output errors
	Dropin         uint64  `json:"dropin"`           // Dropped input packets
	Dropout        uint64  `json:"dropout"`          // Dropped output packets
	Connections    uint64  `json:"connections"`      // Active network connections
	ThroughputIn   float64 `json:"throughput_in"`    // Current input throughput (MB/s)
	ThroughputOut  float64 `json:"throughput_out"`   // Current output throughput (MB/s)
}

// ProcessMetrics represents process and system activity
type ProcessMetrics struct {
	Total       uint64        `json:"total"`        // Total number of processes
	Running     uint64        `json:"running"`      // Number of running processes
	Sleeping    uint64        `json:"sleeping"`     // Number of sleeping processes
	Zombie      uint64        `json:"zombie"`       // Number of zombie processes
	TopCPU      []ProcessInfo `json:"top_cpu"`      // Top 5 processes by CPU
	TopMemory   []ProcessInfo `json:"top_memory"`   // Top 5 processes by memory
}

// ProcessInfo represents individual process information
type ProcessInfo struct {
	PID     int32   `json:"pid"`      // Process ID
	Name    string  `json:"name"`     // Process name
	CPUUsage float64 `json:"cpu_usage"` // CPU usage percentage
	MemUsage float64 `json:"mem_usage"` // Memory usage in MB
	MemPercent float64 `json:"mem_percent"` // Memory usage percentage
}

// LoadMetrics represents system load information
type LoadMetrics struct {
	Load1  float64 `json:"load1"`  // 1-minute load average
	Load5  float64 `json:"load5"`  // 5-minute load average
	Load15 float64 `json:"load15"` // 15-minute load average
}

// SystemMetrics represents general system information
type SystemMetrics struct {
	Uptime       float64 `json:"uptime"`        // System uptime in seconds
	BootTime     uint64  `json:"boot_time"`     // Boot time timestamp
	NumCPU       int     `json:"num_cpu"`       // Number of CPU cores
	Hostname     string  `json:"hostname"`      // System hostname
	Platform     string  `json:"platform"`      // Operating system platform
	KernelVersion string `json:"kernel_version"` // Kernel version
	Temperature   float64 `json:"temperature"`   // CPU temperature if available
}

// HealthCheck represents a health check result
type HealthCheck struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "healthy", "unhealthy", "unknown"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// ServerStatus represents overall server health
type ServerStatus struct {
	Status      string        `json:"status"` // "healthy", "degraded", "unhealthy"
	Uptime      time.Duration `json:"uptime"`
	HealthChecks []HealthCheck `json:"health_checks"`
	LastUpdated time.Time     `json:"last_updated"`
}