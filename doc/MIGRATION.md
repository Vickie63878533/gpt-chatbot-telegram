# 迁移指南

本文档说明如何从旧版本的 Telegram Bot 迁移到支持 SillyTavern 集成的新版本。

## 概述

新版本引入了以下主要变更：

1. **SillyTavern 集成**：角色卡、世界书、预设和正则处理
2. **Web 管理器**：用于管理 SillyTavern 功能的 Web 界面
3. **智能上下文管理**：自动摘要和上下文窗口管理
4. **对话分享**：通过 Telegraph 分享对话
5. **数据库模型扩展**：新增多个表用于存储 SillyTavern 数据
6. **配置变更**：新增和移除部分环境变量

## 兼容性

- ✅ **向后兼容**：现有功能保持不变
- ✅ **数据保留**：现有对话历史自动迁移
- ✅ **可选功能**：SillyTavern 功能可选启用
- ⚠️ **配置变更**：部分环境变量已更改

## 迁移步骤

### 步骤 1: 备份数据

在升级之前，**强烈建议**备份你的数据库：

#### SQLite 备份

```bash
# 停止 bot
docker-compose down

# 备份数据库文件
cp data/bot.db data/bot.db.backup

# 或使用 tar 打包整个 data 目录
tar -czf data-backup-$(date +%Y%m%d).tar.gz data/
```

#### MySQL 备份

```bash
mysqldump -u username -p database_name > backup-$(date +%Y%m%d).sql
```

#### PostgreSQL 备份

```bash
pg_dump -U username database_name > backup-$(date +%Y%m%d).sql
```

### 步骤 2: 更新代码

拉取最新代码：

```bash
git pull origin main
```

或下载最新的 release 版本。

### 步骤 3: 更新环境变量

#### 新增的环境变量

在你的 `.env` 文件或环境配置中添加以下变量：

```bash
# ============================================
# SillyTavern 集成配置
# ============================================

# Web 管理器配置
MANAGER_ENABLED=true              # 是否启用 Web 管理器
MANAGER_PORT=8081                 # 管理器端口

# Telegraph 分享配置
TELEGRAPH_ENABLED=true            # 是否启用 Telegraph 分享功能

# 上下文管理配置
MAX_CONTEXT_LENGTH=8000           # 最大上下文长度（tokens）
SUMMARY_THRESHOLD=0.8             # 触发摘要的阈值（0.0-1.0）
MIN_RECENT_PAIRS=2                # 保留的最小最近消息对数
```

#### 移除的环境变量

以下环境变量已被移除或替换：

- ~~`SYSTEM_INIT_MESSAGE`~~ - 现在使用角色卡的 system_prompt

如果你之前使用了 `SYSTEM_INIT_MESSAGE`，可以：
1. 创建一个角色卡并在其中设置 system_prompt
2. 或者保留该变量作为后备（如果没有激活角色卡，系统仍会使用它）

#### 配置示例

完整的 `.env` 配置示例：

```bash
# 必需配置
TELEGRAM_AVAILABLE_TOKENS=your_bot_token
OPENAI_API_KEY=your_openai_key

# 数据库配置（保持不变）
DB_PATH=./data/bot.db

# 权限控制（保持不变）
ENABLE_USER_SETTING=true
CHAT_ADMIN_KEY=123456789

# 新增：SillyTavern 配置
MANAGER_ENABLED=true
MANAGER_PORT=8081
TELEGRAPH_ENABLED=true
MAX_CONTEXT_LENGTH=8000
SUMMARY_THRESHOLD=0.8
MIN_RECENT_PAIRS=2

# 其他配置（保持不变）
LANGUAGE=zh-cn
STREAM_MODE=true
PORT=8080
```

### 步骤 4: 更新 Docker 配置

如果使用 Docker 部署，需要更新 `docker-compose.yml`：

#### 添加管理器端口映射

```yaml
services:
  telegram-bot:
    ports:
      - "${PORT:-8080}:8080"
      - "${MANAGER_PORT:-8081}:8081"  # 新增：管理器端口
```

#### 添加新的环境变量

```yaml
    environment:
      # ... 现有配置 ...
      
      # 新增：SillyTavern 配置
      - MANAGER_ENABLED=${MANAGER_ENABLED:-true}
      - MANAGER_PORT=8081
      - TELEGRAPH_ENABLED=${TELEGRAPH_ENABLED:-true}
      - MAX_CONTEXT_LENGTH=${MAX_CONTEXT_LENGTH:-8000}
      - SUMMARY_THRESHOLD=${SUMMARY_THRESHOLD:-0.8}
      - MIN_RECENT_PAIRS=${MIN_RECENT_PAIRS:-2}
```

完整的 `docker-compose.yml` 示例请参考项目根目录的文件。

### 步骤 5: 数据库迁移

数据库迁移是**自动**的！当你启动新版本时，GORM 会自动创建新表：

- `character_cards` - 角色卡
- `world_books` - 世界书
- `world_book_entries` - 世界书条目
- `presets` - 预设
- `regex_patterns` - 正则模式
- `login_tokens` - 登录令牌

现有的 `chat_histories` 表会自动添加新字段：
- `timestamp` - 消息时间戳
- `truncated` - 是否被截断

**注意**：
- 现有数据不会丢失
- 新字段会自动添加默认值
- 迁移过程通常在几秒内完成

#### 验证迁移

启动 bot 后，检查日志确认迁移成功：

```bash
docker-compose logs -f telegram-bot
```

你应该看到类似的日志：

```
[INFO] Database migration completed successfully
[INFO] Created table: character_cards
[INFO] Created table: world_books
[INFO] Created table: world_book_entries
[INFO] Created table: presets
[INFO] Created table: regex_patterns
[INFO] Created table: login_tokens
```

### 步骤 6: 启动服务

#### 使用 Docker

```bash
# 重新构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

#### 本地运行

```bash
# 安装依赖
go mod download

# 运行
go run ./cmd/bot
```

### 步骤 7: 验证功能

#### 1. 验证基本功能

发送消息给 bot，确认正常响应。

#### 2. 验证 /login 命令

在私聊中发送：
```
/login
```

应该收到包含用户名和密码的响应。

#### 3. 访问 Web 管理器

打开浏览器访问：
```
http://localhost:8081
```

使用 `/login` 获取的凭据登录。

#### 4. 测试对话分享

发送几条消息后，执行：
```
/share
```

应该收到 Telegraph 链接。

## 常见问题

### Q1: 升级后 bot 无法启动

**可能原因**：
- 环境变量配置错误
- 数据库连接失败
- 端口冲突

**解决方法**：
1. 检查日志：`docker-compose logs telegram-bot`
2. 验证环境变量配置
3. 确认端口 8080 和 8081 未被占用
4. 检查数据库文件权限

### Q2: 数据库迁移失败

**可能原因**：
- 数据库文件损坏
- 权限不足
- 磁盘空间不足

**解决方法**：
1. 恢复备份：`cp data/bot.db.backup data/bot.db`
2. 检查文件权限：`ls -la data/`
3. 检查磁盘空间：`df -h`
4. 手动运行迁移（如果需要）

### Q3: 无法访问 Web 管理器

**可能原因**：
- MANAGER_ENABLED 未设置为 true
- 端口映射错误
- 防火墙阻止

**解决方法**：
1. 确认 `MANAGER_ENABLED=true`
2. 检查端口映射：`docker-compose ps`
3. 测试端口：`curl http://localhost:8081`
4. 检查防火墙规则

### Q4: /login 命令不工作

**可能原因**：
- 在群组中使用（仅支持私聊）
- 数据库写入失败

**解决方法**：
1. 确保在私聊中使用
2. 检查数据库权限
3. 查看日志了解详细错误

### Q5: 现有对话历史丢失

**不应该发生**！如果发生：
1. 立即停止 bot
2. 恢复备份：`cp data/bot.db.backup data/bot.db`
3. 检查日志找出原因
4. 联系支持

### Q6: SYSTEM_INIT_MESSAGE 不再工作

这是预期行为。现在系统提示来自角色卡。

**解决方法**：
1. 创建一个角色卡
2. 在角色卡的 `system_prompt` 字段中设置你的系统提示
3. 激活该角色卡

或者，如果没有激活角色卡，系统仍会使用 `SYSTEM_INIT_MESSAGE` 作为后备。

### Q7: 自动摘要触发太频繁

**解决方法**：
调整配置：
```bash
MAX_CONTEXT_LENGTH=16000      # 增加上下文长度
SUMMARY_THRESHOLD=0.9         # 提高触发阈值
MIN_RECENT_PAIRS=5            # 保留更多最近消息
```

### Q8: 想要禁用 SillyTavern 功能

完全可以！设置：
```bash
MANAGER_ENABLED=false
TELEGRAPH_ENABLED=false
```

Bot 将继续正常工作，只是没有 SillyTavern 功能。

## 回滚到旧版本

如果遇到严重问题需要回滚：

### 步骤 1: 停止服务

```bash
docker-compose down
```

### 步骤 2: 恢复备份

```bash
cp data/bot.db.backup data/bot.db
```

### 步骤 3: 切换到旧版本

```bash
git checkout <old-version-tag>
# 或
docker pull <old-image-tag>
```

### 步骤 4: 启动旧版本

```bash
docker-compose up -d
```

**注意**：回滚后，新版本创建的数据（角色卡、世界书等）将不可用，但不会影响旧版本的功能。

## 性能优化建议

### 1. 数据库优化

对于大量历史记录，考虑：

```bash
# 定期清理旧的登录令牌（自动进行）
# 定期备份和压缩数据库
sqlite3 data/bot.db "VACUUM;"
```

### 2. 上下文管理优化

根据你的使用情况调整：

```bash
# 对于短对话
MAX_CONTEXT_LENGTH=4000
SUMMARY_THRESHOLD=0.7
MIN_RECENT_PAIRS=2

# 对于长对话
MAX_CONTEXT_LENGTH=16000
SUMMARY_THRESHOLD=0.9
MIN_RECENT_PAIRS=5
```

### 3. 内存优化

如果内存有限：

```bash
# 减少上下文长度
MAX_CONTEXT_LENGTH=4000

# 更积极地触发摘要
SUMMARY_THRESHOLD=0.6
```

## 获取帮助

如果遇到问题：

1. 查看日志：`docker-compose logs -f telegram-bot`
2. 检查 [配置文档](CONFIG.md)
3. 查看 [部署文档](DEPLOY.md)
4. 提交 Issue 到 GitHub
5. 加入社区讨论

## 更新日志

### 版本 2.0.0 - SillyTavern 集成

**新增功能**：
- ✨ SillyTavern 角色卡、世界书、预设和正则处理
- ✨ Web 管理器界面
- ✨ 智能上下文管理和自动摘要
- ✨ Telegraph 对话分享
- ✨ 新的命令：/login、/share、/clear_all_chat

**改进**：
- 🔧 优化数据库模型
- 🔧 改进权限控制系统
- 🔧 增强错误处理和日志记录

**变更**：
- ⚠️ SYSTEM_INIT_MESSAGE 现在使用角色卡的 system_prompt
- ⚠️ /clear 命令现在创建截断标记而不是删除历史

**修复**：
- 🐛 修复多个并发请求的问题
- 🐛 改进流式响应的稳定性

## 总结

迁移到新版本应该是平滑的：

1. ✅ 备份数据
2. ✅ 更新代码
3. ✅ 添加新的环境变量
4. ✅ 更新 Docker 配置
5. ✅ 启动服务（自动迁移数据库）
6. ✅ 验证功能

如果遇到任何问题，请参考常见问题部分或寻求帮助。

祝使用愉快！🎉
