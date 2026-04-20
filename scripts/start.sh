#!/bin/bash

# Mango Crew 启动脚本

# 设置环境变量
export GO_ENV=production

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 进入项目目录
cd "$PROJECT_DIR"

# 检查配置文件
if [ ! -f "configs/config.yaml" ]; then
    echo "Error: configs/config.yaml not found"
    exit 1
fi

# 构建 Linux amd64 二进制文件（适用于 Ubuntu 服务器）
echo "Building mango-crew for Linux amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mango-crew cmd/main.go
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi
echo "Build success: mango-crew"
