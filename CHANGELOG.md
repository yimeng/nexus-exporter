# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-04-09

### Added
- Initial release of nexus-exporter
- Prometheus metrics collection for Nexus Repository Manager 3.x
- Support for system status monitoring (`nexus_up`, `nexus_version_info`)
- Blob store metrics (`nexus_blobstore_bytes_total`, `nexus_blobstore_bytes_free`, `nexus_blobstore_blobs_count`)
- Repository metrics (`nexus_repository_info`, `nexus_repository_components_count`)
- JVM metrics (`nexus_jvm_memory_used_bytes`, `nexus_jvm_memory_max_bytes`, `nexus_jvm_threads_count`)
- Task metrics (`nexus_task_status`, `nexus_task_last_run_timestamp`)
- Docker support with multi-arch build (amd64, arm64)
- GitHub Actions CI/CD workflows
- Security scanning with GitLeaks and TruffleHog
- Multi-platform binary releases (Linux, macOS, Windows)
