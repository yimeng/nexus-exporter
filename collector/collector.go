package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"nexus-exporter/nexus"
)

// NexusCollector 收集所有 Nexus 指标
type NexusCollector struct {
	client *nexus.Client

	// 系统指标
	Up          *prometheus.Desc
	VersionInfo *prometheus.Desc

	// Blob 存储指标
	BlobStoreTotalBytes *prometheus.Desc
	BlobStoreFreeBytes  *prometheus.Desc
	BlobStoreBlobCount  *prometheus.Desc

	// 仓库指标
	RepositoryInfo          *prometheus.Desc
	RepositoryComponentCount *prometheus.Desc
	RepositoryOnline        *prometheus.Desc
	RepositorySize          *prometheus.Desc
	RepositoryAssetCount    *prometheus.Desc

	// JVM 指标
	JVMMemoryUsed *prometheus.Desc
	JVMMemoryMax  *prometheus.Desc
	JVMThreads    *prometheus.Desc

	// 任务指标
	TaskStatus      *prometheus.Desc
	TaskLastRunTime *prometheus.Desc
}

// NewNexusCollector 创建新的收集器
func NewNexusCollector(client *nexus.Client) *NexusCollector {
	return &NexusCollector{
		client: client,

		Up: prometheus.NewDesc(
			"nexus_up",
			"Nexus service is up (1) or down (0)",
			nil, nil,
		),
		VersionInfo: prometheus.NewDesc(
			"nexus_version_info",
			"Nexus version information",
			[]string{"version", "edition", "node_id"}, nil,
		),
		BlobStoreTotalBytes: prometheus.NewDesc(
			"nexus_blobstore_bytes_total",
			"Total bytes in blob store",
			[]string{"name", "type"}, nil,
		),
		BlobStoreFreeBytes: prometheus.NewDesc(
			"nexus_blobstore_bytes_free",
			"Available bytes in blob store",
			[]string{"name", "type"}, nil,
		),
		BlobStoreBlobCount: prometheus.NewDesc(
			"nexus_blobstore_blobs_count",
			"Number of blobs in blob store",
			[]string{"name", "type"}, nil,
		),
		RepositoryInfo: prometheus.NewDesc(
			"nexus_repository_info",
			"Repository information",
			[]string{"name", "format", "type", "blob_store"}, nil,
		),
		RepositoryComponentCount: prometheus.NewDesc(
			"nexus_repository_components_count",
			"Number of components in repository",
			[]string{"name"}, nil,
		),
		RepositoryOnline: prometheus.NewDesc(
			"nexus_repository_online",
			"Repository online status (1=online, 0=offline)",
			[]string{"name"}, nil,
		),
		RepositorySize: prometheus.NewDesc(
			"nexus_repository_size_bytes",
			"Total size of repository in bytes (estimated from assets)",
			[]string{"name"}, nil,
		),
		RepositoryAssetCount: prometheus.NewDesc(
			"nexus_repository_assets_count",
			"Number of assets in repository",
			[]string{"name"}, nil,
		),
		JVMMemoryUsed: prometheus.NewDesc(
			"nexus_jvm_memory_used_bytes",
			"JVM memory used in bytes",
			[]string{"area"}, nil,
		),
		JVMMemoryMax: prometheus.NewDesc(
			"nexus_jvm_memory_max_bytes",
			"JVM memory max in bytes",
			[]string{"area"}, nil,
		),
		JVMThreads: prometheus.NewDesc(
			"nexus_jvm_threads_count",
			"Number of JVM threads",
			nil, nil,
		),
		TaskStatus: prometheus.NewDesc(
			"nexus_task_status",
			"Task status (1=healthy, 0=unhealthy)",
			[]string{"id", "name", "type"}, nil,
		),
		TaskLastRunTime: prometheus.NewDesc(
			"nexus_task_last_run_timestamp",
			"Task last run timestamp",
			[]string{"id", "name"}, nil,
		),
	}
}

// Describe 实现 prometheus.Collector 接口
func (c *NexusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.VersionInfo
	ch <- c.BlobStoreTotalBytes
	ch <- c.BlobStoreFreeBytes
	ch <- c.BlobStoreBlobCount
	ch <- c.RepositoryInfo
	ch <- c.RepositoryComponentCount
	ch <- c.RepositoryOnline
	ch <- c.RepositorySize
	ch <- c.RepositoryAssetCount
	ch <- c.JVMMemoryUsed
	ch <- c.JVMMemoryMax
	ch <- c.JVMThreads
	ch <- c.TaskStatus
	ch <- c.TaskLastRunTime
}

// Collect 实现 prometheus.Collector 接口
func (c *NexusCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	// 系统状态
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.collectSystem(ch)
	}()

	// Blob 存储
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.collectBlobStores(ch)
	}()

	// 仓库
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.collectRepositories(ch)
	}()

	// JVM 指标
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.collectJVMMetrics(ch)
	}()

	// 任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.collectTasks(ch)
	}()

	wg.Wait()
}
