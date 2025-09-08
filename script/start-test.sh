#!/bin/bash
# 测试环境启动脚本
# 激活测试环境配置

# 设置脚本执行失败时立即退出
set -e

echo "正在启动测试环境..."
echo "使用配置：application.properties + application-test.properties（测试环境，密码：test）"
echo "数据初始化：已启用"

# 启动Docker中的MySQL服务
echo "检查并启动MySQL服务..."
docker-compose up -d mysql
if [ $? -ne 0 ]; then
  echo "MySQL服务启动失败，程序退出"
  exit 1
fi

echo "MySQL服务启动成功，等待3秒..."
sleep 3

# 构建并启动应用
echo "开始构建项目..."
mvn clean package -DskipTests

echo "构建完成，开始启动应用（测试环境）..."
java -jar -Dspring.profiles.active=test target/*.jar