# Telegram Bot - Go Version

这是 ChatGPT Telegram Bot 的 Go 语言实现版本，支持多种数据库后端（SQLite、MySQL、PostgreSQL）。

## 功能特性

- 支持多种 AI 提供商（OpenAI、Azure、Gemini、Anthropic、Workers AI 等）
- 支持流式响应输出
- 支持图片生成和图片识别
- 支持多语言（中文、英文、葡萄牙语、繁体中文）
- 支持群组聊天和私聊
- 支持插件系统
- 支持多种数据库后端（SQLite、MySQL、PostgreSQL）
- 支持用户设置权限控制
- 支持 Docker 部署

## 快速开始

### 本地运行

1. 安装 Go 1.21 或更高版本

2. 克隆项目并进入目录：
```bash
cd go_version
```

3. 安装依赖：
```bash
go mod download
```

4. 配置环境变量（创建 `.env` 文件或直接设置）：
```bash
export TELEGRAM_AVAILABLE_TOKENS="your_bot_token"
export OPENAI_API_KEY="your_openai_key"

# 数据库配置（可选，默认使用 SQLite）
export DB_PATH="./data/bot.db"
# 或使用其他数据库
# export DSN="mysql://user:password@tcp(localhost:3306)/dbname"
# export DSN="postgres://user:password@localhost:5432/dbname"
```

5. 运行程序：
```bash
go run ./cmd/bot
```

### Docker 部署

1. 构建镜像：
```bash
docker build -t telegram-bot-go .
```

2. 运行容器：
```bash
docker run -d \
  -p 8080:8080 \
  -e TELEGRAM_AVAILABLE_TOKENS="your_bot_token" \
  -e OPENAI_API_KEY="your_openai_key" \
  -v $(pwd)/data:/root/data \
  telegram-bot-go
```

## 环境变量配置

### 必需配置

- `TELEGRAM_AVAILABLE_TOKENS`: Telegram Bot Token（必需）
- `OPENAI_API_KEY`: OpenAI API Key（如果使用 OpenAI）

### 数据库配置

默认使用 SQLite，也可以配置使用 MySQL 或 PostgreSQL：

**SQLite（默认）**：
```bash
export DB_PATH="./data/bot.db"
```

**MySQL**：
```bash
export DSN="mysql://username:password@tcp(host:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
```

**PostgreSQL**：
```bash
export DSN="postgres://username:password@host:5432/database?sslmode=disable"
# 或
export DSN="postgresql://username:password@host:5432/database?sslmode=disable"
```

注意：如果同时设置了 `DSN` 和 `DB_PATH`，将优先使用 `DSN`。

### 权限控制配置

控制用户是否可以修改模型、温度等参数：

```bash
# 允许所有用户修改自己的配置（默认：true）
export ENABLE_USER_SETTING=true

# 仅允许管理员修改配置
export ENABLE_USER_SETTING=false

# 指定管理员用户 ID（Telegram User ID，逗号分隔）
export CHAT_ADMIN_KEY="123456789,987654321"
```

当 `ENABLE_USER_SETTING=false` 时：
- 只有管理员可以修改配置
- 普通用户看不到配置相关命令
- 所有用户使用全局配置

### 其他可选配置

- `AI_PROVIDER`: AI 提供商（默认：auto）
- `LANGUAGE`: 语言设置（默认：zh-cn）
- `PORT`: HTTP 服务器端口（默认：8080）
- `STREAM_MODE`: 是否启用流式输出（默认：true）
- `SAFE_MODE`: 是否启用安全模式（默认：true）

更多配置选项请参考 [配置文档](doc/CONFIG.md)。

## 项目结构

```
go_version/
├── cmd/
│   └── bot/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/                  # 配置管理和权限系统
│   ├── storage/                 # 数据库存储层（GORM）
│   ├── telegram/                # Telegram 集成
│   ├── agent/                   # AI Agent 系统
│   ├── i18n/                    # 国际化
│   └── plugin/                  # 插件系统
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## 开发指南

### 构建

基本构建：
```bash
go build -o bot ./cmd/bot
```

使用 Makefile 构建（推荐）：
```bash
make build
```

构建发布版本：
```bash
make build-release
```

构建多平台版本：
```bash
make build-all
```

### 测试

运行所有测试：
```bash
go test ./...
```

运行测试并生成覆盖率报告：
```bash
make test-coverage
```

### 代码格式化

```bash
go fmt ./...
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
