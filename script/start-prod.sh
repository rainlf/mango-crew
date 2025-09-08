#!/bin/bash
# 线上环境启动脚本
# 使用生产环境配置（默认application.properties）

# 设置脚本执行失败时立即退出
set -e

echo "正在启动线上环境..."
echo "使用配置：application.properties"
echo "数据初始化：已禁用"

# 构建并启动应用
echo "开始构建项目..."
mvn clean package -DskipTests

echo "构建完成，开始启动应用..."
java -jar target/*.jar