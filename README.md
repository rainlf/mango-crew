# MGTT-Go

一个使用 Go 语言重构的麻将计分系统后端服务，简洁、高效、优雅。

## 🎯 项目特点

- **简洁架构**：采用清晰的分层架构（Handler -> Service -> Repository）
- **高性能**：基于 Gin 框架，性能优异
- **优雅代码**：遵循 Go 语言最佳实践，代码简洁易读
- **完整功能**：支持多种麻将牌型的分数计算
- **易于部署**：单二进制文件，配置简单

## 🛠️ 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.23+ | 开发语言 |
| Gin | v1.10.0 | Web 框架 |
| GORM | v1.25.12 | ORM 框架 |
| MySQL | 8.x | 数据库 |
| Zap | v1.27.0 | 日志框架 |
| Viper | v1.19.0 | 配置管理 |

## 📁 项目结构

```
mgtt-go/
├── cmd/
│   └── main.go              # 应用入口
├── internal/
│   ├── config/              # 配置管理
│   │   └── config.go
│   ├── handler/             # HTTP 处理器
│   │   ├── user.go
│   │   ├── game.go
│   │   └── health.go
│   ├── middleware/          # 中间件
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   └── cors.go
│   ├── model/               # 数据模型
│   │   ├── user.go
│   │   ├── game.go
│   │   └── dto.go
│   ├── repository/          # 数据访问层
│   │   ├── user.go
│   │   └── game.go
│   └── service/             # 业务逻辑层
│       ├── user.go
│       └── game.go
├── pkg/                     # 公共包
│   ├── logger/              # 日志工具
│   │   ├── logger.go
│   │   └── field.go
│   └── response/            # 响应封装
│       └── response.go
├── configs/
│   └── config.yaml          # 配置文件
├── go.mod
└── README.md
```

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/rainlf/mgtt-go.git
cd mgtt-go
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

修改 `configs/config.yaml` 中的数据库配置：

```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: mgtt
```

### 4. 运行项目

```bash
go run cmd/main.go
```

或编译后运行：

```bash
go build -o mgtt-go cmd/main.go
./mgtt-go
```

## 📡 API 接口

### 用户相关

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/user/login` | 微信登录 |
| GET | `/api/user/info` | 获取用户信息 |
| POST | `/api/user/info` | 更新用户信息（含头像） |
| POST | `/api/user/username` | 更新用户名 |
| GET | `/api/user/rank` | 获取用户排名 |

### 麻将游戏相关

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/majiang/games` | 获取游戏记录列表 |
| GET | `/api/majiang/user/games` | 获取指定用户的游戏记录 |
| POST | `/api/majiang/game` | 保存一局麻将游戏 |
| DELETE | `/api/majiang/game` | 删除游戏记录 |
| GET | `/api/majiang/game/players` | 获取玩家列表 |

### 健康检查

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/health` | 服务健康检查 |

## 🎮 支持的麻将牌型

1. **平胡** (PingHu) - 1个赢家，1个输家
2. **自摸** (ZiMo) - 1个赢家，3个输家
3. **一炮双响** (YiPaoShuangXiang) - 2个赢家，1个输家
4. **一炮三响** (YiPaoSanXiang) - 3个赢家，1个输家
5. **相公** (XiangGong) - 3个赢家，1个输家
6. **运动** (YunDong) - 单人运动记录，对手为"银行"

## 🎴 支持的番型

| 番型 | 倍数 |
|------|------|
| 无花果 | 1 |
| 碰碰胡 | 2 |
| 一条龙 | 2 |
| 混一色 | 2 |
| 清一色 | 4 |
| 小七对 | 4 |
| 龙七对 | 8 |
| 大吊车 | 2 |
| 门前清 | 2 |
| 杠开花 | 2 |

## ⚙️ 配置说明

`configs/config.yaml`：

```yaml
server:
  port: 8080              # 服务端口
  mode: debug             # 运行模式: debug/release

database:
  host: localhost         # 数据库主机
  port: 3306             # 数据库端口
  username: root         # 数据库用户名
  password: 123456       # 数据库密码
  database: mgtt         # 数据库名
  max_idle_conns: 10     # 最大空闲连接数
  max_open_conns: 100    # 最大打开连接数

log:
  level: info            # 日志级别: debug/info/warn/error
  format: json           # 日志格式: json/console
  output: stdout         # 日志输出: stdout/file_path

wechat:
  app_id: your_app_id           # 微信小程序 AppID
  app_secret: your_app_secret   # 微信小程序 AppSecret
  login_url: https://api.weixin.qq.com/sns/jscode2session  # 微信登录接口
```

## 🧪 测试

运行测试：

```bash
go test ./...
```

## 📦 构建

构建 Linux 版本：

```bash
GOOS=linux GOARCH=amd64 go build -o mgtt-go-linux cmd/main.go
```

构建 Windows 版本：

```bash
GOOS=windows GOARCH=amd64 go build -o mgtt-go.exe cmd/main.go
```

构建 macOS 版本：

```bash
GOOS=darwin GOARCH=amd64 go build -o mgtt-go-mac cmd/main.go
```

## 📝 许可证

MIT License

## 🙏 致谢

原项目 [mgtt-server](https://github.com/rainlf/mgtt) 使用 Java Spring Boot 开发，本项目是其 Go 语言重构版本。
