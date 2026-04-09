package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// collectRepositories 收集仓库指标
func (c *NexusCollector) collectRepositories(ch chan<- prometheus.Metric) {
	repos, err := c.client.GetRepositories()
	if err != nil {
		slog.Error("Failed to get repositories", "error", err)
		return
	}

	for _, repo := range repos {
		blobStore := ""
		if repo.Storage != nil {
			blobStore = repo.Storage.BlobStoreName
		}

		ch <- prometheus.MustNewConstMetric(
			c.RepositoryInfo,
			prometheus.GaugeValue,
			1.0,
			repo.Name,
			repo.Format,
			repo.Type,
			blobStore,
		)

		// 获取组件数量（可选，可能较慢）
		// 这里我们只获取第一页来估算数量
		components, err := c.client.GetComponents(repo.Name)
		if err != nil {
			slog.Debug("Failed to get components for repository",
				"repository", repo.Name, "error", err)
			continue
		}

		count := len(components.Items)
		// 如果有更多分页，我们至少知道有 "count+" 个
		if components.ContinuationToken != "" {
			slog.Debug("Repository has more components",
				"repository", repo.Name, "visible", count)
		}

		ch <- prometheus.MustNewConstMetric(
			c.RepositoryComponentCount,
			prometheus.GaugeValue,
			float64(count),
			repo.Name,
		)
	}

	slog.Debug("Collected repository metrics", "count", len(repos))
}
