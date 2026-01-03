# 部署文档

本文档描述如何部署 Telegram Bot Go 版本。

## 目录

- [本地部署](#本地部署)
- [Docker 部署](#docker-部署)
- [服务器部署](#服务器部署)
- [反向代理配置](#反向代理配置)
- [云平台部署](#云平台部署)
- [监控和维护](#监控和维护)
- [故障排查](#故障排查)

## 本地部署

### 前置要求

- Go 1.21 或更高版本
- SQLite 3
- Git

### 步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd go_version
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**

复制示例配置文件：
```bash
cp .env.example .env
```

编辑 `.env` 文件，至少配置以下必需项：
```bash
# 必需配置
TELEGRAM_AVAILABLE_TOKENS="your_bot_token_from_botfather"
OPENAI_API_KEY="your_openai_api_key"

# 数据库配置（可选）
# 默认使用 SQLite
DB_PATH=./data/bot.db

# 或使用 MySQL
# DSN="mysql://user:password@tcp(localhost:3306)/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local"

# 或使用 PostgreSQL
# DSN="postgres://user:password@localhost:5432/telegram_bot?sslmode=disable"

# 权限控制（可选）
# 允许所有用户修改配置（默认：true）
ENABLE_USER_SETTING=true

# 指定管理员用户 ID（Telegram User ID，逗号分隔）
# CHAT_ADMIN_KEY="123456789,987654321"

# 其他可选配置
AI_PROVIDER=openai
LANGUAGE=zh-cn
PORT=8080
```

4. **创建数据目录**
```bash
mkdir -p data
```

5. **构建程序**

基本构建：
```bash
go build -o bot ./cmd/bot
```

或使用 Makefile（推荐）：
```bash
make build
```

构建发布版本（包含版本信息）：
```bash
make build-release
```

6. **运行程序**
```bash
./bot
```

或直接运行（不构建）：
```bash
go run ./cmd/bot
```

7. **设置 Webhook**

程序启动后，访问以下 URL 来设置 Telegram Webhook：
```
http://localhost:8080/init
```

这将自动配置 webhook 并设置机器人命令列表。

### 验证部署

1. 访问 `http://localhost:8080/` 查看欢迎页面
2. 在 Telegram 中向机器人发送 `/start` 命令
3. 发送消息测试 AI 对话功能

## Docker 部署

### 使用 Docker

1. **构建镜像**
```bash
docker build -t telegram-bot-go .
```

2. **运行容器**

基本运行（使用 SQLite）：
```bash
docker run -d \
  --name telegram-bot \
  -p 8080:8080 \
  -e TELEGRAM_AVAILABLE_TOKENS="your_bot_token" \
  -e OPENAI_API_KEY="your_openai_key" \
  -e DB_PATH="/app/data/bot.db" \
  -v $(pwd)/data:/app/data \
  --restart unless-stopped \
  telegram-bot-go
```

使用 MySQL：
```bash
docker run -d \
  --name telegram-bot \
  -p 8080:8080 \
  -e TELEGRAM_AVAILABLE_TOKENS="your_bot_token" \
  -e OPENAI_API_KEY="your_openai_key" \
  -e DSN="mysql://user:password@tcp(mysql-host:3306)/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local" \
  --restart unless-stopped \
  telegram-bot-go
```

使用 PostgreSQL：
```bash
docker run -d \
  --name telegram-bot \
  -p 8080:8080 \
  -e TELEGRAM_AVAILABLE_TOKENS="your_bot_token" \
  -e OPENAI_API_KEY="your_openai_key" \
  -e DSN="postgres://user:password@postgres-host:5432/telegram_bot?sslmode=disable" \
  --restart unless-stopped \
  telegram-bot-go
```

使用环境变量文件：
```bash
docker run -d \
  --name telegram-bot \
  -p 8080:8080 \
  --env-file .env \
  -v $(pwd)/data:/app/data \
  --restart unless-stopped \
  telegram-bot-go
```

配置权限控制：
```bash
docker run -d \
  --name telegram-bot \
  -p 8080:8080 \
  -e TELEGRAM_AVAILABLE_TOKENS="your_bot_token" \
  -e OPENAI_API_KEY="your_openai_key" \
  -e ENABLE_USER_SETTING="false" \
  -e CHAT_ADMIN_KEY="123456789,987654321" \
  -v $(pwd)/data:/app/data \
  --restart unless-stopped \
  telegram-bot-go
```

3. **管理容器**

查看日志：
```bash
docker logs -f telegram-bot
```

停止容器：
```bash
docker stop telegram-bot
```

重启容器：
```bash
docker restart telegram-bot
```

删除容器：
```bash
docker stop telegram-bot
docker rm telegram-bot
```

### 使用 Docker Compose（推荐）

Docker Compose 可以方便地管理多个容器，特别是当使用外部数据库时。

#### 使用 SQLite（默认）

1. **配置环境变量**

复制示例配置：
```bash
cp .env.example .env
```

编辑 `.env` 文件配置必需的环境变量。

2. **启动服务**
```bash
docker-compose up -d
```

#### 使用 MySQL

1. **创建 docker-compose.mysql.yml**

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: telegram-bot-mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: telegram_bot
      MYSQL_USER: botuser
      MYSQL_PASSWORD: botpassword
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  bot:
    build: .
    container_name: telegram-bot
    depends_on:
      mysql:
        condition: service_healthy
    environment:
      TELEGRAM_AVAILABLE_TOKENS: "${TELEGRAM_AVAILABLE_TOKENS}"
      OPENAI_API_KEY: "${OPENAI_API_KEY}"
      DSN: "mysql://botuser:botpassword@tcp(mysql:3306)/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local"
      ENABLE_USER_SETTING: "${ENABLE_USER_SETTING:-true}"
      CHAT_ADMIN_KEY: "${CHAT_ADMIN_KEY}"
    ports:
      - "8080:8080"
    restart: unless-stopped

volumes:
  mysql_data:
```

2. **启动服务**
```bash
docker-compose -f docker-compose.mysql.yml up -d
```

#### 使用 PostgreSQL

1. **创建 docker-compose.postgres.yml**

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    container_name: telegram-bot-postgres
    environment:
      POSTGRES_DB: telegram_bot
      POSTGRES_USER: botuser
      POSTGRES_PASSWORD: botpassword
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U botuser"]
      interval: 10s
      timeout: 5s
      retries: 5

  bot:
    build: .
    container_name: telegram-bot
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      TELEGRAM_AVAILABLE_TOKENS: "${TELEGRAM_AVAILABLE_TOKENS}"
      OPENAI_API_KEY: "${OPENAI_API_KEY}"
      DSN: "postgres://botuser:botpassword@postgres:5432/telegram_bot?sslmode=disable"
      ENABLE_USER_SETTING: "${ENABLE_USER_SETTING:-true}"
      CHAT_ADMIN_KEY: "${CHAT_ADMIN_KEY}"
    ports:
      - "8080:8080"
    restart: unless-stopped

volumes:
  postgres_data:
```

2. **启动服务**
```bash
docker-compose -f docker-compose.postgres.yml up -d
```

#### 管理服务

3. **查看日志**
```bash
docker-compose logs -f

# 只查看 bot 日志
docker-compose logs -f bot

# 只查看数据库日志
docker-compose logs -f mysql
# 或
docker-compose logs -f postgres
```

4. **停止服务**
```bash
docker-compose down

# 停止并删除数据卷（注意：会删除数据库数据）
docker-compose down -v
```

5. **重启服务**
```bash
docker-compose restart
```

6. **更新服务**
```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

### Docker 镜像优化

构建的 Docker 镜像已经过优化：

- **多阶段构建**：分离构建和运行环境
- **最小化基础镜像**：使用 Alpine Linux
- **非 root 用户**：提高安全性
- **健康检查**：自动监控服务状态
- **体积优化**：去除调试信息，最终镜像约 20-30MB

## 服务器部署

### 使用 systemd（推荐用于 Linux 服务器）

1. **准备部署目录**
```bash
sudo mkdir -p /opt/telegram-bot
sudo mkdir -p /opt/telegram-bot/data
```

2. **复制程序文件**
```bash
# 构建程序
make build-release

# 复制到部署目录
sudo cp bot /opt/telegram-bot/
sudo chmod +x /opt/telegram-bot/bot
```

3. **创建专用用户**
```bash
sudo useradd -r -s /bin/false telegram-bot
sudo chown -R telegram-bot:telegram-bot /opt/telegram-bot
```

4. **创建环境变量文件**
```bash
sudo nano /opt/telegram-bot/.env
```

添加配置（参考 `.env.example`）。

5. **创建 systemd 服务文件**

创建 `/etc/systemd/system/telegram-bot.service`：
```ini
[Unit]
Description=Telegram Bot Go
Documentation=https://github.com/your-repo/telegram-bot-go
After=network.target

[Service]
Type=simple
User=telegram-bot
Group=telegram-bot
WorkingDirectory=/opt/telegram-bot

# Load environment variables from file
EnvironmentFile=/opt/telegram-bot/.env

# Run the bot
ExecStart=/opt/telegram-bot/bot

# Restart policy
Restart=on-failure
RestartSec=10
StartLimitInterval=60
StartLimitBurst=3

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/telegram-bot/data

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=telegram-bot

[Install]
WantedBy=multi-user.target
```

6. **启动服务**
```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable telegram-bot

# 启动服务
sudo systemctl start telegram-bot
```

7. **管理服务**

查看状态：
```bash
sudo systemctl status telegram-bot
```

查看日志：
```bash
# 实时日志
sudo journalctl -u telegram-bot -f

# 最近 100 行
sudo journalctl -u telegram-bot -n 100

# 今天的日志
sudo journalctl -u telegram-bot --since today
```

重启服务：
```bash
sudo systemctl restart telegram-bot
```

停止服务：
```bash
sudo systemctl stop telegram-bot
```

### 使用 Supervisor

1. **安装 Supervisor**
```bash
sudo apt-get install supervisor  # Debian/Ubuntu
sudo yum install supervisor      # CentOS/RHEL
```

2. **创建配置文件**

创建 `/etc/supervisor/conf.d/telegram-bot.conf`：
```ini
[program:telegram-bot]
command=/opt/telegram-bot/bot
directory=/opt/telegram-bot
user=telegram-bot
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/telegram-bot/bot.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=10
environment=TELEGRAM_AVAILABLE_TOKENS="your_token",OPENAI_API_KEY="your_key",DB_PATH="/opt/telegram-bot/data/bot.db"
```

3. **启动服务**
```bash
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start telegram-bot
```

## 反向代理配置

### Nginx

1. **安装 Nginx**
```bash
sudo apt-get install nginx  # Debian/Ubuntu
```

2. **配置站点**

创建 `/etc/nginx/sites-available/telegram-bot`：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Logging
    access_log /var/log/nginx/telegram-bot-access.log;
    error_log /var/log/nginx/telegram-bot-error.log;

    # Proxy settings
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

3. **启用站点**
```bash
sudo ln -s /etc/nginx/sites-available/telegram-bot /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

4. **配置 SSL（使用 Let's Encrypt）**
```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

### Caddy（更简单的选择）

1. **安装 Caddy**
```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

2. **配置 Caddyfile**

编辑 `/etc/caddy/Caddyfile`：
```
your-domain.com {
    reverse_proxy localhost:8080
    
    # Automatic HTTPS
    tls your-email@example.com
    
    # Logging
    log {
        output file /var/log/caddy/telegram-bot.log
    }
}
```

3. **重启 Caddy**
```bash
sudo systemctl reload caddy
```

Caddy 会自动处理 SSL 证书申请和续期。

## 云平台部署

### Railway

1. **准备项目**
   - 确保项目已推送到 GitHub

2. **部署步骤**
   - 访问 [Railway.app](https://railway.app)
   - 点击 "New Project"
   - 选择 "Deploy from GitHub repo"
   - 选择你的仓库
   - Railway 会自动检测 Dockerfile 并构建

3. **配置环境变量**
   - 在 Railway 项目设置中添加环境变量
   - 至少配置 `TELEGRAM_AVAILABLE_TOKENS` 和 AI 提供商密钥
   - 可选配置数据库：
     - 使用内置 SQLite：设置 `DB_PATH=/app/data/bot.db`
     - 使用 Railway PostgreSQL：添加 PostgreSQL 插件，然后设置 `DSN=${{Postgres.DATABASE_URL}}`
     - 使用 Railway MySQL：添加 MySQL 插件，然后设置 `DSN=${{MySQL.DATABASE_URL}}`
   - 可选配置权限控制：
     - `ENABLE_USER_SETTING=false`（仅管理员可修改配置）
     - `CHAT_ADMIN_KEY=123456789,987654321`（管理员用户 ID）

4. **配置域名**
   - 在 Settings 中生成域名或绑定自定义域名

5. **设置 Webhook**
   - 访问 `https://your-app.railway.app/init`

### Render

1. **创建 Web Service**
   - 访问 [Render.com](https://render.com)
   - 点击 "New +" → "Web Service"
   - 连接 GitHub 仓库

2. **配置构建**
   - Build Command: `go build -o bot ./cmd/bot`
   - Start Command: `./bot`
   - 或者选择 "Docker" 并使用 Dockerfile

3. **配置环境变量**
   - 在 Environment 标签页添加所有必需的环境变量
   - 数据库配置：
     - Render 提供免费的 PostgreSQL 数据库
     - 创建 PostgreSQL 数据库后，在环境变量中设置 `DSN`
     - 或使用内置 SQLite（注意：Render 重启后文件会丢失）
   - 权限控制配置：
     - `ENABLE_USER_SETTING`
     - `CHAT_ADMIN_KEY`

4. **部署**
   - 点击 "Create Web Service"
   - Render 会自动构建和部署

### Fly.io

1. **安装 flyctl**
```bash
curl -L https://fly.io/install.sh | sh
```

2. **登录**
```bash
flyctl auth login
```

3. **初始化项目**
```bash
cd go_version
flyctl launch
```

4. **配置 fly.toml**

编辑生成的 `fly.toml`：
```toml
app = "telegram-bot-go"
primary_region = "hkg"  # 选择离你最近的区域

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8080"
  DB_PATH = "/data/bot.db"

[[mounts]]
  source = "bot_data"
  destination = "/data"

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [[services.ports]]
    handlers = ["http"]
    port = 80

  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443

  [[services.http_checks]]
    interval = "30s"
    timeout = "5s"
    grace_period = "10s"
    method = "GET"
    path = "/"
```

5. **设置环境变量**
```bash
flyctl secrets set TELEGRAM_AVAILABLE_TOKENS="your_token"
flyctl secrets set OPENAI_API_KEY="your_key"

# 可选：配置权限控制
flyctl secrets set ENABLE_USER_SETTING="false"
flyctl secrets set CHAT_ADMIN_KEY="123456789,987654321"

# 可选：使用外部数据库
# flyctl secrets set DSN="postgres://user:pass@host:5432/dbname"
```

注意：Fly.io 的持久化存储使用 volumes，适合 SQLite。如果需要使用 MySQL 或 PostgreSQL，建议使用外部数据库服务。

6. **创建持久化存储**
```bash
flyctl volumes create bot_data --size 1
```

7. **部署**
```bash
flyctl deploy
```

8. **查看日志**
```bash
flyctl logs
```

### Heroku

1. **安装 Heroku CLI**
```bash
curl https://cli-assets.heroku.com/install.sh | sh
```

2. **登录**
```bash
heroku login
```

3. **创建应用**
```bash
heroku create your-app-name
```

4. **设置 Buildpack**
```bash
heroku buildpacks:set heroku/go
```

5. **配置环境变量**
```bash
heroku config:set TELEGRAM_AVAILABLE_TOKENS="your_token"
heroku config:set OPENAI_API_KEY="your_key"

# 数据库配置
# Heroku 提供 PostgreSQL 插件
heroku addons:create heroku-postgresql:mini
# 数据库 URL 会自动设置为 DATABASE_URL
# 需要在代码中将 DATABASE_URL 映射到 DSN

# 或使用 JawsDB MySQL
heroku addons:create jawsdb:kitefin
# MySQL URL 会设置为 JAWSDB_URL

# 权限控制配置
heroku config:set ENABLE_USER_SETTING="false"
heroku config:set CHAT_ADMIN_KEY="123456789,987654321"
```

注意：Heroku 的文件系统是临时的，重启后数据会丢失。强烈建议使用 Heroku PostgreSQL 或其他数据库插件。

## 监控和维护

### 健康检查

程序提供了内置的健康检查端点：

```bash
# 检查服务状态
curl http://localhost:8080/

# 应该返回欢迎页面和版本信息
```

### 日志管理

#### systemd 日志

```bash
# 实时查看日志
sudo journalctl -u telegram-bot -f

# 查看最近的错误
sudo journalctl -u telegram-bot -p err

# 查看特定时间段的日志
sudo journalctl -u telegram-bot --since "2024-01-01" --until "2024-01-02"

# 导出日志
sudo journalctl -u telegram-bot > telegram-bot.log
```

#### Docker 日志

```bash
# 查看日志
docker logs telegram-bot

# 实时查看
docker logs -f telegram-bot

# 查看最近 100 行
docker logs --tail 100 telegram-bot

# 查看特定时间段
docker logs --since 1h telegram-bot
```

### 数据备份

#### 手动备份

```bash
# 停止服务（可选，但推荐）
sudo systemctl stop telegram-bot

# 备份数据库
cp /opt/telegram-bot/data/bot.db /opt/telegram-bot/backups/bot.db.$(date +%Y%m%d_%H%M%S)

# 重启服务
sudo systemctl start telegram-bot
```

#### 自动备份脚本

创建 `/opt/telegram-bot/backup.sh`：
```bash
#!/bin/bash

BACKUP_DIR="/opt/telegram-bot/backups"
DB_PATH="/opt/telegram-bot/data/bot.db"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/bot.db.$TIMESTAMP"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份数据库
cp "$DB_PATH" "$BACKUP_FILE"

# 压缩备份
gzip "$BACKUP_FILE"

# 删除 30 天前的备份
find "$BACKUP_DIR" -name "bot.db.*.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

设置定时任务：
```bash
# 编辑 crontab
sudo crontab -e

# 添加每天凌晨 2 点备份
0 2 * * * /opt/telegram-bot/backup.sh >> /var/log/telegram-bot-backup.log 2>&1
```

### 性能监控

#### 使用 Prometheus 和 Grafana

1. **添加 Prometheus 指标**（需要在代码中实现）

2. **配置 Prometheus**

`prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'telegram-bot'
    static_configs:
      - targets: ['localhost:8080']
```

3. **配置 Grafana 仪表板**
   - 添加 Prometheus 数据源
   - 创建仪表板监控关键指标

### 更新和升级

#### 本地部署更新

```bash
# 拉取最新代码
git pull

# 重新构建
make build-release

# 重启服务
sudo systemctl restart telegram-bot
```

#### Docker 部署更新

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

## 故障排查

### 常见问题

#### 1. 无法连接到 Telegram API

**症状**：
- 日志显示连接超时
- 无法接收或发送消息

**解决方案**：
```bash
# 检查网络连接
curl https://api.telegram.org

# 如果在受限地区，配置自定义 API 域名
export TELEGRAM_API_DOMAIN="https://your-proxy-domain.com"
```

#### 2. 数据库锁定错误

**症状**：
- 日志显示 "database is locked"
- 操作失败

**解决方案**：
```bash
# 检查是否有多个实例运行
ps aux | grep bot

# 停止所有实例
sudo systemctl stop telegram-bot

# 检查数据库文件权限
ls -la /opt/telegram-bot/data/

# 修复权限
sudo chown telegram-bot:telegram-bot /opt/telegram-bot/data/bot.db

# 重启服务
sudo systemctl start telegram-bot
```

#### 3. 内存占用过高

**症状**：
- 系统内存不足
- 服务被 OOM killer 终止

**解决方案**：
```bash
# 调整历史记录配置
export MAX_HISTORY_LENGTH=10
export AUTO_TRIM_HISTORY=true

# 限制 Docker 容器内存
docker run -d --memory="512m" ...
```

#### 4. Webhook 设置失败

**症状**：
- 访问 /init 返回错误
- 无法接收消息

**解决方案**：
```bash
# 检查 Bot Token 是否正确
echo $TELEGRAM_AVAILABLE_TOKENS

# 检查域名是否可访问
curl https://your-domain.com/

# 手动设置 webhook
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" \
  -d "url=https://your-domain.com/telegram/<YOUR_BOT_TOKEN>/webhook"
```

#### 5. AI 提供商 API 错误

**症状**：
- 无法获取 AI 响应
- 日志显示 API 错误

**解决方案**：
```bash
# 检查 API 密钥
echo $OPENAI_API_KEY

# 测试 API 连接
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# 切换到其他提供商
export AI_PROVIDER=gemini
export GOOGLE_API_KEY="your_google_key"
```

### 调试模式

启用调试模式获取更多信息：

```bash
# 启用调试模式
export DEBUG_MODE=true
export DEV_MODE=true

# 重启服务
sudo systemctl restart telegram-bot

# 查看详细日志
sudo journalctl -u telegram-bot -f
```

### 性能优化建议

1. **数据库优化**
   - SQLite 已默认启用 WAL 模式
   - 定期执行 VACUUM 清理数据库

2. **历史记录管理**
   - 合理设置 `MAX_HISTORY_LENGTH`
   - 启用 `AUTO_TRIM_HISTORY`
   - 使用 `HISTORY_IMAGE_PLACEHOLDER` 减少存储

3. **网络优化**
   - 使用 CDN 加速
   - 配置反向代理缓存
   - 启用 HTTP/2

4. **资源限制**
   - 使用 systemd 限制资源使用
   - Docker 容器设置内存和 CPU 限制

## 安全建议

1. **环境变量安全**
   - 不要在代码中硬编码密钥
   - 使用 `.env` 文件并添加到 `.gitignore`
   - 在生产环境使用密钥管理服务

2. **访问控制**
   - 配置白名单限制访问
   - 不要启用 `I_AM_A_GENEROUS_PERSON` 在生产环境
   - 定期审查访问日志

3. **网络安全**
   - 使用 HTTPS
   - 配置防火墙规则
   - 使用反向代理隐藏真实 IP

4. **系统安全**
   - 使用非 root 用户运行
   - 定期更新系统和依赖
   - 启用 SELinux 或 AppArmor

5. **数据安全**
   - 定期备份数据
   - 加密敏感数据
   - 设置合理的文件权限

## 支持和帮助

如果遇到问题：

1. 查看日志获取错误信息
2. 参考本文档的故障排查部分
3. 查看项目 Issues
4. 提交新的 Issue 并附上详细信息

## 相关文档

- [配置文档](CONFIG.md) - 详细的配置选项说明
- [README](../README.md) - 项目概述和快速开始

## 数据库部署最佳实践

### 选择合适的数据库

**SQLite**：
- 适合：小型部署、单机部署、开发测试
- 优点：零配置、轻量级、易于备份
- 缺点：不支持高并发、不适合分布式部署

**MySQL**：
- 适合：中大型部署、需要高可用性
- 优点：成熟稳定、工具丰富、性能好
- 缺点：需要额外的数据库服务器

**PostgreSQL**：
- 适合：需要高级功能、复杂查询
- 优点：功能强大、标准兼容性好、扩展性强
- 缺点：相对复杂、资源占用较高

### 数据库连接字符串示例

**MySQL**：
```bash
# 基本格式
DSN="mysql://username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local"

# 本地 MySQL
DSN="mysql://root:password@tcp(localhost:3306)/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local"

# 云数据库（如 AWS RDS）
DSN="mysql://admin:password@tcp(mydb.abc123.us-east-1.rds.amazonaws.com:3306)/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local"

# 使用 SSL
DSN="mysql://user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local&tls=true"
```

**PostgreSQL**：
```bash
# 基本格式
DSN="postgres://username:password@host:port/database?sslmode=disable"

# 本地 PostgreSQL
DSN="postgres://postgres:password@localhost:5432/telegram_bot?sslmode=disable"

# 云数据库（如 Heroku Postgres）
DSN="postgres://user:pass@ec2-xxx.compute-1.amazonaws.com:5432/dbname?sslmode=require"

# 使用 postgresql:// 前缀也可以
DSN="postgresql://user:pass@host:5432/db?sslmode=disable"
```

**SQLite**：
```bash
# 使用 DB_PATH（推荐）
DB_PATH="./data/bot.db"

# 或使用 DSN
DSN="sqlite://./data/bot.db"
```

### 数据库迁移

如果需要从 SQLite 迁移到 MySQL 或 PostgreSQL：

1. **使用 GORM 自动迁移（推荐）**

GORM 会自动创建表结构，只需更改 DSN 配置即可：

```bash
# 停止服务
sudo systemctl stop telegram-bot

# 更新 .env 文件中的 DSN
# 从：DB_PATH="./data/bot.db"
# 改为：DSN="mysql://user:pass@host:3306/telegram_bot?charset=utf8mb4&parseTime=True&loc=Local"

# 启动服务（GORM 会自动创建表）
sudo systemctl start telegram-bot
```

注意：这种方式会创建新的空数据库，不会迁移旧数据。

2. **手动迁移数据（如需保留旧数据）**

```bash
# 导出 SQLite 数据
sqlite3 bot.db .dump > backup.sql

# 转换并导入到新数据库
# MySQL
mysql -u username -p database_name < backup.sql

# PostgreSQL
psql -U username -d database_name -f backup.sql
```

注意：可能需要调整 SQL 语法以适配目标数据库。

### 数据库性能优化

**连接池配置**（在代码中已配置）：
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

**索引优化**：
- GORM 已自动创建必要的索引
- 根据查询模式可添加额外索引

**定期维护**：
```bash
# MySQL
mysql -u root -p -e "OPTIMIZE TABLE telegram_bot.chat_histories, telegram_bot.user_configurations;"

# PostgreSQL
psql -U postgres -d telegram_bot -c "VACUUM ANALYZE;"

# SQLite
sqlite3 bot.db "VACUUM;"
```

### 数据库备份

**SQLite**：
```bash
# 简单复制
cp /opt/telegram-bot/data/bot.db /backup/bot.db.$(date +%Y%m%d)

# 使用 SQLite 备份命令
sqlite3 /opt/telegram-bot/data/bot.db ".backup '/backup/bot.db.$(date +%Y%m%d)'"
```

**MySQL**：
```bash
# 使用 mysqldump
mysqldump -u username -p telegram_bot > backup_$(date +%Y%m%d).sql

# 压缩备份
mysqldump -u username -p telegram_bot | gzip > backup_$(date +%Y%m%d).sql.gz
```

**PostgreSQL**：
```bash
# 使用 pg_dump
pg_dump -U username telegram_bot > backup_$(date +%Y%m%d).sql

# 压缩备份
pg_dump -U username telegram_bot | gzip > backup_$(date +%Y%m%d).sql.gz
```

## 权限控制部署

### 配置管理员权限

1. **获取 Telegram User ID**

向你的 bot 发送消息，然后查看日志获取 User ID：
```bash
# 查看日志
sudo journalctl -u telegram-bot -f

# 或在代码中添加日志输出
```

或使用 [@userinfobot](https://t.me/userinfobot) 获取你的 User ID。

2. **配置管理员列表**

编辑 `.env` 文件：
```bash
# 单个管理员
CHAT_ADMIN_KEY="123456789"

# 多个管理员（逗号分隔）
CHAT_ADMIN_KEY="123456789,987654321,555666777"
```

3. **启用权限控制**

```bash
# 仅管理员可修改配置
ENABLE_USER_SETTING=false

# 所有用户可修改配置（默认）
ENABLE_USER_SETTING=true
```

4. **重启服务**
```bash
sudo systemctl restart telegram-bot
```

### 权限控制场景

**场景 1：公共 Bot（默认）**
```bash
ENABLE_USER_SETTING=true
# CHAT_ADMIN_KEY 不需要设置
```
- 所有用户可以修改自己的配置
- 适合个人使用或小团队

**场景 2：企业 Bot**
```bash
ENABLE_USER_SETTING=false
CHAT_ADMIN_KEY="123456789,987654321"
```
- 只有管理员可以修改配置
- 所有用户使用统一配置
- 普通用户看不到配置命令
- 适合企业统一管理

**场景 3：混合模式**
```bash
ENABLE_USER_SETTING=true
CHAT_ADMIN_KEY="123456789"
```
- 所有用户可以修改自己的配置
- 管理员有额外权限（如果未来添加管理功能）
