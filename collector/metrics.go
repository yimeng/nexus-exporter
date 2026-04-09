package collector

import (
	"log/slog"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// collectJVMMetrics 收集 JVM 性能指标
func (c *NexusCollector) collectJVMMetrics(ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetMetrics()
	if err != nil {
		slog.Error("Failed to get metrics", "error", err)
		return
	}

	// 提取 JVM 内存指标
	for name, gauge := range metrics.Gauges {
		value := gauge.GetFloatValue()

		// JVM 内存使用
		if strings.HasPrefix(name, "jvm.memory.heap.used") {
			ch <- prometheus.MustNewConstMetric(
				c.JVMMemoryUsed,
				prometheus.GaugeValue,
				value,
				"heap",
			)
		}
		if strings.HasPrefix(name, "jvm.memory.non-heap.used") {
			ch <- prometheus.MustNewConstMetric(
				c.JVMMemoryUsed,
				prometheus.GaugeValue,
				value,
				"non_heap",
			)
		}

		// JVM 内存最大值
		if strings.HasPrefix(name, "jvm.memory.heap.max") {
			ch <- prometheus.MustNewConstMetric(
				c.JVMMemoryMax,
				prometheus.GaugeValue,
				value,
				"heap",
			)
		}
		if strings.HasPrefix(name, "jvm.memory.non-heap.max") {
			ch <- prometheus.MustNewConstMetric(
				c.JVMMemoryMax,
				prometheus.GaugeValue,
				value,
				"non_heap",
			)
		}

		// 线程数
		if name == "jvm.threads.count" {
			ch <- prometheus.MustNewConstMetric(
				c.JVMThreads,
				prometheus.GaugeValue,
				value,
			)
		}
	}

	slog.Debug("Collected JVM metrics")
}
