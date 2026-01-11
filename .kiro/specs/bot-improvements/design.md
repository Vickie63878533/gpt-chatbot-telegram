# 设计文档

## 概述

本设计文档描述了 Telegram 聊天机器人系统的四个主要改进：
1. 更新包名以反映新的 GitHub 仓库
2. 重构聊天历史存储为单条消息记录模式
3. 在 /clear 命令时自动归档对话到 Telegraph 并删除数据库记录
4. 添加调试模式以便管理员监控用户对话

这些改进将提高系统的可维护性、性能和可观察性。

## 架构

### 当前架构

当前系统使用以下架构：
- **存储层**: GORM ORM，支持 SQLite、MySQL、PostgreSQL
- **聊天历史**: 存储为 JSON 序列化的消息数组在单个 `ChatHistory` 记录中
- **上下文管理**: `ContextManager` 处理摘要、截断和上下文窗口
- **Telegraph 集成**: 用于分享对话的外部服务

### 新架构

改进后的架构将包括：
- **消息表**: 新的 `Message` 表存储单条消息记录
- **隐藏标志**: 消息支持 `Hidden` 标志用于摘要后的消息
- **调试模式**: 新的 `DEBUG_CHAT` 环境变量控制管理员通知
- **统一端口**: 所有服务在单一端口上运行，通过路径区分

## 组件和接口

### 1. 包名更新

#### 影响的文件
- `go.mod`: 模块路径
- 所有 `.go` 文件: import 语句

#### 更新策略
使用 Go 的模块替换功能进行批量更新：
```bash
# 更新 go.mod
go mod edit -module github.com/Vickie63878533/gpt-chatbot-telegram

# 更新所有 import 语句
find . -name "*.go" -type f -exec sed -i 's|github.com/tbxark/ChatGPT-Telegram-Workers/go_version|github.com/Vickie63878533/gpt-chatbot-telegram|g' {} +
```

### 2. 消息存储模型

#### 新数据模型

```go
// Message represents a single chat message
type Message struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // Session identifiers
    ChatID   int64  `gorm:"not null;index:idx_message_session,priority:1"`
    BotID    int64  `gorm:"not null;index:idx_message_session,priority:2"`
    UserID   *int64 `gorm:"index:idx_message_session,priority:3"`
    ThreadID *int64 `gorm:"index:idx_message_session,priority:4"`
    
    // Message data
    Role      string `gorm:"not null;index"`
    Content   string `gorm:"type:text;not null"`
    Timestamp int64  `gorm:"not null;index"`
    
    // Flags
    Truncated bool `gorm:"default:false;index"`
    Hidden    bool `gorm:"default:false;index"`
}
```

#### 索引策略
- 复合索引 `idx_message_session`: (ChatID, BotID, UserID, ThreadID) 用于快速会话查询
- 单列索引 `Timestamp`: 用于时间排序
- 单列索引 `Hidden`: 用于过滤隐藏消息
- 单列索引 `Truncated`: 用于查找截断标记

### 3. 存储接口更新

#### 新方法

```go
// Storage interface additions
type Storage interface {
    // ... existing methods ...
    
    // Message operations
    SaveMessage(ctx *SessionContext, msg *Message) error
    GetMessages(ctx *SessionContext, includeHidden bool) ([]*Message, error)
    GetMessagesAfterTruncation(ctx *SessionContext, includeHidden bool) ([]*Message, error)
    HideMessagesBefore(ctx *SessionContext, beforeTimestamp int64) error
    DeleteMessages(ctx *SessionContext) error
    
    // Migration helper
    MigrateHistoryToMessages() error
}
```

#### 实现细节

**SaveMessage**: 插入单条消息记录
```go
func (s *GORMStorage) SaveMessage(ctx *SessionContext, msg *Message) error {
    msg.ChatID = ctx.ChatID
    msg.BotID = ctx.BotID
    msg.UserID = ctx.UserID
    msg.ThreadID = ctx.ThreadID
    
    result := s.db.Create(msg)
    return result.Error
}
```

**GetMessages**: 检索所有消息或仅非隐藏消息
```go
func (s *GORMStorage) GetMessages(ctx *SessionContext, includeHidden bool) ([]*Message, error) {
    query := s.buildSessionQuery(ctx)
    
    if !includeHidden {
        query = query.Where("hidden = ?", false)
    }
    
    var messages []*Message
    result := query.Order("timestamp ASC").Find(&messages)
    return messages, result.Error
}
```

**GetMessagesAfterTruncation**: 获取最后一个截断标记之后的消息
```go
func (s *GORMStorage) GetMessagesAfterTruncation(ctx *SessionContext, includeHidden bool) ([]*Message, error) {
    // Find last truncation marker
    var lastTruncation Message
    truncQuery := s.buildSessionQuery(ctx).
        Where("truncated = ?", true).
        Order("timestamp DESC").
        Limit(1)
    
    if err := truncQuery.First(&lastTruncation).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // No truncation, return all messages
            return s.GetMessages(ctx, includeHidden)
        }
        return nil, err
    }
    
    // Get messages after truncation
    query := s.buildSessionQuery(ctx).
        Where("timestamp > ?", lastTruncation.Timestamp)
    
    if !includeHidden {
        query = query.Where("hidden = ?", false)
    }
    
    var messages []*Message
    result := query.Order("timestamp ASC").Find(&messages)
    return messages, result.Error
}
```

### 4. 上下文管理器更新

#### 修改的方法

**AddMessage**: 添加消息并检查上下文限制
```go
func (m *ContextManager) AddMessage(ctx *SessionContext, role string, content interface{}) error {
    // 序列化 content
    contentJSON, err := json.Marshal(content)
    if err != nil {
        return err
    }
    
    // 创建消息
    msg := &storage.Message{
        Role:      role,
        Content:   string(contentJSON),
        Timestamp: time.Now().Unix(),
        Truncated: false,
        Hidden:    false,
    }
    
    // 保存消息
    if err := m.storage.SaveMessage(ctx, msg); err != nil {
        return err
    }
    
    // 检查上下文长度
    return m.checkContextLength(ctx)
}
```

**checkContextLength**: 新方法处理上下文边界
```go
func (m *ContextManager) checkContextLength(ctx *SessionContext) error {
    // 获取构建历史（非隐藏消息）
    messages, err := m.storage.GetMessagesAfterTruncation(ctx, false)
    if err != nil {
        return err
    }
    
    // 估算令牌数
    tokens := m.estimateTokens(messages)
    
    // 情况 1: 超过最大长度 - 删除最早的消息
    for tokens > m.config.MaxContextLength {
        // 找到最早的非系统、非摘要消息
        var oldestMsg *storage.Message
        for _, msg := range messages {
            if msg.Role != "system" && msg.Role != "summary" {
                oldestMsg = msg
                break
            }
        }
        
        if oldestMsg == nil {
            break // 没有可删除的消息
        }
        
        // 标记为隐藏
        if err := m.storage.HideMessage(oldestMsg.ID); err != nil {
            return err
        }
        
        // 重新计算
        messages, err = m.storage.GetMessagesAfterTruncation(ctx, false)
        if err != nil {
            return err
        }
        tokens = m.estimateTokens(messages)
    }
    
    // 情况 2: 达到摘要阈值 - 触发摘要
    threshold := int(float64(m.config.MaxContextLength) * m.config.SummaryThreshold)
    if tokens >= threshold {
        go m.TriggerSummary(ctx)
    }
    
    return nil
}
```

**TriggerSummary**: 更新以使用新的消息模型
```go
func (m *ContextManager) TriggerSummary(ctx *SessionContext) error {
    // 获取非隐藏消息
    messages, err := m.storage.GetMessagesAfterTruncation(ctx, false)
    if err != nil {
        return err
    }
    
    // 分离消息类型
    var systemMsgs, summaryMsgs, conversationMsgs []*storage.Message
    for _, msg := range messages {
        switch msg.Role {
        case "system":
            systemMsgs = append(systemMsgs, msg)
        case "summary":
            summaryMsgs = append(summaryMsgs, msg)
        case "user", "assistant":
            conversationMsgs = append(conversationMsgs, msg)
        }
    }
    
    // 计算要保留的最近消息对数
    recentPairsToKeep := m.config.MinRecentPairs * 2
    
    if len(conversationMsgs) <= recentPairsToKeep {
        return nil // 消息不足，无需摘要
    }
    
    // 分割为要摘要的和要保留的
    toSummarize := conversationMsgs[:len(conversationMsgs)-recentPairsToKeep]
    toKeep := conversationMsgs[len(conversationMsgs)-recentPairsToKeep:]
    
    // 生成摘要（调用 AI）
    summaryText, err := m.generateSummary(toSummarize)
    if err != nil {
        return err
    }
    
    // 创建摘要消息
    summaryMsg := &storage.Message{
        Role:      "summary",
        Content:   fmt.Sprintf("Previous conversation summary: %s", summaryText),
        Timestamp: time.Now().Unix(),
        Truncated: false,
        Hidden:    false,
    }
    
    if err := m.storage.SaveMessage(ctx, summaryMsg); err != nil {
        return err
    }
    
    // 隐藏已摘要的消息
    for _, msg := range toSummarize {
        if err := m.storage.HideMessage(msg.ID); err != nil {
            return err
        }
    }
    
    return nil
}
```

**GetBuildHistory**: 更新以使用新的消息模型
```go
func (m *ContextManager) GetBuildHistory(ctx *SessionContext) ([]HistoryItem, error) {
    // 获取截断后的非隐藏消息
    messages, err := m.storage.GetMessagesAfterTruncation(ctx, false)
    if err != nil {
        return nil, err
    }
    
    // 转换为 HistoryItem
    items := make([]HistoryItem, len(messages))
    for i, msg := range messages {
        var content interface{}
        if err := json.Unmarshal([]byte(msg.Content), &content); err != nil {
            content = msg.Content // 回退到字符串
        }
        
        items[i] = HistoryItem{
            Role:      msg.Role,
            Content:   content,
            Timestamp: msg.Timestamp,
            Truncated: msg.Truncated,
        }
    }
    
    return items, nil
}
```

### 5. Clear 命令更新

#### 新流程

1. 获取完整对话历史
2. 生成 Telegraph 归档
3. 发送归档链接给用户
4. 如果启用 DEBUG_CHAT，发送给所有管理员
5. 删除数据库中的所有消息

#### 实现

```go
func (c *ClearCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
    bot := ctx.Bot.(*tgbotapi.BotAPI)
    sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
    
    // 1. 获取完整历史
    messages, err := c.storage.GetMessages(sessionCtx, true)
    if err != nil {
        return err
    }
    
    // 2. 过滤用户和助手消息
    var conversationMsgs []*storage.Message
    for _, msg := range messages {
        if msg.Role == "user" || msg.Role == "assistant" {
            conversationMsgs = append(conversationMsgs, msg)
        }
    }
    
    // 3. 生成 Telegraph 归档
    telegraphURL := ""
    if len(conversationMsgs) > 0 {
        telegraphClient, err := telegraph.NewClient()
        if err != nil {
            return err
        }
        
        htmlContent := formatMessagesForTelegraph(conversationMsgs)
        title := fmt.Sprintf("Conversation - %s", time.Now().Format("2006-01-02 15:04"))
        
        telegraphURL, err = telegraphClient.CreatePage(title, htmlContent)
        if err != nil {
            return err
        }
    }
    
    // 4. 发送给用户
    if telegraphURL != "" {
        responseText := fmt.Sprintf("✅ 对话已清除并归档\n\n🔗 %s", telegraphURL)
        msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
        bot.Send(msg)
    } else {
        msg := tgbotapi.NewMessage(message.Chat.ID, "✅ 对话已清除（无历史记录）")
        bot.Send(msg)
    }
    
    // 5. 如果启用调试模式，发送给管理员
    if c.config.DebugChat && telegraphURL != "" {
        c.sendToAdmins(bot, message, telegraphURL)
    }
    
    // 6. 删除数据库记录
    if err := c.storage.DeleteMessages(sessionCtx); err != nil {
        return err
    }
    
    return nil
}

func (c *ClearCommand) sendToAdmins(bot *tgbotapi.BotAPI, originalMsg *tgbotapi.Message, url string) {
    username := originalMsg.From.UserName
    if username == "" {
        username = originalMsg.From.FirstName
    }
    
    adminText := fmt.Sprintf(
        "🔍 调试通知\n\n"+
        "用户: @%s (ID: %d)\n"+
        "聊天: %d\n"+
        "归档: %s",
        username,
        originalMsg.From.ID,
        originalMsg.Chat.ID,
        url,
    )
    
    for _, adminIDStr := range c.config.ChatAdminKey {
        adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
        if err != nil {
            continue
        }
        
        msg := tgbotapi.NewMessage(adminID, adminText)
        bot.Send(msg)
    }
}
```

### 6. 配置更新

#### 新环境变量

```go
type Config struct {
    // ... existing fields ...
    
    // Debug mode
    DebugChat bool `env:"DEBUG_CHAT" default:"false"`
    
    // Context management (updated defaults)
    MaxContextLength int     `env:"MAX_CONTEXT_LENGTH" default:"64000"`
    SummaryThreshold float64 `env:"SUMMARY_THRESHOLD" default:"0.8"`
    MinRecentPairs   int     `env:"MIN_RECENT_PAIRS" default:"2"`
}
```

### 7. 服务器端口统一

#### 当前状态
- 主服务器: PORT (默认 8080)
- 管理器: MANAGER_PORT (默认 8081)

#### 新设计
- 单一端口: PORT (默认 8080)
- 路由:
  - `/api/manager/*` → 管理器 API
  - 其他 → 主应用

#### 实现

```go
func (s *Server) setupRoutes() {
    // 管理器路由
    if s.config.ManagerEnabled {
        managerGroup := s.router.Group("/api/manager")
        managerGroup.Use(s.authMiddleware())
        
        // 角色卡
        managerGroup.GET("/characters", s.handleListCharacters)
        managerGroup.POST("/characters", s.handleCreateCharacter)
        // ... 其他管理器路由
    }
    
    // 主应用路由
    s.router.GET("/health", s.handleHealth)
    // ... 其他主应用路由
}
```

## 数据模型

### Message 表结构

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| id | uint | PRIMARY | 主键 |
| created_at | timestamp | YES | 创建时间 |
| updated_at | timestamp | NO | 更新时间 |
| deleted_at | timestamp | YES | 软删除时间 |
| chat_id | int64 | COMPOSITE | 聊天 ID |
| bot_id | int64 | COMPOSITE | 机器人 ID |
| user_id | int64 | COMPOSITE | 用户 ID（可空） |
| thread_id | int64 | COMPOSITE | 线程 ID（可空） |
| role | string | YES | 消息角色 |
| content | text | NO | 消息内容（JSON） |
| timestamp | int64 | YES | Unix 时间戳 |
| truncated | bool | YES | 是否为截断标记 |
| hidden | bool | YES | 是否隐藏 |

### 迁移策略

#### 步骤 1: 创建新表
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    chat_id INTEGER NOT NULL,
    bot_id INTEGER NOT NULL,
    user_id INTEGER,
    thread_id INTEGER,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    truncated BOOLEAN DEFAULT 0,
    hidden BOOLEAN DEFAULT 0
);

CREATE INDEX idx_message_session ON messages(chat_id, bot_id, user_id, thread_id);
CREATE INDEX idx_message_timestamp ON messages(timestamp);
CREATE INDEX idx_message_hidden ON messages(hidden);
CREATE INDEX idx_message_truncated ON messages(truncated);
CREATE INDEX idx_message_role ON messages(role);
```

#### 步骤 2: 迁移数据
```go
func (s *GORMStorage) MigrateHistoryToMessages() error {
    var histories []ChatHistory
    if err := s.db.Find(&histories).Error; err != nil {
        return err
    }
    
    for _, history := range histories {
        var items []HistoryItem
        if err := json.Unmarshal([]byte(history.History), &items); err != nil {
            continue // 跳过损坏的记录
        }
        
        sessionCtx := &SessionContext{
            ChatID:   history.ChatID,
            BotID:    history.BotID,
            UserID:   history.UserID,
            ThreadID: history.ThreadID,
        }
        
        for _, item := range items {
            contentJSON, _ := json.Marshal(item.Content)
            
            msg := &Message{
                Role:      item.Role,
                Content:   string(contentJSON),
                Timestamp: item.Timestamp,
                Truncated: item.Truncated,
                Hidden:    false,
            }
            
            if err := s.SaveMessage(sessionCtx, msg); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

#### 步骤 3: 验证和清理
- 验证迁移的消息数量
- 备份旧的 `chat_histories` 表
- 可选：删除旧表

## 错误处理

### 错误场景

1. **Telegraph API 失败**
   - 重试机制：最多 3 次
   - 回退：如果失败，仍然清除对话但通知用户归档失败

2. **数据库迁移失败**
   - 事务回滚
   - 保留旧数据
   - 记录详细错误日志

3. **上下文长度超限**
   - 自动删除最早的消息
   - 如果仍然超限，触发紧急摘要
   - 最坏情况：保留最近 MIN_RECENT_PAIRS 对消息

4. **管理员通知失败**
   - 不影响用户操作
   - 记录失败的管理员 ID
   - 继续处理其他管理员

### 错误日志

```go
type ErrorLog struct {
    Timestamp time.Time
    Operation string
    Error     error
    Context   map[string]interface{}
}

func logError(op string, err error, ctx map[string]interface{}) {
    log := ErrorLog{
        Timestamp: time.Now(),
        Operation: op,
        Error:     err,
        Context:   ctx,
    }
    // 写入日志文件或监控系统
}
```

## 测试策略

我们将使用双重测试方法：单元测试用于具体示例和边缘情况，属性测试用于验证通用正确性属性。两者互补，共同提供全面的测试覆盖。

### 单元测试

单元测试专注于：
- 具体示例以演示正确行为
- 组件之间的集成点
- 边缘情况和错误条件

**测试覆盖范围**:

1. **存储层测试**
   - 测试消息 CRUD 操作
   - 测试会话查询
   - 测试隐藏标志过滤
   - 测试截断标记查找
   - 测试数据库迁移功能

2. **上下文管理器测试**
   - 测试消息添加
   - 测试上下文长度检查
   - 测试摘要触发
   - 测试边界情况（空上下文、单消息等）

3. **Clear 命令测试**
   - 测试 Telegraph 归档生成
   - 测试管理员通知
   - 测试数据库清理
   - 测试空对话处理

4. **配置测试**
   - 测试环境变量加载
   - 测试默认值
   - 测试调试模式切换

### 属性测试

属性测试验证通用正确性属性在所有输入下都成立。

**属性测试库**: 使用 Go 的 `testing/quick` 包或 `gopter` 库

**测试配置**:
- 每个属性测试最少运行 100 次迭代
- 使用随机生成的测试数据
- 每个测试引用其设计文档中的属性编号

**属性测试实现示例**:

```go
// 属性 1: 消息持久化完整性
// Feature: bot-improvements, Property 1: 消息持久化完整性
func TestProperty_MessagePersistenceIntegrity(t *testing.T) {
    f := func(role string, content string, timestamp int64) bool {
        // 设置测试数据库
        db := setupTestDB(t)
        defer db.Close()
        
        ctx := &storage.SessionContext{
            ChatID: 12345,
            BotID:  67890,
        }
        
        // 保存消息
        msg := &storage.Message{
            Role:      role,
            Content:   content,
            Timestamp: timestamp,
        }
        
        if err := db.SaveMessage(ctx, msg); err != nil {
            return false
        }
        
        // 检索消息
        messages, err := db.GetMessages(ctx, true)
        if err != nil || len(messages) == 0 {
            return false
        }
        
        // 验证字段
        retrieved := messages[len(messages)-1]
        return retrieved.Role == role &&
               retrieved.Content == content &&
               retrieved.Timestamp == timestamp &&
               retrieved.ChatID == ctx.ChatID &&
               retrieved.BotID == ctx.BotID
    }
    
    if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}

// 属性 3: 隐藏消息过滤
// Feature: bot-improvements, Property 3: 隐藏消息过滤
func TestProperty_HiddenMessageFiltering(t *testing.T) {
    f := func(visibleCount, hiddenCount uint8) bool {
        if visibleCount == 0 && hiddenCount == 0 {
            return true // 跳过空情况
        }
        
        db := setupTestDB(t)
        defer db.Close()
        
        ctx := &storage.SessionContext{
            ChatID: 12345,
            BotID:  67890,
        }
        
        // 创建可见消息
        for i := 0; i < int(visibleCount); i++ {
            msg := &storage.Message{
                Role:      "user",
                Content:   fmt.Sprintf("visible-%d", i),
                Timestamp: int64(i),
                Hidden:    false,
            }
            db.SaveMessage(ctx, msg)
        }
        
        // 创建隐藏消息
        for i := 0; i < int(hiddenCount); i++ {
            msg := &storage.Message{
                Role:      "user",
                Content:   fmt.Sprintf("hidden-%d", i),
                Timestamp: int64(visibleCount + i),
                Hidden:    true,
            }
            db.SaveMessage(ctx, msg)
        }
        
        // 检索非隐藏消息
        messages, err := db.GetMessages(ctx, false)
        if err != nil {
            return false
        }
        
        // 验证没有隐藏消息
        for _, msg := range messages {
            if msg.Hidden {
                return false
            }
        }
        
        // 验证数量
        return len(messages) == int(visibleCount)
    }
    
    if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### 集成测试

1. **端到端对话流程**
   - 创建对话
   - 添加多条消息
   - 触发摘要
   - 清除对话
   - 验证归档

2. **迁移测试**
   - 创建旧格式数据
   - 运行迁移
   - 验证数据完整性
   - 验证功能正常

3. **多会话并发测试**
   - 多个会话同时操作
   - 验证数据隔离
   - 验证无竞态条件

### 性能测试

1. **消息检索性能**
   - 测试 1000 条消息的查询时间（目标: <50ms）
   - 测试 10000 条消息的查询时间（目标: <200ms）
   - 验证索引效果

2. **并发性能**
   - 100 个并发会话
   - 每个会话 10 条消息/秒
   - 验证响应时间和吞吐量

3. **数据库性能**
   - 测试不同数据库后端（SQLite, MySQL, PostgreSQL）
   - 比较性能差异
   - 优化慢查询

### 测试数据生成器

```go
// 生成随机消息
func GenerateRandomMessage() *storage.Message {
    roles := []string{"user", "assistant", "system"}
    return &storage.Message{
        Role:      roles[rand.Intn(len(roles))],
        Content:   randomString(100),
        Timestamp: time.Now().Unix(),
        Hidden:    rand.Float32() < 0.2, // 20% 隐藏
        Truncated: false,
    }
}

// 生成随机会话上下文
func GenerateRandomSessionContext() *storage.SessionContext {
    return &storage.SessionContext{
        ChatID: rand.Int63(),
        BotID:  rand.Int63(),
        UserID: randomOptionalInt64(),
        ThreadID: randomOptionalInt64(),
    }
}

// 生成随机对话
func GenerateRandomConversation(messageCount int) []*storage.Message {
    messages := make([]*storage.Message, messageCount)
    for i := 0; i < messageCount; i++ {
        messages[i] = GenerateRandomMessage()
        messages[i].Timestamp = int64(i) // 确保顺序
    }
    return messages
}
```

### 测试覆盖率目标

- 单元测试覆盖率: ≥ 80%
- 属性测试覆盖所有正确性属性: 100%
- 集成测试覆盖主要用户流程: 100%
- 性能测试覆盖关键操作: 100%

## 正确性属性

属性是关于系统应该满足的特征或行为的形式化陈述，它应该在所有有效执行中保持为真。属性是人类可读规范和机器可验证正确性保证之间的桥梁。

### 属性 1: 消息持久化完整性

*对于任何*会话上下文和消息，当保存消息后立即检索，应该返回包含所有原始字段（角色、内容、时间戳、会话标识符）的消息记录

**验证: 需求 2.1, 2.3**

### 属性 2: 消息时间顺序性

*对于任何*会话的消息集合，检索时返回的消息应该按时间戳严格递增排序

**验证: 需求 2.2**

### 属性 3: 隐藏消息过滤

*对于任何*包含隐藏和非隐藏消息的会话，当检索用于 AI 请求的历史时，返回的消息集合应该不包含任何隐藏标志为 true 的消息

**验证: 需求 2.7, 5.7**

### 属性 4: 摘要触发隐藏标记

*对于任何*触发摘要的会话，摘要完成后，被摘要的旧消息应该全部标记为隐藏（hidden=true）但仍存在于数据库中

**验证: 需求 2.6, 5.6**

### 属性 5: Clear 命令完整流程

*对于任何*包含消息的会话，执行 clear 命令后应该：
1. 生成包含所有用户和助手消息的 Telegraph 归档
2. 向用户发送包含归档 URL 的消息
3. 从数据库中删除该会话的所有消息记录

**验证: 需求 3.1, 3.2, 3.3, 3.4**

### 属性 6: 调试模式管理员通知

*对于任何*在调试模式启用时执行的 clear 命令，所有配置的管理员用户应该收到包含源用户 ChatID、用户名和归档 URL 的通知消息

**验证: 需求 4.3, 4.4, 4.6**

### 属性 7: 调试模式禁用时无通知

*对于任何*在调试模式禁用时执行的 clear 命令，管理员用户不应该收到任何通知消息

**验证: 需求 4.5**

### 属性 8: 上下文长度自动管理

*对于任何*会话，当添加新消息导致总令牌数超过 MAX_CONTEXT_LENGTH 时，系统应该自动删除最早的非系统、非摘要消息直到总令牌数低于限制

**验证: 需求 5.2**

### 属性 9: 摘要阈值触发

*对于任何*会话，当上下文使用率达到或超过 SUMMARY_THRESHOLD 时，系统应该触发摘要生成

**验证: 需求 5.3**

### 属性 10: 摘要后消息保留

*对于任何*触发摘要的会话，摘要完成后应该保留：
1. 所有系统消息
2. 新生成的摘要消息
3. 最近 MIN_RECENT_PAIRS 对用户-助手消息

**验证: 需求 5.4**

### 属性 11: 构建上下文包含系统提示

*对于任何*会话，构建用于 AI 请求的上下文时，应该始终包含系统提示消息（如果存在）

**验证: 需求 5.5**

### 属性 12: 路由路径正确性

*对于任何*HTTP 请求，如果路径以 `/api/manager/` 开头，应该路由到管理器处理器；否则应该路由到主应用处理器

**验证: 需求 6.2, 6.3**

## 部署计划

### 阶段 1: 准备（第 1 周）
- 更新包名
- 创建新的 Message 模型
- 实现存储接口

### 阶段 2: 迁移（第 2 周）
- 实现迁移脚本
- 在测试环境运行迁移
- 验证数据完整性

### 阶段 3: 功能更新（第 3 周）
- 更新上下文管理器
- 更新 Clear 命令
- 添加调试模式

### 阶段 4: 测试和部署（第 4 周）
- 运行完整测试套件
- 性能测试
- 生产环境部署
- 监控和调优

## 回滚计划

如果部署后发现问题：

1. **立即回滚**
   - 恢复旧版本代码
   - 数据库保持不变（新表不影响旧代码）

2. **数据恢复**
   - 如果需要，从备份恢复 `chat_histories` 表
   - 新的 `messages` 表可以保留或删除

3. **渐进式修复**
   - 修复问题
   - 在测试环境重新验证
   - 重新部署
