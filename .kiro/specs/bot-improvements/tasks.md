# 实现计划: Bot Improvements

## 概述

本实现计划将机器人改进分解为离散的编码步骤，包括包名更新、数据库重构、对话归档和调试功能。每个任务都引用具体的需求以确保可追溯性。

## 任务

- [x] 1. 更新包名和导入路径
  - 更新 go.mod 中的模块路径为 "github.com/Vickie63878533/gpt-chatbot-telegram"
  - 使用 sed 或 Go 工具批量更新所有 .go 文件中的 import 语句
  - 运行 `go mod tidy` 验证依赖
  - 运行 `go build ./...` 验证编译成功
  - _需求: 1.1, 1.2, 1.3_

- [ ] 2. 创建新的 Message 数据模型
  - [ ] 2.1 在 internal/storage/models.go 中定义 Message 结构体
    - 添加字段: ID, CreatedAt, UpdatedAt, DeletedAt, ChatID, BotID, UserID, ThreadID
    - 添加字段: Role, Content, Timestamp, Truncated, Hidden
    - 添加 GORM 标签和索引定义
    - _需求: 2.1, 2.3_

  - [ ]* 2.2 为 Message 模型编写属性测试
    - **属性 1: 消息持久化完整性**
    - **验证: 需求 2.1, 2.3**

- [ ] 3. 实现存储接口的消息操作
  - [ ] 3.1 在 internal/storage/types.go 中添加新方法签名
    - SaveMessage(ctx *SessionContext, msg *Message) error
    - GetMessages(ctx *SessionContext, includeHidden bool) ([]*Message, error)
    - GetMessagesAfterTruncation(ctx *SessionContext, includeHidden bool) ([]*Message, error)
    - HideMessage(id uint) error
    - DeleteMessages(ctx *SessionContext) error
    - _需求: 2.1, 2.2_

  - [ ] 3.2 在 internal/storage/gorm.go 中实现消息操作
    - 实现 SaveMessage 方法
    - 实现 GetMessages 方法（支持 includeHidden 参数）
    - 实现 GetMessagesAfterTruncation 方法
    - 实现 HideMessage 方法
    - 实现 DeleteMessages 方法
    - _需求: 2.1, 2.2, 2.7_

  - [ ]* 3.3 为消息操作编写属性测试
    - **属性 2: 消息时间顺序性**
    - **验证: 需求 2.2**
    - **属性 3: 隐藏消息过滤**
    - **验证: 需求 2.7, 5.7**

- [ ] 4. 实现数据库迁移功能
  - [ ] 4.1 在 internal/storage/gorm.go 中实现 MigrateHistoryToMessages 方法
    - 读取所有 ChatHistory 记录
    - 解析 JSON 历史为 HistoryItem 数组
    - 为每个 HistoryItem 创建 Message 记录
    - 使用事务确保原子性
    - _需求: 2.4_

  - [ ]* 4.2 编写迁移测试
    - 创建旧格式测试数据
    - 运行迁移
    - 验证数据完整性和数量
    - _需求: 2.4_

- [ ] 5. 检查点 - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户。


- [ ] 6. 更新上下文管理器以使用新的消息模型
  - [ ] 6.1 更新 AddMessage 方法
    - 修改为使用 storage.SaveMessage 而不是操作 JSON 数组
    - 序列化 content 为 JSON 字符串
    - 调用 checkContextLength 检查上下文限制
    - _需求: 2.1, 5.1_

  - [ ] 6.2 实现 checkContextLength 方法
    - 获取非隐藏消息并估算令牌数
    - 如果超过 MAX_CONTEXT_LENGTH，删除最早的非系统/非摘要消息
    - 如果达到 SUMMARY_THRESHOLD，触发摘要
    - _需求: 5.2, 5.3_

  - [ ]* 6.3 为上下文长度管理编写属性测试
    - **属性 8: 上下文长度自动管理**
    - **验证: 需求 5.2**
    - **属性 9: 摘要阈值触发**
    - **验证: 需求 5.3**

  - [ ] 6.4 更新 TriggerSummary 方法
    - 使用 GetMessagesAfterTruncation 获取消息
    - 生成摘要后保存为新的 Message 记录
    - 调用 HideMessage 隐藏已摘要的消息
    - 保留最近 MIN_RECENT_PAIRS 对消息
    - _需求: 2.6, 5.4, 5.6_

  - [ ]* 6.5 为摘要功能编写属性测试
    - **属性 4: 摘要触发隐藏标记**
    - **验证: 需求 2.6, 5.6**
    - **属性 10: 摘要后消息保留**
    - **验证: 需求 5.4**

  - [ ] 6.6 更新 GetBuildHistory 方法
    - 使用 GetMessagesAfterTruncation 获取非隐藏消息
    - 转换 Message 为 HistoryItem 格式
    - 确保包含系统提示
    - _需求: 2.7, 5.5, 5.7_

  - [ ]* 6.7 为构建历史编写属性测试
    - **属性 11: 构建上下文包含系统提示**
    - **验证: 需求 5.5**

- [ ] 7. 添加 DEBUG_CHAT 配置
  - [ ] 7.1 在 internal/config/config.go 中添加 DebugChat 字段
    - 添加 `DebugChat bool` 字段，默认值 false
    - 在 LoadConfig 中读取 DEBUG_CHAT 环境变量
    - 更新 Validate 方法（如需要）
    - _需求: 4.1, 4.2_

  - [ ]* 7.2 编写配置测试
    - 测试 DEBUG_CHAT=true 启用调试模式
    - 测试 DEBUG_CHAT=false 或未设置禁用调试模式
    - _需求: 4.1, 4.2, 4.7_

- [ ] 8. 更新 Clear 命令实现
  - [ ] 8.1 重构 Clear 命令以生成 Telegraph 归档
    - 使用 storage.GetMessages 获取完整历史（includeHidden=true）
    - 过滤用户和助手消息
    - 调用 telegraph.CreatePage 生成归档
    - 向用户发送归档 URL
    - _需求: 3.1, 3.2, 3.4_

  - [ ] 8.2 实现管理员通知功能
    - 检查 config.DebugChat 标志
    - 如果启用，遍历 config.ChatAdminKey
    - 向每个管理员发送包含用户信息和归档 URL 的消息
    - 格式化消息包含 ChatID、用户名和 URL
    - _需求: 4.3, 4.4, 4.6_

  - [ ] 8.3 删除数据库记录
    - 调用 storage.DeleteMessages 删除会话的所有消息
    - 确保在发送归档后执行
    - _需求: 3.3_

  - [ ]* 8.4 为 Clear 命令编写属性测试
    - **属性 5: Clear 命令完整流程**
    - **验证: 需求 3.1, 3.2, 3.3, 3.4**
    - **属性 6: 调试模式管理员通知**
    - **验证: 需求 4.3, 4.4, 4.6**
    - **属性 7: 调试模式禁用时无通知**
    - **验证: 需求 4.5**

- [ ] 9. 检查点 - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [ ] 10. 统一服务器端口
  - [ ] 10.1 更新 internal/server/server.go 路由配置
    - 移除单独的管理器服务器启动逻辑
    - 在主服务器上添加 `/api/manager/*` 路由组
    - 将管理器处理器挂载到路由组
    - 保留其他路由在根路径
    - _需求: 6.1, 6.2, 6.3_

  - [ ] 10.2 更新配置和文档
    - 在 config.go 中标记 MANAGER_PORT 为已弃用（保留向后兼容）
    - 更新 README.md 移除 MANAGER_PORT 引用
    - 更新 doc/CONFIG.md 说明端口统一
    - _需求: 6.4, 6.5_

  - [ ]* 10.3 为路由编写属性测试
    - **属性 12: 路由路径正确性**
    - **验证: 需求 6.2, 6.3**

- [ ] 11. 更新默认配置值
  - [ ] 11.1 在 internal/config/config.go 中更新默认值
    - 将 MAX_CONTEXT_LENGTH 默认值改为 64000
    - 确认 SUMMARY_THRESHOLD 默认值为 0.8
    - 确认 MIN_RECENT_PAIRS 默认值为 2
    - _需求: 5.1, 5.2, 5.3, 5.4_

  - [ ]* 11.2 编写配置默认值测试
    - 测试未设置环境变量时的默认值
    - 验证 MAX_CONTEXT_LENGTH = 64000
    - 验证 SUMMARY_THRESHOLD = 0.8
    - 验证 MIN_RECENT_PAIRS = 2

- [ ] 12. 创建迁移命令或脚本
  - [ ] 12.1 创建 cmd/migrate/main.go
    - 加载配置
    - 连接数据库
    - 调用 MigrateHistoryToMessages
    - 输出迁移统计信息
    - _需求: 2.4_

  - [ ]* 12.2 编写端到端迁移测试
    - 创建包含多个会话的旧格式数据
    - 运行迁移命令
    - 验证所有消息正确迁移
    - 验证会话隔离
    - _需求: 2.4_

- [ ] 13. 集成测试和验证
  - [ ] 13.1 编写端到端对话流程测试
    - 创建新对话
    - 添加多条消息
    - 触发摘要
    - 验证隐藏消息
    - 执行 clear 命令
    - 验证 Telegraph 归档
    - 验证数据库清理

  - [ ]* 13.2 编写并发测试
    - 多个会话同时添加消息
    - 验证数据隔离
    - 验证无竞态条件

  - [ ]* 13.3 编写性能测试
    - 测试 1000 条消息查询性能
    - 测试 10000 条消息查询性能
    - 验证索引效果

- [ ] 14. 最终检查点 - 确保所有测试通过
  - 运行完整测试套件
  - 验证所有属性测试通过
  - 验证所有单元测试通过
  - 验证所有集成测试通过
  - 如有问题请询问用户。

## 注意事项

- 标记 `*` 的任务是可选的，可以跳过以加快 MVP 开发
- 每个任务都引用具体需求以确保可追溯性
- 检查点确保增量验证
- 属性测试验证通用正确性属性
- 单元测试验证具体示例和边缘情况
