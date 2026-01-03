# 配置文档

本文档描述了 Telegram Bot Go 版本的所有环境变量配置选项。

## 必需配置

### TELEGRAM_AVAILABLE_TOKENS
- **类型**: 字符串数组（逗号分隔）
- **必需**: 是
- **描述**: Telegram Bot Token 列表
- **示例**: `"123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"`

## AI 提供商配置

### 通用配置

#### AI_PROVIDER
- **类型**: 字符串
- **默认值**: `auto`
- **描述**: AI 提供商选择（auto、openai、azure、gemini、anthropic、workersai、mistral、cohere、deepseek、groq、xai）

#### AI_IMAGE_PROVIDER
- **类型**: 字符串
- **默认值**: `auto`
- **描述**: 图片生成提供商选择

#### SYSTEM_INIT_MESSAGE
- **类型**: 字符串
- **描述**: 系统初始化消息

### OpenAI 配置

#### OPENAI_API_KEY
- **类型**: 字符串数组（逗号分隔）
- **描述**: OpenAI API Key

#### OPENAI_CHAT_MODEL
- **类型**: 字符串
- **默认值**: `gpt-4o-mini`
- **描述**: OpenAI 聊天模型

#### OPENAI_API_BASE
- **类型**: 字符串
- **默认值**: `https://api.openai.com/v1`
- **描述**: OpenAI API 基础 URL

### Azure OpenAI 配置

#### AZURE_API_KEY
- **类型**: 字符串
- **描述**: Azure API Key

#### AZURE_RESOURCE_NAME
- **类型**: 字符串
- **描述**: Azure 资源名称

#### AZURE_CHAT_MODEL
- **类型**: 字符串
- **默认值**: `gpt-4o-mini`
- **描述**: Azure 聊天模型

### Google Gemini 配置

#### GOOGLE_API_KEY
- **类型**: 字符串
- **描述**: Google API Key

#### GOOGLE_CHAT_MODEL
- **类型**: 字符串
- **默认值**: `gemini-1.5-flash`
- **描述**: Gemini 聊天模型

### Anthropic 配置

#### ANTHROPIC_API_KEY
- **类型**: 字符串
- **描述**: Anthropic API Key

#### ANTHROPIC_CHAT_MODEL
- **类型**: 字符串
- **默认值**: `claude-3-5-haiku-latest`
- **描述**: Claude 聊天模型

## 权限配置

### I_AM_A_GENEROUS_PERSON
- **类型**: 布尔值
- **默认值**: `false`
- **描述**: 开放模式，允许所有用户使用

### CHAT_WHITE_LIST
- **类型**: 字符串数组（逗号分隔）
- **描述**: 私聊白名单（用户 ID）

### CHAT_GROUP_WHITE_LIST
- **类型**: 字符串数组（逗号分隔）
- **描述**: 群组白名单（群组 ID）

## 用户设置权限控制

### ENABLE_USER_SETTING
- **类型**: 布尔值
- **默认值**: `true`
- **描述**: 是否允许用户修改自己的配置（模型、温度、system prompt 等）
- **行为**:
  - 当设置为 `true` 时：所有用户都可以修改自己的配置，配置命令对所有用户可见
  - 当设置为 `false` 时：仅管理员可以修改配置，配置命令对普通用户隐藏，所有用户使用全局配置
- **向后兼容**: 未设置时默认为 `true`，保持现有行为
- **示例**:
  ```
  # 允许用户修改配置（默认）
  ENABLE_USER_SETTING=true

  # 仅允许管理员修改配置
  ENABLE_USER_SETTING=false
  ```

### CHAT_ADMIN_KEY
- **类型**: 字符串数组（逗号分隔 Telegram User ID）
- **默认值**: 空
- **描述**: 管理员用户 ID 列表，这些用户可以始终修改配置，不受 ENABLE_USER_SETTING 限制
- **优先级**: 
  1. 首先检查用户 ID 是否在 CHAT_ADMIN_KEY 列表中
  2. 如果不在列表中，检查用户是否是 Telegram 群组管理员
- **示例**:
  ```
  # 单个管理员
  CHAT_ADMIN_KEY=123456789

  # 多个管理员
  CHAT_ADMIN_KEY=123456789,987654321,555555555
  ```
- **获取 Telegram User ID**: 
  - 在 Telegram 中发送 `/start` 给 @userinfobot
  - 或在群组中使用 `/whoami` 命令（如果 bot 支持）

## 历史记录配置

### MAX_HISTORY_LENGTH
- **类型**: 整数
- **默认值**: `20`
- **描述**: 最大历史记录长度

### MAX_TOKEN_LENGTH
- **类型**: 整数
- **默认值**: `-1`（不限制）
- **描述**: 最大 Token 长度

### AUTO_TRIM_HISTORY
- **类型**: 布尔值
- **默认值**: `true`
- **描述**: 自动裁剪历史记录

## 服务器配置

### PORT
- **类型**: 整数
- **默认值**: `8080`
- **描述**: HTTP 服务器监听端口

### DB_PATH
- **类型**: 字符串
- **默认值**: `./data/bot.db`
- **描述**: SQLite 数据库文件路径（当 DSN 未设置时使用）

## 数据库配置

### DSN
- **类型**: 字符串
- **默认值**: 空（使用 SQLite）
- **描述**: 数据库连接字符串（Data Source Name）
- **优先级**: 当同时设置 DSN 和 DB_PATH 时，DSN 优先
- **支持的格式**:
  - **SQLite**: `sqlite:///path/to/db.db` 或留空使用 DB_PATH
  - **MySQL**: `mysql://user:password@host:port/dbname`
  - **PostgreSQL**: `postgres://user:password@host:port/dbname` 或 `postgresql://user:password@host:port/dbname`
- **示例**:
  ```
  # SQLite (默认)
  DSN=

  # MySQL
  DSN=mysql://root:password@localhost:3306/telegram_bot

  # PostgreSQL
  DSN=postgres://user:password@localhost:5432/telegram_bot
  ```

## 特性开关

### STREAM_MODE
- **类型**: 布尔值
- **默认值**: `true`
- **描述**: 启用流式输出

### SAFE_MODE
- **类型**: 布尔值
- **默认值**: `true`
- **描述**: 启用安全模式（防止重复消息）

### DEBUG_MODE
- **类型**: 布尔值
- **默认值**: `false`
- **描述**: 启用调试模式

### DEV_MODE
- **类型**: 布尔值
- **默认值**: `false`
- **描述**: 启用开发模式

## 语言配置

### LANGUAGE
- **类型**: 字符串
- **默认值**: `zh-cn`
- **可选值**: `zh-cn`, `en`, `pt`, `zh-hant`
- **描述**: 界面语言

## 完整配置示例

### 基础配置（SQLite）

```bash
# 必需配置
TELEGRAM_AVAILABLE_TOKENS="123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# AI 配置
AI_PROVIDER="openai"
OPENAI_CHAT_MODEL="gpt-4o-mini"

# 权限配置
CHAT_WHITE_LIST="123456789,987654321"

# 用户设置权限控制
ENABLE_USER_SETTING=true
CHAT_ADMIN_KEY="123456789"

# 服务器配置
PORT="8080"
DB_PATH="./data/bot.db"

# 特性开关
STREAM_MODE="true"
SAFE_MODE="true"

# 语言配置
LANGUAGE="zh-cn"
```

### MySQL 数据库配置

```bash
# 必需配置
TELEGRAM_AVAILABLE_TOKENS="123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# 数据库配置（MySQL）
DSN="mysql://root:password@localhost:3306/telegram_bot"

# 用户设置权限控制
ENABLE_USER_SETTING=false
CHAT_ADMIN_KEY="123456789,987654321"

# 其他配置...
```

### PostgreSQL 数据库配置

```bash
# 必需配置
TELEGRAM_AVAILABLE_TOKENS="123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# 数据库配置（PostgreSQL）
DSN="postgres://user:password@localhost:5432/telegram_bot"

# 用户设置权限控制
ENABLE_USER_SETTING=false
CHAT_ADMIN_KEY="123456789"

# 其他配置...
```
