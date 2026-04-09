package collector

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// collectSystem 收集系统状态指标
func (c *NexusCollector) collectSystem(ch chan<- prometheus.Metric) {
	// 默认 Nexus 是 up 的，只要能成功调用其他 API
	up := 1.0

	// 检查状态检查端点（需要认证）
	_, err := c.client.GetStatusCheck()
	if err != nil {
		slog.Error("Failed to check nexus status", "error", err)
		up = 0.0
	}

	ch <- prometheus.MustNewConstMetric(c.Up, prometheus.GaugeValue, up)

	// 获取系统信息 (OSS 版本此端点可能不可用)
	sysInfo, err := c.client.GetSystemInfo()
	if err != nil {
		// OSS 版本 system/info 不可用，使用固定值
		slog.Debug("System info endpoint not available, using default values", "error", err)
		ch <- prometheus.MustNewConstMetric(
			c.VersionInfo,
			prometheus.GaugeValue,
			1.0,
			"3.76.1",
			"OSS",
			"unknown",
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			c.VersionInfo,
			prometheus.GaugeValue,
			1.0,
			sysInfo.Version,
			sysInfo.Edition,
			sysInfo.NodeID,
		)
	}
}

// parseTime 解析时间字符串为 Unix 时间戳
func parseTime(timeStr string) float64 {
	if timeStr == "" {
		return 0
	}

	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return float64(t.Unix())
		}
	}

	// 尝试直接解析时间戳
	if ts, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return float64(ts)
	}

	return 0
}
