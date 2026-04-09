.PHONY: all build run clean deps test docker

# 变量
BINARY_NAME=nexus-exporter
DOCKER_IMAGE=nexus-exporter:latest
GO=go
GOFLAGS=-v

# 默认目标
all: deps build

# 下载依赖
deps:
	$(GO) mod tidy
	$(GO) mod download

# 构建二进制文件
build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# 运行（需要设置环境变量）
run: build
	./$(BINARY_NAME)

# 清理
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

# 测试
test:
	$(GO) test -v ./...

# 交叉编译
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-linux-amd64 .

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-darwin-amd64 .

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-windows-amd64.exe .

# Docker 构建
docker:
	docker build -t $(DOCKER_IMAGE) .

# 安装到系统
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# 格式化代码
fmt:
	$(GO) fmt ./...

# 代码检查
lint:
	golangci-lint run

# 开发模式（带热重载，需要 air）
dev:
	air
