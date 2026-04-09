package nexus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 是 Nexus API 的 HTTP 客户端
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient 创建一个新的 Nexus API 客户端
func NewClient(baseURL, username, password string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(method, path string) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// SystemStatus 表示 Nexus 系统状态响应
type SystemStatus struct {
	Status string `json:"status"`
}

// WritableStatus 表示可写状态响应
type WritableStatus struct {
	Writable bool `json:"writable"`
}

// StatusCheck 表示详细状态检查结果
type StatusCheck struct {
	Healthy   bool                     `json:"healthy"`
	Message   string                   `json:"message"`
	Error     string                   `json:"error,omitempty"`
	Details   map[string]interface{}   `json:"details,omitempty"`
}

// SystemInfo 表示系统信息响应 (部分端点在 OSS 版本中不可用)
type SystemInfo struct {
	Version         string `json:"version"`
	Edition         string `json:"edition"`
	BuildRevision   string `json:"buildRevision"`
	BuildTimestamp  string `json:"buildTimestamp"`
	NodeID          string `json:"nodeId"`
	BasePath        string `json:"basePath"`
}

// BlobStore 表示 Blob 存储信息
type BlobStore struct {
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	Path                  string `json:"path,omitempty"`
	BlobCount             int64  `json:"blobCount"`
	TotalSizeInBytes      int64  `json:"totalSizeInBytes"`
	AvailableSpaceInBytes int64  `json:"availableSpaceInBytes,omitempty"`
	SoftQuota             *SoftQuota `json:"softQuota,omitempty"`
}

// SoftQuota 表示软配额设置
type SoftQuota struct {
	Type  string `json:"type"`
	Limit int64  `json:"limit"`
}

// Repository 表示仓库信息
type Repository struct {
	Name        string            `json:"name"`
	Format      string            `json:"format"`
	Type        string            `json:"type"` // hosted, proxy, group
	URL         string            `json:"url"`
	Online      bool              `json:"online"`
	Storage     *RepositoryStorage `json:"storage,omitempty"`
	Cleanup     *RepositoryCleanup `json:"cleanup,omitempty"`
}

// RepositoryStorage 表示仓库存储配置
type RepositoryStorage struct {
	BlobStoreName               string `json:"blobStoreName"`
	StrictContentTypeValidation bool   `json:"strictContentTypeValidation"`
	WritePolicy                 string `json:"writePolicy,omitempty"`
}

// RepositoryCleanup 表示仓库清理配置
type RepositoryCleanup struct {
	PolicyNames []string `json:"policyNames"`
}

// Task 表示任务信息
type Task struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Message     string `json:"message"`
	CurrentState string `json:"currentState"` // WAITING, RUNNING, COMPLETED, FAILED
	LastRunResult string `json:"lastRunResult"`
	LastRun     string `json:"lastRun"`
	NextRun     string `json:"nextRun"`
}

// MetricsData 表示 Nexus 指标数据 (JSON 格式)
type MetricsData struct {
	Version string                 `json:"version"`
	Gauges  map[string]MetricGauge `json:"gauges"`
	Counters map[string]MetricCounter `json:"counters"`
	Meters  map[string]interface{} `json:"meters"`
	Timers  map[string]interface{} `json:"timers"`
	Histograms map[string]interface{} `json:"histograms"`
}

// MetricGauge 表示 Gauge 类型的指标
type MetricGauge struct {
	Value interface{} `json:"value"` // 可能是 int 或 float
}

// GetFloatValue 获取 float64 值
func (m MetricGauge) GetFloatValue() float64 {
	switch v := m.Value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// MetricCounter 表示 Counter 类型的指标
type MetricCounter struct {
	Count int64 `json:"count"`
}

// ComponentList 表示组件列表响应
type ComponentList struct {
	Items             []Component `json:"items"`
	ContinuationToken string      `json:"continuationToken,omitempty"`
}

// Component 表示组件信息
type Component struct {
	ID         string            `json:"id"`
	Repository string            `json:"repository"`
	Format     string            `json:"format"`
	Group      string            `json:"group,omitempty"`
	Name       string            `json:"name"`
	Version    string            `json:"version,omitempty"`
	Assets     []Asset           `json:"assets"`
}

// Asset 表示资产信息
type Asset struct {
	ID            string            `json:"id"`
	Path          string            `json:"path"`
	FileSize      int64             `json:"fileSize"`
	LastUpdated   string            `json:"lastUpdated"`
	LastDownloaded string           `json:"lastDownloaded,omitempty"`
	Attributes    map[string]interface{} `json:"attributes"`
}

// CheckStatus 检查 Nexus 基础状态
func (c *Client) CheckStatus() (*SystemStatus, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/status")
	if err != nil {
		return nil, err
	}

	var status SystemStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// CheckWritable 检查 Nexus 是否可写
func (c *Client) CheckWritable() (*WritableStatus, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/status/writable")
	if err != nil {
		return nil, err
	}

	var status WritableStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// GetStatusCheck 获取详细状态检查
func (c *Client) GetStatusCheck() (*StatusCheck, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/status/check")
	if err != nil {
		return nil, err
	}

	var check StatusCheck
	if err := json.Unmarshal(data, &check); err != nil {
		return nil, err
	}

	return &check, nil
}

// GetSystemInfo 获取系统信息
// 注意: 此端点在 Nexus OSS 3.76.1 中返回 404，保留用于兼容性
func (c *Client) GetSystemInfo() (*SystemInfo, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/system/info")
	if err != nil {
		return nil, err
	}

	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetBlobStores 获取 Blob 存储列表
func (c *Client) GetBlobStores() ([]BlobStore, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/blobstores")
	if err != nil {
		return nil, err
	}

	var stores []BlobStore
	if err := json.Unmarshal(data, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// GetRepositories 获取仓库列表
func (c *Client) GetRepositories() ([]Repository, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/repositories")
	if err != nil {
		return nil, err
	}

	var repos []Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// TaskList 表示任务列表响应
type TaskList struct {
	Items []Task `json:"items"`
}

// GetTasks 获取任务列表
func (c *Client) GetTasks() ([]Task, error) {
	data, err := c.doRequest("GET", "/service/rest/v1/tasks")
	if err != nil {
		return nil, err
	}

	var list TaskList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return list.Items, nil
}

// GetMetrics 获取 JSON 格式指标
func (c *Client) GetMetrics() (*MetricsData, error) {
	data, err := c.doRequest("GET", "/service/metrics/data")
	if err != nil {
		return nil, err
	}

	var metrics MetricsData
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// GetComponents 获取组件列表
func (c *Client) GetComponents(repository string) (*ComponentList, error) {
	path := "/service/rest/v1/components?repository=" + repository
	data, err := c.doRequest("GET", path)
	if err != nil {
		return nil, err
	}

	var list ComponentList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return &list, nil
}
