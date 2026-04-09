package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"nexus-exporter/collector"
	"nexus-exporter/nexus"
)

func main() {
	// 初始化日志
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// 从环境变量读取配置
	nexusURL := getEnv("NEXUS_URL", "http://localhost:8081")
	nexusUser := getEnv("NEXUS_USERNAME", "admin")
	nexusPass := getEnv("NEXUS_PASSWORD", "")
	exporterPort := getEnv("EXPORTER_PORT", "8082")
	insecure := getEnv("NEXUS_INSECURE", "false") == "true"

	if nexusPass == "" {
		slog.Error("NEXUS_PASSWORD environment variable is required")
		os.Exit(1)
	}

	slog.Info("Starting Nexus Exporter",
		"nexus_url", nexusURL,
		"exporter_port", exporterPort,
	)

	// 创建 Nexus 客户端
	client := nexus.NewClient(nexusURL, nexusUser, nexusPass, insecure)

	// 创建收集器
	nexusCollector := collector.NewNexusCollector(client)

	// 注册收集器
	prometheus.MustRegister(nexusCollector)

	// 注册内置指标 (使用新版 collectors 包)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	// 设置 HTTP 路由
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 检查 Nexus 是否可用 (通过状态检查端点)
		check, err := client.GetStatusCheck()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, writeErr := w.Write([]byte(`{"status":"unhealthy","error":"` + err.Error() + `"}`)); writeErr != nil {
				slog.Error("Failed to write response", "error", writeErr)
			}
			return
		}

		healthy := "true"
		if !check.Healthy {
			healthy = "false"
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy","nexus_healthy":` + healthy + `}`)); err != nil {
			slog.Error("Failed to write response", "error", err)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(`<html>
<head><title>Nexus Exporter</title></head>
<body>
<h1>Nexus Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/healthz">Health Check</a></p>
</body>
</html>`)); err != nil {
			slog.Error("Failed to write response", "error", err)
		}
	})

	// 启动 HTTP 服务
	addr := ":" + exporterPort
	slog.Info("Server starting", "address", addr)

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
