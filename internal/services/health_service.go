package services

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"go-fiber-template/internal/database"
)

// HealthService provides health check functionality
type HealthService struct {
	startTime time.Time
}

// NewHealthService creates a new health service
func NewHealthService() *HealthService {
	return &HealthService{
		startTime: time.Now(),
	}
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]interface{} `json:"checks"`
}

// CheckHealth performs comprehensive health checks
func (h *HealthService) CheckHealth(ctx context.Context) *HealthStatus {
	timestamp := time.Now()
	uptime := timestamp.Sub(h.startTime)

	status := &HealthStatus{
		Status:    "healthy",
		Service:   "go-fiber-template",
		Version:   "1.0.0",
		Timestamp: timestamp,
		Uptime:    formatUptime(uptime),
		Checks:    make(map[string]interface{}),
	}

	// Check database health
	dbHealth := database.CheckHealth(ctx)
	status.Checks["database"] = map[string]interface{}{
		"status":        dbHealth.Status,
		"message":       dbHealth.Message,
		"response_time": dbHealth.ResponseTime.String(),
		"stats":         dbHealth.Stats,
	}

	// Check memory usage
	memStatus := h.checkMemory()
	status.Checks["memory"] = memStatus

	// Check disk space
	diskStatus := h.checkDiskSpace()
	status.Checks["disk_space"] = diskStatus

	// Determine overall status
	if dbHealth.Status != "healthy" {
		status.Status = "unhealthy"
	} else if dbHealth.Status == "degraded" || memStatus["status"] == "warning" || diskStatus["status"] == "warning" {
		status.Status = "degraded"
	}

	return status
}

// checkMemory checks memory usage
func (h *HealthService) checkMemory() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert bytes to MB for readability
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	status := "normal"
	message := "Memory usage is normal"

	// Simple thresholds - in production, these should be configurable
	if allocMB > 100 { // 100MB threshold
		status = "warning"
		message = "Memory usage is high"
	}

	return map[string]interface{}{
		"status":       status,
		"message":      message,
		"alloc_mb":     allocMB,
		"sys_mb":       sysMB,
		"num_gc":       m.NumGC,
		"goroutines":   runtime.NumGoroutine(),
		"heap_objects": m.HeapObjects,
		"next_gc_mb":   float64(m.NextGC) / 1024 / 1024,
	}
}

// checkDiskSpace checks available disk space
func (h *HealthService) checkDiskSpace() map[string]interface{} {
	// For Windows, we'll use a simple approach
	// In production, you might want to use a more sophisticated disk space check
	status := "sufficient"
	message := "Disk space is sufficient"

	// Get disk usage for current directory (simplified for Windows)
	free, total := h.getDiskUsage()

	usedPercent := float64(total-free) / float64(total) * 100

	if usedPercent > 90 {
		status = "critical"
		message = "Disk space is critically low"
	} else if usedPercent > 80 {
		status = "warning"
		message = "Disk space is running low"
	}

	return map[string]interface{}{
		"status":       status,
		"message":      message,
		"free_gb":      float64(free) / 1024 / 1024 / 1024,
		"total_gb":     float64(total) / 1024 / 1024 / 1024,
		"used_percent": usedPercent,
	}
}

// getDiskUsage returns free and total disk space in bytes (Windows-specific)
func (h *HealthService) getDiskUsage() (free, total uint64) {
	// Windows-specific disk space check
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return 0, 0
	}
	defer syscall.FreeLibrary(kernel32)

	getDiskFreeSpaceEx, err := syscall.GetProcAddress(kernel32, "GetDiskFreeSpaceExW")
	if err != nil {
		return 0, 0
	}

	// Get current directory
	path, _ := syscall.UTF16PtrFromString(".")

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	r1, _, _ := syscall.Syscall6(getDiskFreeSpaceEx, 4,
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
		0, 0)

	if r1 == 0 {
		return 0, 0
	}

	return freeBytesAvailable, totalNumberOfBytes
}

// formatUptime formats uptime duration into a human-readable string
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
