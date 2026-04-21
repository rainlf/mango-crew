# Mango Crew

Mango Crew 是一个基于 Go 开发的麻将战绩与积分服务，提供用户登录、场次管理、对局记录、积分结算和玩家排行等能力。

## 项目概览

- 技术栈：Go 1.23、Gin、GORM、MySQL、Viper、Zap
- 默认端口：`8080`
- 配置文件：推荐通过环境变量 `CONFIG_PATH=config.yaml` 指定
- 数据库名：`mango_crew`
- 健康检查：`/api/health`

## 功能列表

- 微信小程序登录，按 `code` 换取用户身份
- 用户资料查询、更新、排行榜、用户列表
- 场次创建、结束、列表查询、进行中场次查询
- 对局创建、结算、取消、按场次查询、最近对局查询
- 当前玩家列表与历史玩家列表查询
- 启动时自动执行表结构迁移

## 目录结构

```text
mango-crew/
├── cmd/main.go                # 应用入口
├── configs/                   # 配置目录
│   ├── config.yaml            # 默认配置
│   └── config.dev.yaml        # 开发环境示例配置
├── internal/
│   ├── config/                # 配置加载
│   ├── handler/               # HTTP 处理层
│   ├── middleware/            # 中间件
│   ├── model/                 # 模型与 DTO
│   ├── repository/            # 数据访问层
│   └── service/               # 业务逻辑层
├── migrations/init.sql        # 数据库初始化脚本
├── pkg/                       # 公共组件
├── scripts/start.sh           # 启动脚本
├── API.md                     # 接口说明
└── README.md
```

## 环境要求

- Go `1.23+`
- MySQL `8.x`
- 可用的微信小程序 `AppID` / `AppSecret`

## 快速开始

### 1. 安装依赖

```bash
cd mango-crew
go mod download
```

### 2. 初始化数据库

先创建数据库 `mango_crew`，再导入初始化脚本：

```bash
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS mango_crew CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -uroot -p mango_crew < migrations/init.sql
```

说明：

- `migrations/init.sql` 会创建核心表，并初始化 `win_type` 字典数据。
- 程序启动时也会执行 GORM `AutoMigrate`，用于补齐或同步表结构。

### 3. 配置文件与环境变量

推荐在项目根目录使用 `config.yaml`，并设置以下环境变量：

```bash
export CONFIG_PATH=config.yaml
export WECHAT_APP_ID=your_app_id
export WECHAT_APP_SECRET=your_app_secret
```

如果你暂时不想复制配置文件，程序也兼容当前仓库内的 `configs/config.yaml`：当 `CONFIG_PATH=config.yaml` 但根目录不存在该文件时，会自动回退到 `configs/config.yaml`。

可以直接复制一份本地配置：

```bash
cp configs/config.yaml config.yaml
```

示例配置如下：

```yaml
server:
  port: 8080
  mode: release

database:
  host: 127.0.0.1
  port: 3306
  username: mango_crew
  password: your_password
  database: mango_crew
  max_idle_conns: 10
  max_open_conns: 100

log:
  level: info
  format: json
  output: stdout

wechat:
  login_url: https://api.weixin.qq.com/sns/jscode2session
```

### 4. 启动服务

直接运行：

```bash
CONFIG_PATH=config.yaml go run ./cmd/main.go
```

或使用脚本启动：

```bash
./scripts/start.sh
```

如果需要构建 Linux 版本：

```bash
./scripts/start.sh build-linux
```

### 5. 验证服务

```bash
curl http://localhost:8080/api/health
```

返回 `OK` 表示服务启动成功。

## 数据库说明

项目默认使用 `mango_crew` 数据库，核心表如下：

| 表名 | 说明 |
| --- | --- |
| `user` | 用户信息，包含微信 `open_id`、昵称、头像等 |
| `game_session` | 一次牌局场次，包含创建者、状态、开始/结束时间 |
| `game` | 单盘对局记录 |
| `game_player` | 对局中的玩家记录、角色与积分 |
| `game_player_win_type` | 玩家命中的番型及倍数 |
| `win_type` | 番型字典表 |

积分相关说明：

- `game.type` 记录对局类型：平胡、自摸、一炮双响、一炮三响、相公、运动。
- `game_player.role` 记录玩家角色：赢家、输家、记录者。
- `final_points` 由基础分和番型倍数计算得出。
- `win_type` 初始化数据由 `migrations/init.sql` 导入。

## 接口概览

详细接口定义见 `API.md`。

| 模块 | 接口 |
| --- | --- |
| 健康检查 | `GET /api/health` |
| 用户 | `GET /api/user/login`、`GET /api/user/info`、`POST /api/user/update`、`GET /api/user/rank`、`GET /api/user/list` |
| 场次 | `POST /api/session`、`POST /api/session/end`、`GET /api/session/list`、`GET /api/session/active` |
| 对局 | `POST /api/game`、`POST /api/game/settle`、`POST /api/game/cancel`、`GET /api/game/list`、`GET /api/game/recent`、`GET /api/game/players` |

## 请求约定

- 用户身份通过请求头 `X-User-ID` 或查询参数 `userId` 传递。
- `POST /api/user/update` 使用表单提交，不是 JSON。
- 部分业务失败会返回 HTTP `200`，同时响应体中的 `code` 为非 `0`；参数错误通常返回 HTTP `400`。

## 测试与构建

运行测试：

```bash
go test ./...
```

本地构建：

```bash
go build -o mango-crew ./cmd/main.go
```

## 参考

- 接口文档：`API.md`
- 初始化 SQL：`migrations/init.sql`
- 默认配置：`configs/config.yaml`
