# 需求文档

## 简介

本文档定义了将 SillyTavern 核心功能集成到现有 Telegram bot 的需求。该集成将使用户能够通过基于 Web 的管理界面管理角色卡、预设、世界书和正则表达式模式，同时保持 SillyTavern 的请求构建逻辑和上下文管理能力。

## 术语表

- **Bot（机器人）**: Telegram bot 应用程序
- **Manager（管理器）**: 用于管理 SillyTavern 功能的 Web 管理界面
- **Character_Card（角色卡）**: SillyTavern 角色定义，包含个性和行为
- **Preset（预设）**: AI 模型参数的配置模板
- **World_Book（世界书）**: 基于触发器注入信息的上下文知识条目
- **Regex_Pattern（正则模式）**: 用于输入/输出处理的文本转换规则
- **Session（会话）**: 由 ChatID、BotID、UserID 和 ThreadID 标识的对话上下文
- **Summary（摘要）**: 压缩版本的对话历史，用于管理上下文长度
- **Telegraph_API**: Telegram 的内容发布服务，用于分享对话
- **Login_Token（登录令牌）**: 有时限的管理器访问认证令牌
- **Message_Role（消息角色）**: 消息的类型分类（user、assistant、system、summary）
- **Context_Window（上下文窗口）**: 可以发送给 AI 模型的最大 token 数量
- **Build_Request（构建请求）**: 从对话历史构建 API 请求的过程

## 需求

### 需求 1: SillyTavern 功能集成

**用户故事:** 作为机器人用户，我希望使用 SillyTavern 的角色卡、预设、世界书和正则模式，以便我可以与可自定义的 AI 个性进行丰富的、上下文感知的对话。

#### 验收标准

1. WHEN 用户通过管理器上传角色卡 THEN THE Bot SHALL 将其存储在数据库中并使其可供选择
2. WHEN 用户通过管理器上传预设 THEN THE Bot SHALL 将其存储在数据库中并应用于 AI 请求
3. WHEN 用户通过管理器上传世界书 THEN THE Bot SHALL 将条目存储在数据库中并根据触发条件注入它们
4. WHEN 用户通过管理器上传正则模式 THEN THE Bot SHALL 将其存储在数据库中并对消息应用转换
5. WHEN 构建 AI 请求时 THEN THE Bot SHALL 使用 SillyTavern 的逻辑组合角色卡、预设、世界书和正则模式
6. WHEN 渲染响应时 THEN THE Bot SHALL 通过 Telegram 输出而不是 Web UI

### 需求 2: 上下文长度管理

**用户故事:** 作为机器人管理员，我希望配置最大上下文长度和自动摘要，以便对话可以无限期地继续而不超过模型限制。

#### 验收标准

1. WHEN 上下文长度超过配置的阈值 THEN THE Bot SHALL 自动摘要较旧的消息
2. WHEN 摘要后构建请求 THEN THE Bot SHALL 包含系统消息、摘要和最近的消息对
3. THE Bot SHALL 支持可配置的 MAX_CONTEXT_LENGTH 环境变量
4. THE Bot SHALL 支持可配置的 SUMMARY_THRESHOLD 环境变量
5. WHEN 发生摘要时 THEN THE Bot SHALL 保留消息历史但将较旧的消息标记为已摘要

### 需求 3: 请求构建逻辑

**用户故事:** 作为开发者，我希望使用 SillyTavern 的请求构建逻辑，以便角色卡、世界书和预设能够正确组合到 AI 请求中。

#### 验收标准

1. WHEN 构建请求时 THEN THE Bot SHALL 使用 SillyTavern 的逻辑组合上下文元素
2. WHEN 构建请求时 THEN THE Bot SHALL 强制执行严格的角色交替（user/assistant）
3. WHEN 构建请求时 THEN THE Bot SHALL 使用 OpenAI 请求格式作为中间表示
4. WHEN 发送到不同提供商时 THEN THE Bot SHALL 将 OpenAI 格式转换为提供商特定格式
5. WHEN 构建请求时 THEN THE Bot SHALL 记录构建的请求以便调试
6. THE Bot SHALL 移除 SYSTEM_INIT_MESSAGE 配置，改用角色卡系统提示
7. THE Bot SHALL 将模型、温度和 token 设置存储为全局变量

### 需求 4: 管理器界面

**用户故事:** 作为机器人用户，我希望有一个基于 Web 的管理器界面，以便我可以上传和管理角色卡、预设、世界书和正则模式。

#### 验收标准

1. WHEN 访问管理器时 THEN THE Manager SHALL 要求通过 Telegram 用户 ID 和登录令牌进行身份验证
2. WHEN 用户在私聊中发送 /login THEN THE Bot SHALL 生成有效期为 24 小时的登录令牌
3. WHEN 用户发送 /login THEN THE Bot SHALL 返回格式化消息："用户名：{tg_userid}\n密码：{token}\n有效期：24小时"
4. WHEN 登录令牌过期时 THEN THE Bot SHALL 自动从数据库中删除它
5. WHEN 用户在群组中发送 /login THEN THE Bot SHALL 拒绝该命令并提示仅支持私聊
6. WHEN 上传文件时 THEN THE Manager SHALL 支持与 SillyTavern 相同的格式
7. WHEN 管理世界书条目时 THEN THE Manager SHALL 支持启用、禁用和编辑条目
8. WHEN 管理正则模式时 THEN THE Manager SHALL 支持启用、禁用和编辑模式
9. THE Manager SHALL 提供与 SillyTavern 管理界面等效的功能

### 需求 5: 基于权限的配置

**用户故事:** 作为机器人管理员，我希望控制谁可以修改配置，以便我可以在用户之间保持一致的设置或允许个性化。

#### 验收标准

1. WHEN ENABLE_USER_SETTING 为 true THEN THE Manager SHALL 允许管理员修改全局或个人设置
2. WHEN ENABLE_USER_SETTING 为 true THEN THE Manager SHALL 允许普通用户修改自己的设置
3. WHEN ENABLE_USER_SETTING 为 false THEN THE Manager SHALL 阻止普通用户修改任何设置
4. WHEN ENABLE_USER_SETTING 为 false THEN THE Manager SHALL 允许管理员修改全局设置
5. WHEN ENABLE_USER_SETTING 为 false THEN THE Bot SHALL 对所有用户使用全局设置

### 需求 6: 消息类型分类

**用户故事:** 作为开发者，我希望按角色类型对消息进行分类，以便系统可以正确管理对话历史和摘要。

#### 验收标准

1. WHEN 存储消息时 THEN THE Bot SHALL 将它们分类为 user、assistant、system 或 summary
2. WHEN 触发摘要时 THEN THE Bot SHALL 保留系统消息、摘要和最近的两对 user/assistant
3. WHEN 构建请求时 THEN THE Bot SHALL 仅包含 system + summary + 最近的消息对
4. WHEN 执行 /clear 时 THEN THE Bot SHALL 创建截断标记而不删除历史
5. WHEN /clear 后加载历史时 THEN THE Bot SHALL 仅加载截断标记之后的消息

### 需求 7: 管理员历史管理

**用户故事:** 作为机器人管理员，我希望清除所有对话历史，以便我可以执行维护或重置系统。

#### 验收标准

1. WHEN 管理员执行 /clear_all_chat THEN THE Bot SHALL 从数据库中删除所有对话历史
2. WHEN 非管理员尝试 /clear_all_chat THEN THE Bot SHALL 拒绝该命令
3. THE Bot SHALL 记录所有 /clear_all_chat 操作以供审计

### 需求 8: 对话分享

**用户故事:** 作为机器人用户，我希望分享我的对话，以便我可以发布有趣的交流或获得问题帮助。

#### 验收标准

1. WHEN 用户执行 /share THEN THE Bot SHALL 从当前会话收集所有 user 和 assistant 消息
2. WHEN 收集消息以供分享时 THEN THE Bot SHALL 包含消息而不考虑摘要状态
3. WHEN 收集消息时 THEN THE Bot SHALL 为 Telegraph API 格式化它们
4. WHEN 格式化完成时 THEN THE Bot SHALL 发布到 Telegraph API 并接收 URL
5. WHEN 收到 URL 时 THEN THE Bot SHALL 将其发送到聊天

### 需求 9: 严格角色交替

**用户故事:** 作为开发者，我希望在请求中强制执行严格的角色交替，以便所有 AI 提供商都能收到格式正确的对话历史。

#### 验收标准

1. WHEN 构建请求时 THEN THE Bot SHALL 确保消息在 user 和 assistant 角色之间交替
2. WHEN 连续消息具有相同角色时 THEN THE Bot SHALL 将它们合并为单个消息
3. WHEN 第一条消息不是来自 user THEN THE Bot SHALL 在前面添加一个空的 user 消息
4. WHEN 最后一条消息不是来自 user THEN THE Bot SHALL 附加当前 user 消息
5. WHEN 记录请求时 THEN THE Bot SHALL 打印最终消息序列以便调试
