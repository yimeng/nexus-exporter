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

		// 仓库在线状态
		onlineValue := 0.0
		if repo.Online {
			onlineValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.RepositoryOnline,
			prometheus.GaugeValue,
			onlineValue,
			repo.Name,
		)

		// 获取组件数量
		components, err := c.client.GetComponents(repo.Name)
		if err != nil {
			slog.Debug("Failed to get components for repository",
				"repository", repo.Name, "error", err)
			continue
		}

		count := len(components.Items)
		ch <- prometheus.MustNewConstMetric(
			c.RepositoryComponentCount,
			prometheus.GaugeValue,
			float64(count),
			repo.Name,
		)

		// 获取资产信息（用于估算仓库大小）
		assets, err := c.client.GetAssets(repo.Name)
		if err != nil {
			slog.Debug("Failed to get assets for repository",
				"repository", repo.Name, "error", err)
			continue
		}

		// 计算资产总大小
		var totalSize int64
		for _, asset := range assets.Items {
			totalSize += asset.FileSize
		}

		ch <- prometheus.MustNewConstMetric(
			c.RepositorySize,
			prometheus.GaugeValue,
			float64(totalSize),
			repo.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RepositoryAssetCount,
			prometheus.GaugeValue,
			float64(len(assets.Items)),
			repo.Name,
		)

		if assets.ContinuationToken != "" {
			slog.Debug("Repository has more assets not counted",
				"repository", repo.Name, "visible", len(assets.Items))
		}
	}

	slog.Debug("Collected repository metrics", "count", len(repos))
}
