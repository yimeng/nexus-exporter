package main

import (
	"flag"
	"fmt"
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

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	goVersion = "unknown"
)

type Config struct {
	NexusURL     string
	NexusUser    string
	NexusPass    string
	ExporterPort string
	Insecure     bool
	LogLevel     string
}

func main() {
	cfg := parseFlags()

	// 初始化日志
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	} else if cfg.LogLevel == "warn" {
		level = slog.LevelWarn
	} else if cfg.LogLevel == "error" {
		level = slog.LevelError
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	// 验证必需参数
	if cfg.NexusPass == "" {
		slog.Error("Nexus password is required. Use --nexus.password or NEXUS_PASSWORD environment variable")
		fmt.Fprintln(os.Stderr, "\nUse --help for usage information")
		os.Exit(1)
	}

	slog.Info("Starting Nexus Exporter",
		"version", version,
		"commit", commit,
		"nexus_url", cfg.NexusURL,
		"exporter_port", cfg.ExporterPort,
	)

	// 创建 Nexus 客户端
	client := nexus.NewClient(cfg.NexusURL, cfg.NexusUser, cfg.NexusPass, cfg.Insecure)

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
	addr := ":" + cfg.ExporterPort
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

func parseFlags() Config {
	var (
		showVersion  bool
		showHelp     bool
		nexusURL     string
		nexusUser    string
		nexusPass    string
		exporterPort string
		insecure     bool
		logLevel     string
	)

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showHelp, "h", false, "Show help (shorthand)")

	flag.StringVar(&nexusURL, "nexus.url", getEnv("NEXUS_URL", "http://localhost:8081"), "Nexus URL")
	flag.StringVar(&nexusUser, "nexus.username", getEnv("NEXUS_USERNAME", "admin"), "Nexus username")
	flag.StringVar(&nexusPass, "nexus.password", getEnv("NEXUS_PASSWORD", ""), "Nexus password")
	flag.StringVar(&exporterPort, "port", getEnv("EXPORTER_PORT", "8082"), "Exporter port")
	flag.BoolVar(&insecure, "insecure", getEnv("NEXUS_INSECURE", "false") == "true", "Skip TLS verification")
	flag.StringVar(&logLevel, "log.level", getEnv("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Nexus Exporter - Prometheus exporter for Sonatype Nexus Repository\n\n")
		fmt.Fprintf(os.Stdout, "Version: %s (commit: %s, built: %s)\n\n", version, commit, date)
		fmt.Fprintf(os.Stdout, "Usage:\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter [flags]\n\n")
		fmt.Fprintf(os.Stdout, "Flags:\n")
		fmt.Fprintf(os.Stdout, "  -h, --help          Show help\n")
		fmt.Fprintf(os.Stdout, "  -v, --version       Show version information\n")
		fmt.Fprintf(os.Stdout, "      --nexus.url     Nexus URL (default: http://localhost:8081)\n")
		fmt.Fprintf(os.Stdout, "      --nexus.username Nexus username (default: admin)\n")
		fmt.Fprintf(os.Stdout, "      --nexus.password Nexus password (required)\n")
		fmt.Fprintf(os.Stdout, "      --port          Exporter port (default: 8082)\n")
		fmt.Fprintf(os.Stdout, "      --insecure      Skip TLS verification\n")
		fmt.Fprintf(os.Stdout, "      --log.level     Log level: debug, info, warn, error (default: info)\n")
		fmt.Fprintf(os.Stdout, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_URL         Nexus URL (default: http://localhost:8081)\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_USERNAME    Nexus username (default: admin)\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_PASSWORD    Nexus password (required)\n")
		fmt.Fprintf(os.Stdout, "  EXPORTER_PORT     Exporter port (default: 8082)\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_INSECURE    Skip TLS verification (default: false)\n")
		fmt.Fprintf(os.Stdout, "  LOG_LEVEL         Log level: debug, info, warn, error (default: info)\n")
		fmt.Fprintf(os.Stdout, "\nExamples:\n")
		fmt.Fprintf(os.Stdout, "  # Basic usage with environment variables\n")
		fmt.Fprintf(os.Stdout, "  export NEXUS_PASSWORD=secret123\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter\n\n")
		fmt.Fprintf(os.Stdout, "  # Using command line flags\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter --nexus.url=http://nexus:8081 --nexus.username=admin --nexus.password=secret123\n\n")
		fmt.Fprintf(os.Stdout, "  # Show version\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter --version\n")
	}

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nexus-exporter version %s (commit: %s, built: %s, go: %s)\n", version, commit, date, goVersion)
		os.Exit(0)
	}

	return Config{
		NexusURL:     nexusURL,
		NexusUser:    nexusUser,
		NexusPass:    nexusPass,
		ExporterPort: exporterPort,
		Insecure:     insecure,
		LogLevel:     logLevel,
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
