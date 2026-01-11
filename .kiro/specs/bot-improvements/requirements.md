# 需求文档

## 简介

本规范定义了 Telegram 聊天机器人系统的改进，重点关注包名更新、数据库优化、对话归档和调试功能。这些改进旨在使系统更易于维护、更高效和更用户友好。

## 术语表

- **System（系统）**: Telegram 聊天机器人应用程序
- **Package_Name（包名）**: go.mod 中的 Go 模块标识符
- **Chat_History（聊天历史）**: 存储在数据库中的对话消息
- **Message_Record（消息记录）**: 聊天历史中的单条消息条目
- **Telegraph（Telegraph服务）**: 用于发布对话归档的外部服务
- **Clear_Command（清除命令）**: 截断对话上下文的 /clear 命令
- **Admin_User（管理员用户）**: 具有管理权限的用户（在 CHAT_ADMIN_KEY 环境变量中定义）
- **Debug_Mode（调试模式）**: 用于监控对话归档的特殊模式（通过 DEBUG_CHAT 环境变量控制，默认为 false）
- **Session_Context（会话上下文）**: 对话的唯一标识符（ChatID、BotID、UserID、ThreadID）
- **MAX_CONTEXT_LENGTH（最大上下文长度）**: 上下文窗口的最大令牌数（默认 64000）
- **SUMMARY_THRESHOLD（摘要阈值）**: 触发摘要的上下文使用率阈值（默认 0.8）
- **MIN_RECENT_PAIRS（最小保留对数）**: 摘要后必须保留的最近消息对数量（默认 2）
- **Hidden_Message（隐藏消息）**: 已被摘要但仍保留在数据库中的消息，不包含在 AI 请求中

## 需求

### 需求 1: 更新包名

**用户故事:** 作为开发者，我希望包名反映正确的仓库 URL，以便代码库能够被正确识别和维护。

#### 验收标准

1. WHEN 读取 go.mod 文件 THEN THE System SHALL 显示模块路径为 "github.com/Vickie63878533/gpt-chatbot-telegram"
2. WHEN Go 工具解析导入 THEN THE System SHALL 使用新的包路径
3. WHEN 构建包 THEN THE System SHALL 使用新包名成功编译

### 需求 2: 优化聊天历史存储

**用户故事:** 作为系统管理员，我希望聊天历史以单条消息记录的形式存储，以便优化数据库和内存使用。

#### 验收标准

1. WHEN 保存新消息 THEN THE System SHALL 在数据库中创建单条 Message_Record
2. WHEN 检索聊天历史 THEN THE System SHALL 按时间戳顺序查询单条 Message_Records
3. WHEN 存储消息 THEN THE System SHALL 包含角色、内容、时间戳、Session_Context 和隐藏标志
4. WHEN 为历史分配内存 THEN THE System SHALL 使用单个消息对象而不是序列化的 JSON 字符串
5. WHEN 消息被摘要 THEN THE System SHALL 设置隐藏标志为 true
6. WHEN 检索用于 AI 请求的历史 THEN THE System SHALL 过滤掉隐藏的消息

### 需求 3: 清除时归档对话

**用户故事:** 作为用户，我希望在清除对话时收到 Telegraph 链接，以便在不占用数据库存储的情况下访问归档历史。

#### 验收标准

1. WHEN 用户执行 /clear THEN THE System SHALL 生成对话的 Telegraph 归档
2. WHEN 创建 Telegraph 归档 THEN THE System SHALL 向用户发送归档 URL
3. WHEN 发送 Telegraph 归档 THEN THE System SHALL 删除该 Session_Context 的所有 Message_Records
4. WHEN 归档对话 THEN THE System SHALL 包含所有用户和助手消息
5. WHEN 删除数据库记录 THEN THE System SHALL 维护数据库完整性

### 需求 4: 对话监控的调试模式

**用户故事:** 作为管理员，我希望在调试模式下用户清除对话时收到 Telegraph 归档，以便监控和排查用户交互问题。

#### 验收标准

1. WHEN DEBUG_CHAT 环境变量设置为 true THEN THE System SHALL 启用调试模式
2. WHEN DEBUG_CHAT 环境变量未设置或设置为 false THEN THE System SHALL 禁用调试模式（默认行为）
3. WHEN 用户在调试模式下执行 /clear THEN THE System SHALL 向所有 Admin_Users 发送 Telegraph 归档
4. WHEN 发送给管理员 THEN THE System SHALL 在消息中包含用户的 ChatID 和用户名
5. WHEN 调试模式禁用 THEN THE System SHALL NOT 向 Admin_Users 发送归档
6. WHEN 管理员收到调试归档 THEN THE System SHALL 使用清晰的源用户标识格式化消息
7. WHEN 系统启动 THEN THE System SHALL 从 CHAT_ADMIN_KEY 环境变量读取管理员用户列表

### 需求 5: 上下文长度边界处理

**用户故事:** 作为系统，我需要正确处理上下文长度边界情况，以便在消息对超过限制时保持系统稳定。

#### 验收标准

1. WHEN 当前上下文使用率低于 SUMMARY_THRESHOLD THEN THE System SHALL 正常添加新消息对
2. WHEN 添加新消息对会导致总长度超过 MAX_CONTEXT_LENGTH THEN THE System SHALL 删除最早的消息对直到有足够空间
3. WHEN 当前上下文使用率达到或超过 SUMMARY_THRESHOLD THEN THE System SHALL 触发摘要生成
4. WHEN 生成摘要后 THEN THE System SHALL 保留摘要和最近 MIN_RECENT_PAIRS 对消息
5. WHEN 构建请求上下文 THEN THE System SHALL 包含 SYSTEM_PROMPT 和其他默认信息
6. WHEN 历史消息被摘要 THEN THE System SHALL 标记为隐藏但不删除
7. WHEN 检索用于 AI 请求的历史 THEN THE System SHALL 仅包含未隐藏的消息和摘要

### 需求 6: 统一服务端口

**用户故事:** 作为部署人员，我希望所有服务使用同一端口，以便简化网络配置和防火墙规则。

#### 验收标准

1. WHEN 系统启动 THEN THE System SHALL 在单一 PORT 上启动 HTTP 服务器
2. WHEN 访问 /api/manager/* 路径 THEN THE System SHALL 路由到管理器 API 处理器
3. WHEN 访问其他路径 THEN THE System SHALL 路由到主应用处理器
4. WHEN MANAGER_PORT 环境变量存在 THEN THE System SHALL 忽略它并使用 PORT
5. WHEN 配置文档更新 THEN THE System SHALL 移除 MANAGER_PORT 的引用

## 非功能性需求

### 性能

1. 单条记录的消息存储操作 SHALL 在 100ms 内完成
2. 历史检索 SHALL 支持分页以处理大型对话历史
3. 对于最多 1000 条消息的对话，Telegraph 归档生成 SHALL 在 5 秒内完成

### 可扩展性

1. 数据库架构 SHALL 支持数百万条单独的消息记录
2. THE System SHALL 在 Session_Context 字段上使用数据库索引以实现高效查询
3. 内存使用 SHALL 与活动对话数量线性扩展，而不是与消息数量扩展

### 安全性

1. Telegraph 归档 SHALL 仅通过生成的 URL 访问
2. Admin_Users SHALL 在接收调试归档之前进行身份验证
3. 已删除的对话数据 SHALL 从数据库中永久删除

### 可维护性

1. 新包名 SHALL 在所有导入语句中一致使用
2. 数据库迁移 SHALL 可逆以支持回滚场景
3. 调试模式配置 SHALL 在环境变量中清晰记录
