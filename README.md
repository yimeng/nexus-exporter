# Nexus Exporter

一个用 Go 编写的 Prometheus Exporter，用于监控 Sonatype Nexus Repository Manager 3.x。

## 功能特性

- **系统状态**: 监控 Nexus 服务健康状态
- **Blob 存储**: 监控存储使用情况、Blob 数量
- **仓库**: 监控仓库信息和组件数量
- **JVM 指标**: 监控内存使用、线程数
- **任务**: 监控计划任务执行状态

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

## 快速开始

### 环境变量

| 变量名 | 必填 | 默认值 | 描述 |
|--------|------|--------|------|
| `NEXUS_URL` | 否 | `http://localhost:8081` | Nexus 地址 |
| `NEXUS_USERNAME` | 是 | - | Nexus 用户名 |
| `NEXUS_PASSWORD` | 是 | - | Nexus 密码 |
| `EXPORTER_PORT` | 否 | `8082` | Exporter 监听端口 |
| `NEXUS_INSECURE` | 否 | `false` | 跳过 TLS 验证 |

### 运行

#### 方式一：直接运行

```bash
export NEXUS_URL="http://localhost:8081"
export NEXUS_USERNAME="admin"
export NEXUS_PASSWORD="your-password"
export EXPORTER_PORT="8082"

./nexus-exporter
```

#### 方式二：Docker

```bash
docker run -d \
  --name nexus-exporter \
  -p 8082:8082 \
  -e NEXUS_URL="http://nexus:8081" \
  -e NEXUS_USERNAME="admin" \
  -e NEXUS_PASSWORD="your-password" \
  nexus-exporter:latest
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

## Prometheus 配置

```yaml
scrape_configs:
  - job_name: 'nexus'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: /metrics
```

## API 端点

| 端点 | 描述 |
|------|------|
| `/metrics` | Prometheus 指标 |
| `/healthz` | 健康检查 |
| `/` | 状态页面 |

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
