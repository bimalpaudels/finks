package monitor

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// Unicode characters for bars and indicators
const (
	BarFull     = "█"
	BarHigh     = "▓"
	BarMedium   = "▒"
	BarLow      = "░"
	BarEmpty    = "·"
	ArrowUp     = "↗"
	ArrowDown   = "↘"
	ArrowRight  = "→"
	Bullet      = "•"
	CheckMark   = "✓"
	Warning     = "⚠"
	Critical    = "✗"
)

// DisplayMetrics renders comprehensive metrics with visual enhancements
func DisplayMetrics(metrics *ServerMetrics) {
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
	
	// Header
	fmt.Printf("%s%s🦜 Finks System Monitor - %s%s\n",
		ColorBold, ColorCyan,
		metrics.Timestamp.Format("2006-01-02 15:04:05"),
		ColorReset)
	fmt.Printf("%s%s%s\n\n", ColorDim, strings.Repeat("─", 80), ColorReset)
	
	// System Overview
	displaySystemOverview(&metrics.System, &metrics.Load)
	
	// CPU Metrics
	displayCPUMetrics(&metrics.CPU)
	
	// Memory Metrics
	displayMemoryMetrics(&metrics.Memory)
	
	// Disk Metrics
	displayDiskMetrics(&metrics.Disk)
	
	// Network Metrics
	displayNetworkMetrics(&metrics.Network)
	
	// Process Metrics
	displayProcessMetrics(&metrics.Processes)
}

func displaySystemOverview(system *SystemMetrics, load *LoadMetrics) {
	fmt.Printf("%s%s 🖥️  SYSTEM OVERVIEW%s\n", ColorBold, ColorBlue, ColorReset)
	
	uptimeHours := system.Uptime / 3600
	uptimeDays := int(uptimeHours / 24)
	uptimeRemainHours := int(uptimeHours) % 24
	
	fmt.Printf("  %sHostname:%s %s  %sUptime:%s %dd %dh  %sCPU Cores:%s %d  %sPlatform:%s %s\n",
		ColorCyan, ColorReset, system.Hostname,
		ColorCyan, ColorReset, uptimeDays, uptimeRemainHours,
		ColorCyan, ColorReset, system.NumCPU,
		ColorCyan, ColorReset, system.Platform)
	
	// Load averages with color coding
	loadColor1 := getLoadColor(load.Load1, system.NumCPU)
	loadColor5 := getLoadColor(load.Load5, system.NumCPU)
	loadColor15 := getLoadColor(load.Load15, system.NumCPU)
	
	fmt.Printf("  %sLoad Average:%s %s%.2f%s %s%.2f%s %s%.2f%s",
		ColorCyan, ColorReset,
		loadColor1, load.Load1, ColorReset,
		loadColor5, load.Load5, ColorReset,
		loadColor15, load.Load15, ColorReset)
	
	if system.Temperature > 0 {
		tempColor := getTempColor(system.Temperature)
		fmt.Printf("  %sTemp:%s %s%.1f°C%s",
			ColorCyan, ColorReset,
			tempColor, system.Temperature, ColorReset)
	}
	fmt.Println()
	fmt.Println()
}

func displayCPUMetrics(cpu *CPUMetrics) {
	fmt.Printf("%s%s ⚡ CPU USAGE%s\n", ColorBold, ColorYellow, ColorReset)
	
	// Overall CPU usage with bar
	cpuColor := getPercentageColor(cpu.Usage)
	bar := createPercentageBar(cpu.Usage, 30)
	fmt.Printf("  %sOverall:%s %s%5.1f%%%s %s\n",
		ColorCyan, ColorReset,
		cpuColor, cpu.Usage, ColorReset, bar)
	
	// CPU breakdown
	fmt.Printf("  %sBreakdown:%s User %s%.1f%%%s  System %s%.1f%%%s  IOWait %s%.1f%%%s  Idle %s%.1f%%%s\n",
		ColorCyan, ColorReset,
		ColorGreen, cpu.User, ColorReset,
		ColorRed, cpu.System, ColorReset,
		ColorYellow, cpu.IOWait, ColorReset,
		ColorDim, cpu.Idle, ColorReset)
	
	// Per-core usage (show first 8 cores to avoid overwhelming output)
	if len(cpu.PerCore) > 0 {
		fmt.Printf("  %sPer Core:%s ", ColorCyan, ColorReset)
		maxCores := len(cpu.PerCore)
		if maxCores > 8 {
			maxCores = 8
		}
		for i := 0; i < maxCores; i++ {
			coreColor := getPercentageColor(cpu.PerCore[i])
			fmt.Printf("%s%4.0f%%%s ", coreColor, cpu.PerCore[i], ColorReset)
		}
		if len(cpu.PerCore) > 8 {
			fmt.Printf("... (+%d more)", len(cpu.PerCore)-8)
		}
		fmt.Println()
	}
	fmt.Println()
}

func displayMemoryMetrics(memory *MemoryMetrics) {
	fmt.Printf("%s%s 💾 MEMORY%s\n", ColorBold, ColorPurple, ColorReset)
	
	// Memory usage
	memColor := getPercentageColor(memory.UsedPercent)
	memBar := createPercentageBar(memory.UsedPercent, 30)
	fmt.Printf("  %sRAM:%s %s%5.1f%%%s %s (%s used / %s total)\n",
		ColorCyan, ColorReset,
		memColor, memory.UsedPercent, ColorReset, memBar,
		formatBytes(memory.Used), formatBytes(memory.Total))
	
	// Memory details
	fmt.Printf("  %sDetails:%s Available %s  Cached %s  Buffers %s\n",
		ColorCyan, ColorReset,
		formatBytes(memory.Available),
		formatBytes(memory.Cached),
		formatBytes(memory.Buffers))
	
	// Swap usage
	if memory.SwapTotal > 0 {
		swapColor := getPercentageColor(memory.SwapPercent)
		swapBar := createPercentageBar(memory.SwapPercent, 20)
		fmt.Printf("  %sSwap:%s %s%5.1f%%%s %s (%s used / %s total)\n",
			ColorCyan, ColorReset,
			swapColor, memory.SwapPercent, ColorReset, swapBar,
			formatBytes(memory.SwapUsed), formatBytes(memory.SwapTotal))
	}
	fmt.Println()
}

func displayDiskMetrics(disk *DiskMetrics) {
	fmt.Printf("%s%s 💿 STORAGE%s\n", ColorBold, ColorGreen, ColorReset)
	
	// Disk usage
	diskColor := getPercentageColor(disk.UsedPercent)
	diskBar := createPercentageBar(disk.UsedPercent, 30)
	fmt.Printf("  %sUsage:%s %s%5.1f%%%s %s (%s used / %s total)\n",
		ColorCyan, ColorReset,
		diskColor, disk.UsedPercent, ColorReset, diskBar,
		formatBytes(disk.Used), formatBytes(disk.Total))
	
	// Inode usage
	if disk.InodesTotal > 0 {
		inodePercent := float64(disk.InodesUsed) / float64(disk.InodesTotal) * 100
		inodeColor := getPercentageColor(inodePercent)
		fmt.Printf("  %sInodes:%s %s%5.1f%%%s (%d used / %d total)\n",
			ColorCyan, ColorReset,
			inodeColor, inodePercent, ColorReset,
			disk.InodesUsed, disk.InodesTotal)
	}
	
	// I/O stats
	fmt.Printf("  %sI/O:%s Read %s%d IOPS%s %.1f MB/s  Write %s%d IOPS%s %.1f MB/s\n",
		ColorCyan, ColorReset,
		ColorGreen, disk.ReadIOPS, ColorReset, disk.ReadMBps,
		ColorRed, disk.WriteIOPS, ColorReset, disk.WriteMBps)
	fmt.Println()
}

func displayNetworkMetrics(network *NetworkMetrics) {
	fmt.Printf("%s%s 🌐 NETWORK%s\n", ColorBold, ColorCyan, ColorReset)
	
	// Network throughput
	fmt.Printf("  %sThroughput:%s ↓ %s%.1f MB/s%s  ↑ %s%.1f MB/s%s  %sConnections:%s %d\n",
		ColorCyan, ColorReset,
		ColorGreen, network.ThroughputIn, ColorReset,
		ColorYellow, network.ThroughputOut, ColorReset,
		ColorCyan, ColorReset, network.Connections)
	
	// Packet stats
	fmt.Printf("  %sPackets:%s Received %s%s  Sent %s%s\n",
		ColorCyan, ColorReset,
		formatNumber(network.PacketsRecv), ColorReset,
		formatNumber(network.PacketsSent), ColorReset)
	
	// Error stats
	if network.Errin > 0 || network.Errout > 0 || network.Dropin > 0 || network.Dropout > 0 {
		fmt.Printf("  %sErrors:%s ", ColorCyan, ColorReset)
		if network.Errin > 0 {
			fmt.Printf("%sRx Errors:%s %s%d%s ", ColorRed, ColorReset, ColorRed, network.Errin, ColorReset)
		}
		if network.Errout > 0 {
			fmt.Printf("%sTx Errors:%s %s%d%s ", ColorRed, ColorReset, ColorRed, network.Errout, ColorReset)
		}
		if network.Dropin > 0 {
			fmt.Printf("%sRx Dropped:%s %s%d%s ", ColorYellow, ColorReset, ColorYellow, network.Dropin, ColorReset)
		}
		if network.Dropout > 0 {
			fmt.Printf("%sTx Dropped:%s %s%d%s ", ColorYellow, ColorReset, ColorYellow, network.Dropout, ColorReset)
		}
		fmt.Println()
	}
	fmt.Println()
}

func displayProcessMetrics(processes *ProcessMetrics) {
	fmt.Printf("%s%s ⚙️  PROCESSES%s\n", ColorBold, ColorWhite, ColorReset)
	
	// Process counts
	fmt.Printf("  %sTotal:%s %d  %sRunning:%s %s%d%s  %sSleeping:%s %d  %sZombie:%s %s%d%s\n",
		ColorCyan, ColorReset, processes.Total,
		ColorCyan, ColorReset, ColorGreen, processes.Running, ColorReset,
		ColorCyan, ColorReset, processes.Sleeping,
		ColorCyan, ColorReset, getZombieColor(processes.Zombie), processes.Zombie, ColorReset)
	
	// Top CPU processes
	if len(processes.TopCPU) > 0 {
		fmt.Printf("  %sTop CPU:%s\n", ColorCyan, ColorReset)
		for i, proc := range processes.TopCPU {
			if i >= 3 { break } // Show top 3
			cpuColor := getPercentageColor(proc.CPUUsage)
			fmt.Printf("    %s%d.%s %s%-16s%s %s%5.1f%%%s CPU  %7.1f MB\n",
				ColorDim, i+1, ColorReset,
				ColorWhite, truncateString(proc.Name, 16), ColorReset,
				cpuColor, proc.CPUUsage, ColorReset,
				proc.MemUsage)
		}
	}
	
	// Top Memory processes
	if len(processes.TopMemory) > 0 {
		fmt.Printf("  %sTop Memory:%s\n", ColorCyan, ColorReset)
		for i, proc := range processes.TopMemory {
			if i >= 3 { break } // Show top 3
			memColor := getPercentageColor(proc.MemPercent)
			fmt.Printf("    %s%d.%s %s%-16s%s %5.1f%% CPU  %s%7.1f MB%s\n",
				ColorDim, i+1, ColorReset,
				ColorWhite, truncateString(proc.Name, 16), ColorReset,
				proc.CPUUsage,
				memColor, proc.MemUsage, ColorReset)
		}
	}
	fmt.Println()
}

// Helper functions for colors and formatting

func getPercentageColor(percent float64) string {
	switch {
	case percent >= 90:
		return ColorRed
	case percent >= 70:
		return ColorYellow
	case percent >= 50:
		return ColorGreen
	default:
		return ColorCyan
	}
}

func getLoadColor(load float64, numCPU int) string {
	ratio := load / float64(numCPU)
	switch {
	case ratio >= 1.0:
		return ColorRed
	case ratio >= 0.7:
		return ColorYellow
	default:
		return ColorGreen
	}
}

func getTempColor(temp float64) string {
	switch {
	case temp >= 80:
		return ColorRed
	case temp >= 70:
		return ColorYellow
	default:
		return ColorGreen
	}
}

func getZombieColor(zombies uint64) string {
	if zombies > 0 {
		return ColorRed
	}
	return ColorGreen
}

func createPercentageBar(percent float64, width int) string {
	filled := int((percent / 100.0) * float64(width))
	var bar strings.Builder
	
	// Color the bar based on percentage
	color := getPercentageColor(percent)
	bar.WriteString(color)
	
	for i := 0; i < width; i++ {
		if i < filled {
			if i < filled-1 {
				bar.WriteString(BarFull)
			} else {
				// Partial fill for the last character
				remainder := (percent/100.0)*float64(width) - float64(filled)
				if remainder > 0.75 {
					bar.WriteString(BarHigh)
				} else if remainder > 0.5 {
					bar.WriteString(BarMedium)
				} else if remainder > 0.25 {
					bar.WriteString(BarLow)
				} else {
					bar.WriteString(BarEmpty)
				}
			}
		} else {
			bar.WriteString(BarEmpty)
		}
	}
	
	bar.WriteString(ColorReset)
	return bar.String()
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(num uint64) string {
	if num < 1000 {
		return fmt.Sprintf("%d", num)
	} else if num < 1000000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	} else if num < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	} else {
		return fmt.Sprintf("%.1fG", float64(num)/1000000000)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}