package collector

import (
	"log/slog"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// collectTasks 收集任务指标
func (c *NexusCollector) collectTasks(ch chan<- prometheus.Metric) {
	tasks, err := c.client.GetTasks()
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		return
	}

	for _, task := range tasks {
		// 任务状态: 1 = 正常/完成, 0 = 失败
		status := 1.0
		if strings.ToUpper(task.CurrentState) == "FAILED" ||
			task.LastRunResult == "FAILED" {
			status = 0.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.TaskStatus,
			prometheus.GaugeValue,
			status,
			task.ID,
			task.Name,
			task.Type,
		)

		// 最后运行时间
		lastRunTime := parseTime(task.LastRun)
		if lastRunTime > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.TaskLastRunTime,
				prometheus.GaugeValue,
				lastRunTime,
				task.ID,
				task.Name,
			)
		}
	}

	slog.Debug("Collected task metrics", "count", len(tasks))
}
