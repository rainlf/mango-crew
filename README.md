# MGTT 
## SQL
数据库表结构的DDL脚本已移至独立文件，位于项目中的以下位置：

`src/main/resources/sql/init_schema.sql`

该脚本包含完整的数据库初始化语句，包括：
- 创建数据库 `mgtt`
- 创建数据表 `mgtt_config`、`mgtt_log`、`mgtt_user`、`mgtt_majiang_game` 、`mgtt_api_monitor` 和 `mgtt_majiang_game_item`
- 定义各表字段、数据类型、约束条件和默认值

实体类的定义已根据此DDL进行调整，确保自动生成的数据库表结构与DDL保持一致。
## Network
![342318525-b610a992-552d-40d0-8db9-56a1edd6c7c5](https://github.com/user-attachments/assets/2922f53f-b507-4552-bc09-d2bd3d452d4e)

## 启动方法

### 测试环境启动

测试环境使用`application-test.properties`配置，包含以下特性：
- 启用数据初始化
- 使用测试环境数据库配置
- 启动Docker中的MySQL服务

**执行步骤：**

在Linux/Mac环境下：
```bash
# 确保Docker服务已启动
./start-test.sh
```

在Windows环境下（PowerShell）：
```powershell
# 确保Docker服务已启动
# 1. 启动Docker中的MySQL服务
docker-compose up -d mysql
# 2. 等待MySQL服务启动（约3秒）
Start-Sleep -Seconds 3
# 3. 构建并启动应用（测试环境）
.\mvnw.cmd clean package -DskipTests
java -jar -Dspring.profiles.active=test target\mgtt-be-0.0.1-SNAPSHOT.jar
```

### 线上环境启动

线上环境使用默认的`application.properties`配置，包含以下特性：
- 禁用数据初始化
- 使用生产环境数据库配置

**执行步骤：**

在Linux/Mac环境下：
```bash
./start-prod.sh
```

在Windows环境下（PowerShell）：
```powershell
# 构建并启动应用
.\mvnw.cmd clean package -DskipTests
java -jar target\mgtt-be-0.0.1-SNAPSHOT.jar
```
