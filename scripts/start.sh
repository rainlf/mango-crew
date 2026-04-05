#!/bin/bash

# MGTT-Go 启动脚本

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

# 检查可执行文件
if [ ! -f "mgtt-go" ]; then
    echo "Building mgtt-go..."
    go build -o mgtt-go cmd/main.go
    if [ $? -ne 0 ]; then
        echo "Build failed"
        exit 1
    fi
fi

# 启动服务
echo "Starting MGTT-Go server..."
./mgtt-go
