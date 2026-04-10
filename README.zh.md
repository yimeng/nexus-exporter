# Nexus Exporter

[![Release](https://img.shields.io/github/v/release/yimeng/nexus-exporter)](https://github.com/yimeng/nexus-exporter/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://golang.org/)
[![License](https://img.shields.io/github/license/yimeng/nexus-exporter)](LICENSE)

中文 | [English](README.md)

一个用 Go 编写的 Prometheus Exporter，用于监控 Sonatype Nexus Repository Manager 3.x。

[![Release](https://img.shields.io/github/v/release/yimeng/nexus-exporter)](https://github.com/yimeng/nexus-exporter/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://golang.org/)
[![License](https://img.shields.io/github/license/yimeng/nexus-exporter)](LICENSE)

## 功能特性

- **系统状态**: 监控 Nexus 服务健康状态
- **Blob 存储**: 监控存储使用情况、Blob 数量
- **仓库**: 监控仓库信息和组件数量
- **JVM 指标**: 监控内存使用、线程数
- **任务**: 监控计划任务执行状态

## 快速开始

### 下载二进制文件

从 [GitHub Releases](https://github.com/yimeng/nexus-exporter/releases/latest) 下载对应平台的二进制文件。

```bash
# Linux AMD64
curl -LO https://github.com/yimeng/nexus-exporter/releases/latest/download/nexus-exporter-linux-amd64
chmod +x nexus-exporter-linux-amd64
mv nexus-exporter-linux-amd64 nexus-exporter
```

### 使用 Docker

```bash
docker run -d \
  --name nexus-exporter \
  -p 8082:8082 \
  -e NEXUS_URL="http://nexus:8081" \
  -e NEXUS_USERNAME="admin" \
  -e NEXUS_PASSWORD="<your-password>" \
  ghcr.io/yimeng/nexus-exporter:latest
```

## 使用方法

### 命令行参数

```bash
nexus-exporter [flags]
```

#### 可用参数

| 参数 | 短格式 | 环境变量 | 默认值 | 说明 |
|------|--------|----------|--------|------|
| `--help` | `-h` | - | - | 显示帮助信息 |
| `--version` | `-v` | - | - | 显示版本信息 |
| `--config` | - | - | - | 指定 .env 配置文件路径 |
| `--nexus.url` | - | `NEXUS_URL` | `http://localhost:8081` | Nexus URL |
| `--nexus.username` | - | `NEXUS_USERNAME` | `admin` | Nexus 用户名 |
| `--nexus.password` | - | `NEXUS_PASSWORD` | - | Nexus 密码 (必需) |
| `--port` | - | `EXPORTER_PORT` | `8082` | Exporter 监听端口 |
| `--insecure` | - | `NEXUS_INSECURE` | `false` | 跳过 TLS 验证 |
| `--log.level` | - | `LOG_LEVEL` | `info` | 日志级别 (debug/info/warn/error) |

**配置优先级**: 命令行参数 > 环境变量 > 配置文件 (.env) > 默认值

### 使用配置文件

创建 `.env` 文件：

```bash
cat > .env << EOF
NEXUS_URL=http://localhost:8081
NEXUS_USERNAME=admin
NEXUS_PASSWORD=<your-password>
EXPORTER_PORT=8082
NEXUS_INSECURE=false
LOG_LEVEL=info
EOF
```

然后直接运行：

```bash
./nexus-exporter
```

或使用指定配置文件：

```bash
./nexus-exporter --config=/path/to/config.env
```

### 使用环境变量

```bash
export NEXUS_URL="http://localhost:8081"
export NEXUS_USERNAME="admin"
export NEXUS_PASSWORD="<your-password>"
export EXPORTER_PORT="8082"

./nexus-exporter
```

### 使用命令行参数

```bash
./nexus-exporter \
  --nexus.url=http://localhost:8081 \
  --nexus.username=admin \
  --nexus.password=<your-password> \
  --port=8082
```

### Docker 使用 .env 文件

```bash
docker run -d \
  -p 8082:8082 \
  --env-file .env \
  ghcr.io/yimeng/nexus-exporter:latest
```

## 指标列表

| 指标名称 | 类型 | 描述 |
|----------|------|------|
| `nexus_up` | Gauge | Nexus 服务是否可用 (1=up, 0=down) |
| `nexus_version_info` | Gauge | Nexus 版本信息 |
| `nexus_blobstore_bytes_total` | Gauge | Blob 存储总字节数 |
| `nexus_blobstore_bytes_free` | Gauge | Blob 存储可用字节数 |
| `nexus_blobstore_blobs_count` | Gauge | Blob 数量 |
| `nexus_repository_info` | Gauge | 仓库信息 |
| `nexus_repository_components_count` | Gauge | 仓库组件数量 |
| `nexus_jvm_memory_used_bytes` | Gauge | JVM 内存使用量 |
| `nexus_jvm_memory_max_bytes` | Gauge | JVM 内存最大值 |
| `nexus_jvm_threads_count` | Gauge | JVM 线程数 |
| `nexus_task_status` | Gauge | 任务状态 |
| `nexus_task_last_run_timestamp` | Gauge | 任务最后执行时间 |

## Prometheus 配置

```yaml
scrape_configs:
  - job_name: 'nexus'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: /metrics
```

## 告警规则示例

```yaml
groups:
  - name: nexus
    rules:
      - alert: NexusDown
        expr: nexus_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Nexus 服务不可用"
          
      - alert: NexusBlobStoreLowSpace
        expr: nexus_blobstore_bytes_free / nexus_blobstore_bytes_total < 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Nexus Blob 存储空间不足"
          
      - alert: NexusTaskFailed
        expr: nexus_task_status == 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Nexus 任务执行失败"
```

## 构建

```bash
# 构建
go build -o nexus-exporter .

# 或使用 Makefile
make build

# 构建 Docker 镜像
make docker
```

## API 端点

| 端点 | 描述 |
|------|------|
| `/metrics` | Prometheus 指标 |
| `/healthz` | 健康检查 |
| `/` | 状态页面 |

## 开发

```bash
# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 格式化代码
go fmt ./...
```

## 许可证

MIT
