# Nexus Exporter

[![Release](https://img.shields.io/github/v/release/yimeng/nexus-exporter)](https://github.com/yimeng/nexus-exporter/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://golang.org/)
[![License](https://img.shields.io/github/license/yimeng/nexus-exporter)](LICENSE)

中文 | [English](README.md)

一个用 Go 编写的 Prometheus Exporter，用于监控 Sonatype Nexus Repository Manager 3.x。

![Nexus SRE Dashboard](docs/images/dashboard-v1.2.0.png)

## 功能特性

- **系统状态**: 监控 Nexus 服务健康状态
- **Blob 存储**: 监控存储使用情况、Blob 数量
- **仓库**: 监控仓库信息和组件数量
- **JVM 指标**: 监控内存使用、线程数
- **任务**: 监控计划任务执行状态

## 快速开始

### 二进制部署 (Systemd)

在 Linux 上部署 nexus-exporter 作为 systemd 服务。

#### 前置条件: Nexus 用户权限

Exporter 使用的 Nexus 用户账号需要以下权限：

| 权限 | API 端点 | 用途 | Grafana 面板 |
|------|----------|------|--------------|
| `nx-healthcheck-read` | `/service/rest/v1/status` | 检查 Nexus 健康状态 | 服务可用性 |
| `nx-blobstores-read` | `/service/rest/v1/blobstores` | 读取 Blob 存储指标 | Blob 存储使用、磁盘使用率 |
| `nx-repository-view-*-*-read` | `/service/rest/v1/repositories` | 列出仓库 | **仓库总数**、仓库容量与状态 |
| `nx-component-read` | `/service/rest/v1/components` | 读取组件数量 | **各仓库组件数** |
| `nx-assets-read` | `/service/rest/v1/assets` | 读取资产信息 | 仓库大小、资产数量 |
| `nx-tasks-read` | `/service/rest/v1/tasks` | 读取任务状态 | **失败任务**、任务健康 |
| `nx-metrics-read` | `/service/metrics/data` | 读取 JVM 指标 | JVM 内存、JVM 线程数 |

**注意**: 如果以下面板显示 "No data"：
- **仓库总数**: 检查 `nx-repository-view-*-*-read` 权限
- **失败任务**: 检查 `nx-tasks-read` 权限
- **各仓库组件数**: 检查 `nx-component-read` 权限

**推荐**: 使用 admin 账户或创建具有上述权限的专用服务账户。

对于 Nexus OSS，默认的 `admin` 账户拥有所有必需的权限。

如需创建自定义角色，进入 **Nexus UI → 安全性 → 角色 → 创建角色** 并分配上述权限。

#### 1. 下载二进制文件

```bash
# 检测架构
ARCH=$(uname -m)
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

# 下载最新版本
curl -LO "https://github.com/yimeng/nexus-exporter/releases/latest/download/nexus-exporter-linux-${ARCH}"
sudo install -m 755 "nexus-exporter-linux-${ARCH}" /usr/local/bin/nexus-exporter
rm "nexus-exporter-linux-${ARCH}"
```

#### 2. 创建用户和目录

```bash
# 创建专用用户
sudo useradd --system --no-create-home --shell /usr/sbin/nologin nexus-exporter

# 创建配置目录
sudo mkdir -p /etc/nexus-exporter
sudo chmod 750 /etc/nexus-exporter
```

#### 3. 配置

创建包含 Nexus 凭证的配置文件：

```bash
sudo tee /etc/nexus-exporter/nexus-exporter.conf << 'EOF'
NEXUS_URL=http://localhost:8081
NEXUS_USERNAME=admin
NEXUS_PASSWORD=your-nexus-password
EXPORTER_PORT=8082
LOG_LEVEL=info
EOF

# 保护配置文件
sudo chmod 600 /etc/nexus-exporter/nexus-exporter.conf
sudo chown root:nexus-exporter /etc/nexus-exporter/nexus-exporter.conf
```

#### 4. 安装 Systemd 服务

```bash
# 下载服务文件
curl -L -o /tmp/nexus-exporter.service \
  https://raw.githubusercontent.com/yimeng/nexus-exporter/master/systemd/nexus-exporter.service

# 安装并重新加载
sudo install -m 644 /tmp/nexus-exporter.service /etc/systemd/system/
sudo systemctl daemon-reload
```

或手动创建：

```bash
sudo tee /etc/systemd/system/nexus-exporter.service << 'EOF'
[Unit]
Description=Nexus Exporter for Prometheus
After=network.target

[Service]
Type=simple
User=nexus-exporter
Group=nexus-exporter
EnvironmentFile=/etc/nexus-exporter/nexus-exporter.conf
ExecStart=/usr/local/bin/nexus-exporter
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
```

#### 5. 启动服务

```bash
# 启用并启动服务
sudo systemctl enable nexus-exporter
sudo systemctl start nexus-exporter

# 检查状态
sudo systemctl status nexus-exporter

# 查看日志
sudo journalctl -u nexus-exporter -f
```

#### 6. 验证

```bash
# 测试指标端点
curl http://localhost:8082/metrics
```

#### 故障排除

**服务启动失败，提示 "Nexus password is required"**

如果在日志中看到此错误：
```
Nexus password is required. Use --nexus.password, NEXUS_PASSWORD environment variable, or .env file
```

请检查：
1. 配置文件存在且权限正确：
   ```bash
   sudo ls -la /etc/nexus-exporter/nexus-exporter.conf
   # 应该是: -rw------- root nexus-exporter
   ```

2. 环境变量已加载（测试命令）：
   ```bash
   sudo systemctl show nexus-exporter --property=Environment
   ```

3. 替代方案：在服务文件中使用 `--config` 参数：
   ```ini
   ExecStart=/usr/local/bin/nexus-exporter --config=/etc/nexus-exporter/nexus-exporter.conf
   ```

**权限被拒绝错误**

如果日志中看到 401/403 错误，请检查：
1. Nexus 用户凭据是否正确
2. Nexus 用户是否具有所需权限（参见前置条件部分）
3. 对于非管理员用户，确保角色包含：`nx-healthcheck-read`、`nx-blobstores-read`、`nx-repository-view-*-*-read`、`nx-tasks-read`、`nx-metrics-read`

---

### Docker 部署

使用 Docker 或 Docker Compose 运行 nexus-exporter。

#### 方式 1: Docker Run

```bash
docker run -d \
  --name nexus-exporter \
  --restart unless-stopped \
  -p 8082:8082 \
  -e NEXUS_URL="http://nexus:8081" \
  -e NEXUS_USERNAME="admin" \
  -e NEXUS_PASSWORD="your-nexus-password" \
  -e EXPORTER_PORT="8082" \
  -e LOG_LEVEL="info" \
  ghcr.io/yimeng/nexus-exporter:latest
```

#### 方式 2: Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  nexus-exporter:
    image: ghcr.io/yimeng/nexus-exporter:latest
    container_name: nexus-exporter
    restart: unless-stopped
    ports:
      - "8082:8082"
    environment:
      - NEXUS_URL=http://nexus:8081
      - NEXUS_USERNAME=admin
      - NEXUS_PASSWORD=${NEXUS_PASSWORD}
      - EXPORTER_PORT=8082
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8082/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

启动：

```bash
# 创建 .env 文件存储敏感数据
echo "NEXUS_PASSWORD=your-nexus-password" > .env

# 启动容器
docker compose up -d

# 查看日志
docker compose logs -f

# 检查状态
docker compose ps
```

#### 方式 3: 从源码构建

```bash
# 克隆仓库
git clone https://github.com/yimeng/nexus-exporter.git
cd nexus-exporter

# 构建镜像
docker build -t nexus-exporter:local .

# 运行
docker run -d \
  --name nexus-exporter \
  -p 8082:8082 \
  -e NEXUS_PASSWORD="your-nexus-password" \
  nexus-exporter:local
```

---

### 监控配置

配置您的监控系统从 nexus-exporter 抓取指标。

#### Prometheus

添加到 `prometheus.yml`：

```yaml
scrape_configs:
  - job_name: 'nexus-exporter'
    static_configs:
      - targets: ['nexus-exporter:8082']
    metrics_path: /metrics
    scrape_interval: 30s
    scrape_timeout: 10s
```

#### VictoriaMetrics

使用 `scrape_config` 文件：

```yaml
# /etc/victoriametrics/scrape.yml
scrape_configs:
  - job_name: 'nexus-exporter'
    static_configs:
      - targets: ['nexus-exporter:8082']
        labels:
          instance: 'nexus-server-01'
          environment: 'production'
    metrics_path: /metrics
    scrape_interval: 30s
    scrape_timeout: 10s
```

启动 VictoriaMetrics：

```bash
victoria-metrics \
  -promscrape.config=/etc/victoriametrics/scrape.yml \
  -retentionPeriod=30d \
  -httpListenAddr=:8428
```

#### Docker Compose (完整监控栈)

包含 Nexus、Exporter、VictoriaMetrics 和 Grafana 的完整监控栈：

```yaml
version: '3.8'

services:
  nexus:
    image: sonatype/nexus3:3.76.1
    container_name: nexus
    ports:
      - "8081:8081"
    volumes:
      - nexus-data:/nexus-data
    environment:
      - INSTALL4J_ADD_VM_PARAMS=-Xms1g -Xmx1g

  nexus-exporter:
    image: ghcr.io/yimeng/nexus-exporter:latest
    container_name: nexus-exporter
    ports:
      - "8082:8082"
    environment:
      - NEXUS_URL=http://nexus:8081
      - NEXUS_USERNAME=admin
      - NEXUS_PASSWORD=${NEXUS_PASSWORD}
    depends_on:
      - nexus

  victoriametrics:
    image: victoriametrics/victoria-metrics:v1.102.0
    container_name: victoriametrics
    ports:
      - "8428:8428"
    volumes:
      - ./victoriametrics.yml:/etc/victoriametrics/scrape.yml:ro
      - vm-data:/victoria-metrics-data
    command:
      - '--promscrape.config=/etc/victoriametrics/scrape.yml'
      - '--storageDataPath=/victoria-metrics-data'
      - '--retentionPeriod=30d'

  grafana:
    image: grafana/grafana:12.4.0
    container_name: grafana
    ports:
      - "3001:3000"
    volumes:
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin123}

volumes:
  nexus-data:
  vm-data:
  grafana-data:
```

`victoriametrics.yml`：

```yaml
scrape_configs:
  - job_name: 'nexus-exporter'
    static_configs:
      - targets: ['nexus-exporter:8082']
    metrics_path: /metrics
    scrape_interval: 30s
```

#### 告警规则 (Prometheus/VictoriaMetrics)

```yaml
groups:
  - name: nexus-alerts
    rules:
      - alert: NexusDown
        expr: nexus_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Nexus 服务宕机"
          description: "Nexus 已宕机超过 1 分钟"

      - alert: NexusHighDiskUsage
        expr: (nexus_blobstore_bytes_total / (nexus_blobstore_bytes_total + nexus_blobstore_bytes_free)) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Nexus 磁盘使用率过高"
          description: "Blob 存储 {{ $labels.name }} 使用率 {{ $value | humanize }}%"

      - alert: NexusTaskFailed
        expr: nexus_task_status == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Nexus 任务执行失败"
          description: "任务 {{ $labels.name }} 执行失败"

      - alert: NexusHighJVMMemoryUsage
        expr: (nexus_jvm_memory_used_bytes / nexus_jvm_memory_max_bytes) * 100 > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Nexus JVM 内存使用率过高"
          description: "JVM {{ $labels.area }} 内存使用率 {{ $value | humanize }}%"
```

#### Grafana 数据源配置

**Prometheus 数据源：**
- URL: `http://prometheus:9090` (或 VictoriaMetrics: `http://victoriametrics:8428`)
- 访问方式: Server

**导入 Dashboard：**
1. 进入 Grafana → 创建 → 导入
2. 输入 dashboard ID 或粘贴 JSON
3. 选择 Prometheus/VictoriaMetrics 数据源

Dashboard JSON 位于 `grafana/dashboards/nexus-sre-dashboard.json`

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
| `--insecure` | - | `NEXUS_INSECURE` | `false` | 跳过 TLS 验证（用于自签名证书） |
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
| `nexus_repository_info` | Gauge | 仓库信息 (名称、格式、类型、blob_store) |
| `nexus_repository_components_count` | Gauge | 仓库组件数量 |
| `nexus_repository_online` | Gauge | 仓库在线状态 (1=在线, 0=离线) |
| `nexus_repository_size_bytes` | Gauge | 仓库总大小（字节） |
| `nexus_repository_assets_count` | Gauge | 仓库资产数量 |
| `nexus_jvm_memory_used_bytes` | Gauge | JVM 内存使用量 |
| `nexus_jvm_memory_max_bytes` | Gauge | JVM 内存最大值 |
| `nexus_jvm_threads_count` | Gauge | JVM 线程数 |
| `nexus_task_status` | Gauge | 任务状态 (1=健康, 0=失败) |
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

## 故障排除

### HTTPS/HTTP 协议不匹配错误

**错误**: `server gave HTTP response to HTTPS client`

**解决方法**: Nexus 服务器使用的是 HTTP 协议，但你配置了 HTTPS。请修改 URL：
```bash
# 错误
--nexus.url=https://192.168.0.110:8081

# 正确
--nexus.url=http://192.168.0.110:8081
```

### TLS 证书错误

**错误**: `certificate signed by unknown authority`

**解决方法**: 如果使用自签名证书，添加 `--insecure` 参数：
```bash
./nexus-exporter --nexus.url=https://192.168.0.110:8081 --nexus.password=<your-password> --insecure
```

或者在 `.env` 配置文件中添加：
```bash
NEXUS_URL=https://192.168.0.110:8081
NEXUS_INSECURE=true
```

### 正常 HTTPS 证书（非自签名）

如果 Nexus 使用正常的 HTTPS 证书（如 Let's Encrypt 或企业证书），**不需要** `--insecure` 参数：
```bash
./nexus-exporter --nexus.url=https://nexus.example.com --nexus.password=<your-password>
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
