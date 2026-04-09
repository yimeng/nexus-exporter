package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// collectBlobStores 收集 Blob 存储指标
func (c *NexusCollector) collectBlobStores(ch chan<- prometheus.Metric) {
	stores, err := c.client.GetBlobStores()
	if err != nil {
		slog.Error("Failed to get blob stores", "error", err)
		return
	}

	for _, store := range stores {
		ch <- prometheus.MustNewConstMetric(
			c.BlobStoreTotalBytes,
			prometheus.GaugeValue,
			float64(store.TotalSizeInBytes),
			store.Name,
			store.Type,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlobStoreFreeBytes,
			prometheus.GaugeValue,
			float64(store.AvailableSpaceInBytes),
			store.Name,
			store.Type,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlobStoreBlobCount,
			prometheus.GaugeValue,
			float64(store.BlobCount),
			store.Name,
			store.Type,
		)
	}

	slog.Debug("Collected blob stores metrics", "count", len(stores))
}
