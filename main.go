package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
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
		slog.Error("Nexus password is required. Use --nexus.password, NEXUS_PASSWORD environment variable, or .env file")
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
		configFile   string
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
	flag.StringVar(&configFile, "config", "", "Path to .env config file")

	flag.StringVar(&nexusURL, "nexus.url", "", "Nexus URL")
	flag.StringVar(&nexusUser, "nexus.username", "", "Nexus username")
	flag.StringVar(&nexusPass, "nexus.password", "", "Nexus password")
	flag.StringVar(&exporterPort, "port", "", "Exporter port")
	flag.BoolVar(&insecure, "insecure", false, "Skip TLS verification")
	flag.StringVar(&logLevel, "log.level", "", "Log level (debug, info, warn, error)")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Nexus Exporter - Prometheus exporter for Sonatype Nexus Repository\n\n")
		fmt.Fprintf(os.Stdout, "Version: %s (commit: %s, built: %s)\n\n", version, commit, date)
		fmt.Fprintf(os.Stdout, "Usage:\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter [flags]\n\n")
		fmt.Fprintf(os.Stdout, "Flags:\n")
		fmt.Fprintf(os.Stdout, "  -h, --help          Show help\n")
		fmt.Fprintf(os.Stdout, "  -v, --version       Show version information\n")
		fmt.Fprintf(os.Stdout, "      --config        Path to .env config file\n")
		fmt.Fprintf(os.Stdout, "      --nexus.url     Nexus URL\n")
		fmt.Fprintf(os.Stdout, "      --nexus.username Nexus username\n")
		fmt.Fprintf(os.Stdout, "      --nexus.password Nexus password\n")
		fmt.Fprintf(os.Stdout, "      --port          Exporter port\n")
		fmt.Fprintf(os.Stdout, "      --insecure      Skip TLS verification\n")
		fmt.Fprintf(os.Stdout, "      --log.level     Log level: debug, info, warn, error\n")
		fmt.Fprintf(os.Stdout, "\nConfiguration Priority (highest to lowest):\n")
		fmt.Fprintf(os.Stdout, "  1. Command line flags\n")
		fmt.Fprintf(os.Stdout, "  2. Environment variables\n")
		fmt.Fprintf(os.Stdout, "  3. Config file (.env)\n")
		fmt.Fprintf(os.Stdout, "  4. Default values\n")
		fmt.Fprintf(os.Stdout, "\nConfig File (.env):\n")
		fmt.Fprintf(os.Stdout, "  Create a .env file in the working directory or specify with --config:\n\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_URL=http://localhost:8081\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_USERNAME=admin\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_PASSWORD=<your-password>\n")
		fmt.Fprintf(os.Stdout, "  EXPORTER_PORT=8082\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_INSECURE=false\n")
		fmt.Fprintf(os.Stdout, "  LOG_LEVEL=info\n")
		fmt.Fprintf(os.Stdout, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_URL         Nexus URL (default: http://localhost:8081)\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_USERNAME    Nexus username (default: admin)\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_PASSWORD    Nexus password (required)\n")
		fmt.Fprintf(os.Stdout, "  EXPORTER_PORT     Exporter port (default: 8082)\n")
		fmt.Fprintf(os.Stdout, "  NEXUS_INSECURE    Skip TLS verification (default: false)\n")
		fmt.Fprintf(os.Stdout, "  LOG_LEVEL         Log level: debug, info, warn, error (default: info)\n")
		fmt.Fprintf(os.Stdout, "\nExamples:\n")
		fmt.Fprintf(os.Stdout, "  # Use default .env file\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter\n\n")
		fmt.Fprintf(os.Stdout, "  # Use specific config file\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter --config=/path/to/config.env\n\n")
		fmt.Fprintf(os.Stdout, "  # Use command line flags (highest priority)\n")
		fmt.Fprintf(os.Stdout, "  nexus-exporter --nexus.url=http://nexus:8081 --nexus.password=<your-password>\n\n")
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

	// 加载配置文件（如果指定或存在默认的 .env）
	loadConfigFile(configFile)

	// 获取配置值（按优先级：命令行 > 环境变量 > 配置文件 > 默认值）
	return Config{
		NexusURL:     getStringValue(nexusURL, "NEXUS_URL", "http://localhost:8081"),
		NexusUser:    getStringValue(nexusUser, "NEXUS_USERNAME", "admin"),
		NexusPass:    getStringValue(nexusPass, "NEXUS_PASSWORD", ""),
		ExporterPort: getStringValue(exporterPort, "EXPORTER_PORT", "8082"),
		Insecure:     getBoolValue(insecure, "NEXUS_INSECURE", false),
		LogLevel:     getStringValue(logLevel, "LOG_LEVEL", "info"),
	}
}

// loadConfigFile 加载配置文件
func loadConfigFile(configFile string) {
	if configFile != "" {
		// 加载指定的配置文件
		if _, err := os.Stat(configFile); err == nil {
			if err := godotenv.Load(configFile); err != nil {
				// 仅在文件存在但无法读取时警告
				fmt.Fprintf(os.Stderr, "Warning: Failed to load config file %s: %v\n", configFile, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Config file not found: %s\n", configFile)
		}
	} else {
		// 尝试加载默认的 .env 文件
		// 首先在可执行文件所在目录查找
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			envPath := filepath.Join(execDir, ".env")
			if _, err := os.Stat(envPath); err == nil {
				godotenv.Load(envPath)
				return
			}
		}

		// 然后在当前工作目录查找
		if _, err := os.Stat(".env"); err == nil {
			godotenv.Load(".env")
		}
	}
}

// getStringValue 获取字符串配置值（优先级：命令行 > 环境变量 > 默认值）
func getStringValue(flagValue, envKey, defaultValue string) string {
	// 1. 如果命令行参数已设置，使用命令行参数
	if flagValue != "" {
		return flagValue
	}

	// 2. 尝试从环境变量获取
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}

	// 3. 返回默认值
	return defaultValue
}

// getBoolValue 获取布尔配置值（优先级：命令行 > 环境变量 > 默认值）
func getBoolValue(flagValue bool, envKey string, defaultValue bool) bool {
	// 1. 如果命令行参数已设置（且不是默认值 false 或者是显式设置的）
	// 注意：对于 bool 类型，无法区分 "未设置" 和 "设置为 false"
	// 所以这里我们需要检查环境变量

	// 2. 尝试从环境变量获取
	envValue := os.Getenv(envKey)
	if envValue != "" {
		switch envValue {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}

	// 3. 如果命令行参数显式设置，使用命令行参数
	// 由于 bool flag 无法判断是否设置，我们依赖环境变量优先
	// 这里简单处理：如果环境变量未设置，返回 flagValue
	if envValue == "" {
		return flagValue || defaultValue
	}

	return defaultValue
}

// getEnv 获取环境变量，如果不存在则返回默认值（保留用于向后兼容）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
