# Nexus Exporter

[![Release](https://img.shields.io/github/v/release/yimeng/nexus-exporter)](https://github.com/yimeng/nexus-exporter/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://golang.org/)
[![License](https://img.shields.io/github/license/yimeng/nexus-exporter)](LICENSE)

[中文文档](README.zh.md) | English

A Prometheus Exporter written in Go for monitoring Sonatype Nexus Repository Manager 3.x.

## Features

- **System Status**: Monitor Nexus service health status
- **Blob Storage**: Monitor storage usage and blob count
- **Repositories**: Monitor repository information and component count
- **JVM Metrics**: Monitor memory usage and thread count
- **Tasks**: Monitor scheduled task execution status

## Quick Start

### Download Binary

Download the binary for your platform from [GitHub Releases](https://github.com/yimeng/nexus-exporter/releases/latest).

```bash
# Linux AMD64
curl -LO https://github.com/yimeng/nexus-exporter/releases/latest/download/nexus-exporter-linux-amd64
chmod +x nexus-exporter-linux-amd64
mv nexus-exporter-linux-amd64 nexus-exporter
```

### Using Docker

```bash
docker run -d \
  --name nexus-exporter \
  -p 8082:8082 \
  -e NEXUS_URL="http://nexus:8081" \
  -e NEXUS_USERNAME="admin" \
  -e NEXUS_PASSWORD="<your-password>" \
  ghcr.io/yimeng/nexus-exporter:latest
```

## Usage

### Command Line Flags

```bash
nexus-exporter [flags]
```

#### Available Flags

| Flag | Short | Environment Variable | Default | Description |
|------|-------|---------------------|---------|-------------|
| `--help` | `-h` | - | - | Show help information |
| `--version` | `-v` | - | - | Show version information |
| `--config` | - | - | - | Path to .env config file |
| `--nexus.url` | - | `NEXUS_URL` | `http://localhost:8081` | Nexus URL |
| `--nexus.username` | - | `NEXUS_USERNAME` | `admin` | Nexus username |
| `--nexus.password` | - | `NEXUS_PASSWORD` | - | Nexus password (required) |
| `--port` | - | `EXPORTER_PORT` | `8082` | Exporter listen port |
| `--insecure` | - | `NEXUS_INSECURE` | `false` | Skip TLS verification (for self-signed certificates) |
| `--log.level` | - | `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

**Configuration Priority**: Command line flags > Environment variables > Config file (.env) > Default values

### Using Config File

Create a `.env` file:

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

Then run directly:

```bash
./nexus-exporter
```

Or specify config file path:

```bash
./nexus-exporter --config=/path/to/config.env
```

### Using Environment Variables

```bash
export NEXUS_URL="http://localhost:8081"
export NEXUS_USERNAME="admin"
export NEXUS_PASSWORD="<your-password>"
export EXPORTER_PORT="8082"

./nexus-exporter
```

### Using Command Line Flags

```bash
./nexus-exporter \
  --nexus.url=http://localhost:8081 \
  --nexus.username=admin \
  --nexus.password=<your-password> \
  --port=8082
```

### Docker with .env File

```bash
docker run -d \
  -p 8082:8082 \
  --env-file .env \
  ghcr.io/yimeng/nexus-exporter:latest
```

## Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| `nexus_up` | Gauge | Nexus service availability (1=up, 0=down) |
| `nexus_version_info` | Gauge | Nexus version information |
| `nexus_blobstore_bytes_total` | Gauge | Total bytes in blob store |
| `nexus_blobstore_bytes_free` | Gauge | Available bytes in blob store |
| `nexus_blobstore_blobs_count` | Gauge | Number of blobs |
| `nexus_repository_info` | Gauge | Repository information |
| `nexus_repository_components_count` | Gauge | Number of components in repository |
| `nexus_jvm_memory_used_bytes` | Gauge | JVM memory usage |
| `nexus_jvm_memory_max_bytes` | Gauge | JVM memory maximum |
| `nexus_jvm_threads_count` | Gauge | JVM thread count |
| `nexus_task_status` | Gauge | Task status |
| `nexus_task_last_run_timestamp` | Gauge | Task last run timestamp |

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'nexus'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: /metrics
```

## Alerting Rules Example

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
          summary: "Nexus service is down"
          
      - alert: NexusBlobStoreLowSpace
        expr: nexus_blobstore_bytes_free / nexus_blobstore_bytes_total < 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Nexus Blob Store low disk space"
          
      - alert: NexusTaskFailed
        expr: nexus_task_status == 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Nexus task execution failed"
```

## Building

```bash
# Build
go build -o nexus-exporter .

# Or use Makefile
make build

# Build Docker image
make docker
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/metrics` | Prometheus metrics |
| `/healthz` | Health check |
| `/` | Status page |

## Troubleshooting

### HTTPS/HTTP Mismatch Error

**Error**: `server gave HTTP response to HTTPS client`

**Solution**: Your Nexus server is using HTTP, but you specified HTTPS. Change the URL:
```bash
# Wrong
--nexus.url=https://192.168.0.110:8081

# Correct
--nexus.url=http://192.168.0.110:8081
```

### TLS Certificate Error

**Error**: `certificate signed by unknown authority`

**Solution**: If using a self-signed certificate, add the `--insecure` flag:
```bash
./nexus-exporter --nexus.url=https://192.168.0.110:8081 --nexus.password=<your-password> --insecure
```

Or add to `.env` config file:
```bash
NEXUS_URL=https://192.168.0.110:8081
NEXUS_INSECURE=true
```

### Normal HTTPS Certificate (Non Self-Signed)

If Nexus uses a valid HTTPS certificate (e.g., Let's Encrypt or enterprise CA), **no special flags are needed**:
```bash
./nexus-exporter --nexus.url=https://nexus.example.com --nexus.password=<your-password>
```

## Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Format code
go fmt ./...
```

## License

MIT
