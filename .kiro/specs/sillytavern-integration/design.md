# 设计文档

## 概述

本设计文档描述了将 SillyTavern 核心功能集成到现有 Telegram bot 的架构和实现方案。该集成将保留 SillyTavern 的请求构建逻辑、角色卡系统、世界书、预设和正则处理功能，同时通过 Telegram 界面和 Web 管理器提供访问。

### 核心目标

1. 集成 SillyTavern 的角色卡、预设、世界书和正则功能
2. 实现智能上下文管理和自动摘要
3. 提供 Web 管理器用于配置管理
4. 保持与现有 bot 架构的兼容性
5. 支持基于权限的配置控制

### 设计原则

- 模块化：各功能模块独立，易于维护和扩展
- 兼容性：保持与现有代码库的兼容
- 性能：优化数据库查询和内存使用
- 安全性：实现安全的认证和授权机制

## 架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Telegram Bot                            │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Commands   │  │   Handlers   │  │   Manager    │      │
│  │   /login     │  │   Message    │  │   Web UI     │      │
│  │   /share     │  │   Callback   │  │   REST API   │      │
│  │   /clear     │  │   Update     │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────────┤
│                   SillyTavern Integration Layer              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Character   │  │  World Book  │  │   Presets    │      │
│  │    Cards     │  │   Manager    │  │   Manager    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Regex     │  │   Request    │  │   Context    │      │
│  │   Processor  │  │   Builder    │  │   Manager    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────────┤
│                      Storage Layer                           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Database (SQLite/MySQL/PostgreSQL)      │   │
│  │  - chat_histories  - character_cards                 │   │
│  │  - world_books     - presets                         │   │
│  │  - regex_patterns  - login_tokens                    │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

1. **消息接收流程**:
   - Telegram → Handler → Context Manager → Request Builder → AI Provider
   
2. **响应处理流程**:
   - AI Provider → Regex Processor → Context Manager → Handler → Telegram

3. **管理器流程**:
   - Web UI → REST API → Storage Layer → Database

## 组件和接口

### 1. 数据库模型扩展

需要在现有数据库基础上添加新表：


#### CharacterCard 表

```go
type CharacterCard struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // 所有者信息
    UserID    *int64 `gorm:"index"` // nil 表示全局角色卡
    
    // 角色卡元数据
    Name      string `gorm:"not null;index"`
    Avatar    string `gorm:"type:text"` // 头像 URL 或 base64
    
    // SillyTavern V2 格式数据
    Data      string `gorm:"type:text;not null"` // JSON 格式的完整角色卡数据
    
    // 快速访问字段
    IsActive  bool   `gorm:"default:false;index"`
}
```

#### WorldBook 表

```go
type WorldBook struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // 所有者信息
    UserID    *int64 `gorm:"index"` // nil 表示全局世界书
    
    // 世界书元数据
    Name      string `gorm:"not null;index"`
    
    // 世界书数据
    Data      string `gorm:"type:text;not null"` // JSON 格式的世界书数据
    
    // 状态
    IsActive  bool   `gorm:"default:false;index"`
}
```

#### WorldBookEntry 表

```go
type WorldBookEntry struct {
    ID           uint      `gorm:"primarykey"`
    CreatedAt    time.Time `gorm:"index"`
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
    
    // 关联
    WorldBookID  uint   `gorm:"not null;index"`
    
    // 条目数据
    UID          string `gorm:"not null;uniqueIndex"`
    Keys         string `gorm:"type:text;not null"` // JSON 数组
    SecondaryKeys string `gorm:"type:text"`         // JSON 数组
    Content      string `gorm:"type:text;not null"`
    Comment      string `gorm:"type:text"`
    
    // 配置
    Constant     bool   `gorm:"default:false"`
    Selective    bool   `gorm:"default:false"`
    Order        int    `gorm:"default:100"`
    Position     string `gorm:"default:'after_char'"` // before_char, after_char
    Enabled      bool   `gorm:"default:true;index"`
    
    // 扩展字段
    Extensions   string `gorm:"type:text"` // JSON 格式
}
```

#### Preset 表

```go
type Preset struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // 所有者信息
    UserID    *int64 `gorm:"index"` // nil 表示全局预设
    
    // 预设元数据
    Name      string `gorm:"not null;index"`
    APIType   string `gorm:"not null;index"` // openai, anthropic, etc.
    
    // 预设数据
    Data      string `gorm:"type:text;not null"` // JSON 格式的预设参数
    
    // 状态
    IsActive  bool   `gorm:"default:false;index"`
}
```

#### RegexPattern 表

```go
type RegexPattern struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // 所有者信息
    UserID    *int64 `gorm:"index"` // nil 表示全局正则
    
    // 正则元数据
    Name      string `gorm:"not null;index"`
    
    // 正则配置
    Pattern   string `gorm:"type:text;not null"` // 正则表达式
    Replace   string `gorm:"type:text;not null"` // 替换文本
    Type      string `gorm:"not null"`           // input, output
    Order     int    `gorm:"default:100"`
    Enabled   bool   `gorm:"default:true;index"`
}
```

#### LoginToken 表

```go
type LoginToken struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time
    
    // 用户信息
    UserID    int64  `gorm:"not null;uniqueIndex"`
    
    // 令牌
    Token     string `gorm:"not null;uniqueIndex"`
    
    // 过期时间
    ExpiresAt time.Time `gorm:"not null;index"`
}
```

### 2. 消息历史扩展

修改现有的 `HistoryItem` 结构以支持消息类型：

```go
type HistoryItem struct {
    Role      string      `json:"role"`    // "user", "assistant", "system", "summary"
    Content   interface{} `json:"content"` // string or []ContentPart
    Timestamp int64       `json:"timestamp,omitempty"` // Unix timestamp
    Truncated bool        `json:"truncated,omitempty"` // 是否被 /clear 截断
}
```

### 3. SillyTavern 集成层

#### CharacterCardManager

```go
type CharacterCardManager struct {
    storage Storage
}

// 加载角色卡
func (m *CharacterCardManager) LoadCard(userID *int64, cardID uint) (*CharacterCard, error)

// 保存角色卡
func (m *CharacterCardManager) SaveCard(card *CharacterCard) error

// 列出角色卡
func (m *CharacterCardManager) ListCards(userID *int64) ([]*CharacterCard, error)

// 激活角色卡
func (m *CharacterCardManager) ActivateCard(userID *int64, cardID uint) error

// 获取活动角色卡
func (m *CharacterCardManager) GetActiveCard(userID *int64) (*CharacterCard, error)

// 上传角色卡（从 PNG 文件解析）
func (m *CharacterCardManager) UploadCard(userID *int64, imageData []byte) (*CharacterCard, error)
```

#### WorldBookManager

```go
type WorldBookManager struct {
    storage Storage
}

// 加载世界书
func (m *WorldBookManager) LoadBook(userID *int64, bookID uint) (*WorldBook, error)

// 保存世界书
func (m *WorldBookManager) SaveBook(book *WorldBook) error

// 列出世界书
func (m *WorldBookManager) ListBooks(userID *int64) ([]*WorldBook, error)

// 激活世界书
func (m *WorldBookManager) ActivateBook(userID *int64, bookID uint) error

// 获取活动世界书
func (m *WorldBookManager) GetActiveBook(userID *int64) (*WorldBook, error)

// 触发世界书条目（基于消息内容）
func (m *WorldBookManager) TriggerEntries(bookID uint, messages []HistoryItem) ([]*WorldBookEntry, error)

// 更新条目状态
func (m *WorldBookManager) UpdateEntryStatus(entryID uint, enabled bool) error

// 编辑条目
func (m *WorldBookManager) UpdateEntry(entry *WorldBookEntry) error
```

#### PresetManager

```go
type PresetManager struct {
    storage Storage
}

// 加载预设
func (m *PresetManager) LoadPreset(userID *int64, presetID uint) (*Preset, error)

// 保存预设
func (m *PresetManager) SavePreset(preset *Preset) error

// 列出预设
func (m *PresetManager) ListPresets(userID *int64, apiType string) ([]*Preset, error)

// 激活预设
func (m *PresetManager) ActivatePreset(userID *int64, presetID uint) error

// 获取活动预设
func (m *PresetManager) GetActivePreset(userID *int64, apiType string) (*Preset, error)
```

#### RegexProcessor

```go
type RegexProcessor struct {
    storage Storage
}

// 应用输入正则
func (p *RegexProcessor) ProcessInput(userID *int64, text string) (string, error)

// 应用输出正则
func (p *RegexProcessor) ProcessOutput(userID *int64, text string) (string, error)

// 列出正则模式
func (p *RegexProcessor) ListPatterns(userID *int64, patternType string) ([]*RegexPattern, error)

// 更新模式状态
func (p *RegexProcessor) UpdatePatternStatus(patternID uint, enabled bool) error

// 编辑模式
func (p *RegexProcessor) UpdatePattern(pattern *RegexPattern) error
```

### 4. 请求构建器

这是核心组件，负责将 SillyTavern 的各种元素组合成 AI 请求。

```go
type RequestBuilder struct {
    characterManager *CharacterCardManager
    worldBookManager *WorldBookManager
    presetManager    *PresetManager
    regexProcessor   *RegexProcessor
}

// 构建请求的主要方法
func (b *RequestBuilder) BuildRequest(ctx *BuildContext) (*AIRequest, error)

// 构建上下文
type BuildContext struct {
    UserID       *int64
    History      []HistoryItem
    CurrentInput string
    APIType      string
}

// AI 请求（OpenAI 格式）
type AIRequest struct {
    Messages []Message `json:"messages"`
    Model    string    `json:"model"`
    Temperature float64 `json:"temperature,omitempty"`
    MaxTokens   int    `json:"max_tokens,omitempty"`
    TopP        float64 `json:"top_p,omitempty"`
    // ... 其他参数
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

#### 请求构建流程

1. **加载活动配置**
   - 加载活动角色卡
   - 加载活动世界书
   - 加载活动预设

2. **处理输入**
   - 应用输入正则转换
   - 添加到历史记录

3. **构建系统提示**
   - 从角色卡提取 system_prompt
   - 添加角色描述、个性等

4. **注入世界书**
   - 扫描历史记录触发关键词
   - 按优先级和位置注入条目

5. **强制角色交替**
   - 确保 user/assistant 交替
   - 合并连续相同角色的消息

6. **应用预设参数**
   - 设置 temperature、top_p 等
   - 应用模型特定配置

7. **记录调试信息**
   - 打印最终消息序列
   - 记录注入的世界书条目

### 5. 上下文管理器

负责管理对话历史、摘要和上下文窗口。

```go
type ContextManager struct {
    storage Storage
    config  *ContextConfig
}

type ContextConfig struct {
    MaxContextLength  int     // 最大上下文长度（tokens）
    SummaryThreshold  float64 // 触发摘要的阈值（0.0-1.0）
    MinRecentPairs    int     // 保留的最小最近消息对数
}

// 添加消息到历史
func (m *ContextManager) AddMessage(ctx *SessionContext, role string, content interface{}) error

// 获取构建请求用的历史（应用摘要逻辑）
func (m *ContextManager) GetBuildHistory(ctx *SessionContext) ([]HistoryItem, error)

// 触发摘要
func (m *ContextManager) TriggerSummary(ctx *SessionContext) error

// 清除历史（创建截断标记）
func (m *ContextManager) ClearHistory(ctx *SessionContext) error

// 获取完整历史（用于分享）
func (m *ContextManager) GetFullHistory(ctx *SessionContext) ([]HistoryItem, error)

// 估算 token 数量
func (m *ContextManager) EstimateTokens(messages []HistoryItem) (int, error)
```

#### 摘要逻辑

当上下文长度超过 `MaxContextLength * SummaryThreshold` 时：

1. 保留最近的 `MinRecentPairs` 对 user/assistant 消息
2. 将其余消息发送给 AI 进行摘要
3. 创建一个 role="summary" 的消息
4. 标记旧消息为已摘要（不删除）

构建请求时的历史结构：
```
[system messages] + [summary] + [recent user/assistant pairs]
```

### 6. 命令处理器

#### /login 命令

```go
type LoginCommand struct {
    storage Storage
}

func (c *LoginCommand) Execute(ctx *CommandContext) error {
    // 1. 检查是否为私聊
    if ctx.ChatID < 0 {
        return errors.New("此命令仅支持私聊")
    }
    
    // 2. 生成令牌
    token := generateSecureToken()
    
    // 3. 保存到数据库
    loginToken := &LoginToken{
        UserID:    ctx.UserID,
        Token:     token,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    // 4. 返回格式化消息
    message := fmt.Sprintf(
        "用户名：%d\n密码：%s\n有效期：24小时",
        ctx.UserID,
        token,
    )
    
    return ctx.Reply(message)
}
```

#### /share 命令

```go
type ShareCommand struct {
    contextManager *ContextManager
    telegraph      *TelegraphClient
}

func (c *ShareCommand) Execute(ctx *CommandContext) error {
    // 1. 获取完整历史
    history, err := c.contextManager.GetFullHistory(ctx.Session)
    
    // 2. 过滤 user 和 assistant 消息
    filtered := filterMessages(history, []string{"user", "assistant"})
    
    // 3. 格式化为 HTML
    html := formatForTelegraph(filtered)
    
    // 4. 发布到 Telegraph
    url, err := c.telegraph.CreatePage(title, html)
    
    // 5. 返回 URL
    return ctx.Reply(fmt.Sprintf("对话已分享：%s", url))
}
```

#### /clear 命令

```go
type ClearCommand struct {
    contextManager *ContextManager
}

func (c *ClearCommand) Execute(ctx *CommandContext) error {
    // 创建截断标记，不删除历史
    return c.contextManager.ClearHistory(ctx.Session)
}
```

#### /clear_all_chat 命令

```go
type ClearAllChatCommand struct {
    storage Storage
    config  *config.Config
}

func (c *ClearAllChatCommand) Execute(ctx *CommandContext) error {
    // 1. 检查管理员权限
    if !isAdmin(ctx.UserID, c.config) {
        return errors.New("权限不足")
    }
    
    // 2. 删除所有历史
    err := c.storage.DeleteAllChatHistory()
    
    // 3. 记录审计日志
    logAudit("clear_all_chat", ctx.UserID, time.Now())
    
    return err
}
```

### 7. Web 管理器

Web 管理器提供 REST API 和前端界面用于管理 SillyTavern 功能。

#### 认证中间件

```go
type AuthMiddleware struct {
    storage Storage
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 从请求头获取用户 ID 和令牌
        userID := r.Header.Get("X-User-ID")
        token := r.Header.Get("X-Auth-Token")
        
        // 2. 验证令牌
        valid, err := m.storage.ValidateLoginToken(userID, token)
        if !valid || err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // 3. 将用户 ID 添加到上下文
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### REST API 端点

```go
// 角色卡管理
GET    /api/manager/characters          // 列出角色卡
POST   /api/manager/characters          // 上传角色卡
GET    /api/manager/characters/:id      // 获取角色卡详情
PUT    /api/manager/characters/:id      // 更新角色卡
DELETE /api/manager/characters/:id      // 删除角色卡
POST   /api/manager/characters/:id/activate // 激活角色卡

// 世界书管理
GET    /api/manager/worldbooks          // 列出世界书
POST   /api/manager/worldbooks          // 上传世界书
GET    /api/manager/worldbooks/:id      // 获取世界书详情
PUT    /api/manager/worldbooks/:id      // 更新世界书
DELETE /api/manager/worldbooks/:id      // 删除世界书
POST   /api/manager/worldbooks/:id/activate // 激活世界书

// 世界书条目管理
GET    /api/manager/worldbooks/:id/entries     // 列出条目
PUT    /api/manager/worldbooks/:id/entries/:eid // 更新条目
POST   /api/manager/worldbooks/:id/entries/:eid/toggle // 切换启用状态

// 预设管理
GET    /api/manager/presets             // 列出预设
POST   /api/manager/presets             // 创建预设
GET    /api/manager/presets/:id         // 获取预设详情
PUT    /api/manager/presets/:id         // 更新预设
DELETE /api/manager/presets/:id         // 删除预设
POST   /api/manager/presets/:id/activate // 激活预设

// 正则管理
GET    /api/manager/regex               // 列出正则模式
POST   /api/manager/regex               // 创建正则模式
GET    /api/manager/regex/:id           // 获取正则详情
PUT    /api/manager/regex/:id           // 更新正则
DELETE /api/manager/regex/:id           // 删除正则
POST   /api/manager/regex/:id/toggle    // 切换启用状态

// 全局配置（仅管理员）
GET    /api/manager/config              // 获取全局配置
PUT    /api/manager/config              // 更新全局配置
```

#### 权限控制

基于 `ENABLE_USER_SETTING` 配置：

```go
type PermissionChecker struct {
    config *config.Config
}

func (p *PermissionChecker) CanModifyGlobal(userID int64) bool {
    // 检查是否为管理员
    return isAdmin(userID, p.config)
}

func (p *PermissionChecker) CanModifyPersonal(userID int64) bool {
    // 如果 ENABLE_USER_SETTING=true，所有用户都可以修改自己的
    if p.config.EnableUserSetting {
        return true
    }
    // 否则只有管理员可以
    return isAdmin(userID, p.config)
}

func (p *PermissionChecker) CanAccessResource(userID int64, resourceUserID *int64) bool {
    // 全局资源（resourceUserID == nil）
    if resourceUserID == nil {
        return true
    }
    
    // 自己的资源
    if *resourceUserID == userID {
        return true
    }
    
    // 管理员可以访问所有资源
    return isAdmin(userID, p.config)
}
```

### 8. Telegraph 集成

```go
type TelegraphClient struct {
    accessToken string
}

func NewTelegraphClient() (*TelegraphClient, error) {
    // 创建 Telegraph 账号
    resp, err := http.Post(
        "https://api.telegra.ph/createAccount",
        "application/json",
        bytes.NewBuffer([]byte(`{"short_name":"TelegramBot"}`)),
    )
    // 解析 access_token
    // ...
}

func (c *TelegraphClient) CreatePage(title string, content string) (string, error) {
    // 调用 Telegraph API 创建页面
    data := map[string]interface{}{
        "access_token": c.accessToken,
        "title":        title,
        "content":      content,
        "return_content": false,
    }
    
    resp, err := http.Post(
        "https://api.telegra.ph/createPage",
        "application/json",
        marshalJSON(data),
    )
    
    // 解析响应获取 URL
    // ...
}
```

## 数据模型

### 角色卡数据结构（SillyTavern V2 格式）

```json
{
  "spec": "chara_card_v2",
  "spec_version": "2.0",
  "data": {
    "name": "角色名称",
    "description": "角色描述",
    "personality": "个性特征",
    "scenario": "场景设定",
    "first_mes": "第一条消息",
    "mes_example": "对话示例",
    "creator_notes": "创作者备注",
    "system_prompt": "系统提示",
    "post_history_instructions": "历史后指令",
    "alternate_greetings": ["备选问候1", "备选问候2"],
    "character_book": {
      "entries": [
        {
          "keys": ["关键词1", "关键词2"],
          "content": "注入内容",
          "enabled": true,
          "insertion_order": 100,
          "position": "after_char"
        }
      ]
    },
    "tags": ["标签1", "标签2"],
    "creator": "创作者",
    "character_version": "1.0",
    "extensions": {
      "talkativeness": 0.5,
      "fav": false,
      "world": "world_book_name",
      "depth_prompt": {
        "prompt": "深度提示",
        "depth": 4,
        "role": "system"
      }
    }
  }
}
```

### 世界书数据结构

```json
{
  "name": "世界书名称",
  "entries": {
    "entry_uid_1": {
      "uid": "entry_uid_1",
      "key": ["触发词1", "触发词2"],
      "keysecondary": ["次要触发词"],
      "comment": "备注",
      "content": "注入内容",
      "constant": false,
      "selective": true,
      "order": 100,
      "position": 0,
      "disable": false,
      "excludeRecursion": false,
      "probability": 100,
      "useProbability": false,
      "depth": 4,
      "selectiveLogic": 0,
      "extensions": {}
    }
  },
  "extensions": {}
}
```

### 预设数据结构

```json
{
  "name": "预设名称",
  "temperature": 0.7,
  "top_p": 0.9,
  "top_k": 40,
  "max_tokens": 2048,
  "presence_penalty": 0.0,
  "frequency_penalty": 0.0,
  "repetition_penalty": 1.0,
  "stop_sequences": ["\\n\\nHuman:", "\\n\\nAssistant:"]
}
```

## 错误处理

### 错误类型

```go
var (
    ErrNotFound           = errors.New("resource not found")
    ErrUnauthorized       = errors.New("unauthorized")
    ErrInvalidFormat      = errors.New("invalid format")
    ErrTokenExpired       = errors.New("token expired")
    ErrContextTooLong     = errors.New("context too long")
    ErrInvalidRegex       = errors.New("invalid regex pattern")
    ErrDatabaseError      = errors.New("database error")
)
```

### 错误处理策略

1. **数据库错误**: 记录日志，返回通用错误消息
2. **验证错误**: 返回详细的验证失败信息
3. **认证错误**: 返回 401 状态码
4. **权限错误**: 返回 403 状态码
5. **资源不存在**: 返回 404 状态码

### 日志记录

```go
type Logger struct {
    level LogLevel
}

// 记录请求构建过程
func (l *Logger) LogRequestBuild(ctx *BuildContext, request *AIRequest) {
    log.Printf("[RequestBuilder] User: %v, Messages: %d, Model: %s",
        ctx.UserID, len(request.Messages), request.Model)
    
    for i, msg := range request.Messages {
        log.Printf("  [%d] Role: %s, Content: %s",
            i, msg.Role, truncate(msg.Content, 100))
    }
}

// 记录世界书触发
func (l *Logger) LogWorldBookTrigger(entries []*WorldBookEntry) {
    log.Printf("[WorldBook] Triggered %d entries", len(entries))
    for _, entry := range entries {
        log.Printf("  - UID: %s, Keys: %v", entry.UID, entry.Keys)
    }
}

// 记录摘要操作
func (l *Logger) LogSummary(ctx *SessionContext, oldCount, newCount int) {
    log.Printf("[Summary] Session: %v, Messages: %d -> %d",
        ctx, oldCount, newCount)
}
```

## 测试策略

### 单元测试

针对每个组件编写单元测试：

1. **CharacterCardManager**: 测试加载、保存、激活功能
2. **WorldBookManager**: 测试触发逻辑、条目管理
3. **PresetManager**: 测试预设加载和应用
4. **RegexProcessor**: 测试正则转换
5. **RequestBuilder**: 测试请求构建逻辑
6. **ContextManager**: 测试摘要和历史管理

### 集成测试

测试组件之间的交互：

1. **完整请求流程**: 从消息接收到 AI 响应
2. **管理器 API**: 测试所有 REST 端点
3. **权限控制**: 测试不同权限级别的访问
4. **数据持久化**: 测试数据库操作

### 性能测试

1. **上下文管理**: 测试大量历史消息的处理性能
2. **世界书触发**: 测试大量条目的匹配性能
3. **数据库查询**: 测试复杂查询的性能
4. **并发处理**: 测试多用户同时访问


## 正确性属性

属性是关于系统应该做什么的正式陈述，它应该在所有有效执行中保持为真。属性作为人类可读规范和机器可验证正确性保证之间的桥梁。

### 属性 1: 资源上传和检索一致性

*对于任何*有效的 SillyTavern 资源（角色卡、预设、世界书、正则模式），如果通过管理器上传它，然后通过相同的用户 ID 检索，应该得到等效的资源数据。

**验证: 需求 1.1, 1.2, 1.3, 1.4**

### 属性 2: 世界书触发正确性

*对于任何*世界书条目和消息历史，如果消息中包含条目的触发关键词且条目已启用，则该条目应该被包含在构建的请求中。

**验证: 需求 1.3**

### 属性 3: 正则转换幂等性

*对于任何*正则模式和文本，应用正则转换后的结果再次应用相同的转换应该产生相同的结果（如果模式设计为幂等）。

**验证: 需求 1.4**

### 属性 4: 自动摘要触发

*对于任何*会话，当上下文长度（以 tokens 计）超过 MAX_CONTEXT_LENGTH * SUMMARY_THRESHOLD 时，系统应该自动触发摘要操作。

**验证: 需求 2.1**

### 属性 5: 摘要后消息结构

*对于任何*触发摘要后的会话，构建请求时应该只包含：系统消息 + 摘要消息 + 最近的 N 对 user/assistant 消息（N 由配置决定）。

**验证: 需求 2.2, 6.2, 6.3**

### 属性 6: 摘要不删除历史

*对于任何*会话，触发摘要操作后，所有原始消息应该仍然存在于数据库中，只是被标记为已摘要。

**验证: 需求 2.5**

### 属性 7: 严格角色交替

*对于任何*消息历史，构建的 AI 请求中的消息应该严格在 user 和 assistant 角色之间交替，连续相同角色的消息应该被合并。

**验证: 需求 3.2, 9.1, 9.2**

### 属性 8: OpenAI 格式合规性

*对于任何*构建的请求，输出应该符合 OpenAI Chat Completion API 的格式规范，包含 messages 数组和必需的参数。

**验证: 需求 3.3**

### 属性 9: 登录令牌有效期

*对于任何*生成的登录令牌，它应该在创建后 24 小时内有效，之后应该被自动清理。

**验证: 需求 4.2, 4.4**

### 属性 10: 权限控制一致性

*对于任何*用户和资源，当 ENABLE_USER_SETTING=false 时，只有管理员可以修改设置；当 ENABLE_USER_SETTING=true 时，用户可以修改自己的设置，管理员可以修改任何设置。

**验证: 需求 5.1, 5.2, 5.3, 5.4, 5.5**

### 属性 11: 消息类型分类

*对于任何*添加到历史的消息，它应该被正确分类为 user、assistant、system 或 summary 之一。

**验证: 需求 6.1**

### 属性 12: /clear 不删除历史

*对于任何*会话，执行 /clear 命令后，所有历史消息应该仍然存在于数据库中，但应该有一个截断标记，后续加载历史时只加载标记之后的消息。

**验证: 需求 6.4, 6.5**

### 属性 13: /share 包含完整对话

*对于任何*会话，执行 /share 命令应该收集所有 user 和 assistant 消息，不受摘要状态影响，并生成有效的 Telegraph URL。

**验证: 需求 8.1, 8.2, 8.3**


## 错误处理

### 错误分类

1. **用户错误**: 无效输入、格式错误、权限不足
2. **系统错误**: 数据库故障、网络错误、资源耗尽
3. **集成错误**: AI API 错误、Telegraph API 错误

### 错误处理策略

```go
// 用户友好的错误消息
func handleError(err error, ctx *Context) {
    switch {
    case errors.Is(err, ErrUnauthorized):
        ctx.Reply("❌ 权限不足，请先使用 /login 命令获取访问权限")
    case errors.Is(err, ErrTokenExpired):
        ctx.Reply("❌ 登录令牌已过期，请重新使用 /login 命令")
    case errors.Is(err, ErrInvalidFormat):
        ctx.Reply("❌ 文件格式无效，请上传有效的 SillyTavern 格式文件")
    case errors.Is(err, ErrContextTooLong):
        ctx.Reply("⚠️ 上下文过长，正在自动摘要...")
        // 触发摘要
    case errors.Is(err, ErrDatabaseError):
        ctx.Reply("❌ 系统错误，请稍后重试")
        log.Error("Database error:", err)
    default:
        ctx.Reply("❌ 发生未知错误")
        log.Error("Unknown error:", err)
    }
}
```

### 重试机制

对于临时性错误（网络错误、API 限流），实现指数退避重试：

```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(backoff)
    }
    return err
}
```

## 测试策略

### 单元测试

使用 Go 的标准测试框架和 testify 库：

```go
func TestCharacterCardManager_UploadCard(t *testing.T) {
    // 测试上传有效角色卡
    manager := NewCharacterCardManager(mockStorage)
    cardData := generateValidCardData()
    
    card, err := manager.UploadCard(&userID, cardData)
    assert.NoError(t, err)
    assert.NotNil(t, card)
    assert.Equal(t, "TestCharacter", card.Name)
}

func TestRequestBuilder_EnforceRoleAlternation(t *testing.T) {
    // 测试角色交替
    builder := NewRequestBuilder(...)
    history := []HistoryItem{
        {Role: "user", Content: "Hello"},
        {Role: "user", Content: "How are you?"},
        {Role: "assistant", Content: "I'm fine"},
    }
    
    request, err := builder.BuildRequest(&BuildContext{
        History: history,
    })
    
    assert.NoError(t, err)
    // 验证角色交替
    for i := 1; i < len(request.Messages); i++ {
        assert.NotEqual(t, request.Messages[i-1].Role, request.Messages[i].Role)
    }
}
```

### 属性测试

使用 gopter 库进行属性测试：

```go
func TestProperty_UploadRetrieveConsistency(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("上传后检索应该得到相同数据", 
        prop.ForAll(
            func(cardData CharacterCardData) bool {
                manager := NewCharacterCardManager(testStorage)
                
                // 上传
                card, err := manager.UploadCard(&userID, cardData)
                if err != nil {
                    return false
                }
                
                // 检索
                retrieved, err := manager.LoadCard(&userID, card.ID)
                if err != nil {
                    return false
                }
                
                // 验证等效性
                return areEquivalent(card, retrieved)
            },
            genCharacterCardData(),
        ),
    )
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### 集成测试

测试完整的端到端流程：

```go
func TestIntegration_CompleteConversationFlow(t *testing.T) {
    // 设置测试环境
    db := setupTestDatabase()
    bot := setupTestBot(db)
    
    // 1. 用户登录
    loginResp := bot.SendCommand("/login", privateChat)
    assert.Contains(t, loginResp, "用户名：")
    
    // 2. 上传角色卡
    token := extractToken(loginResp)
    uploadResp := uploadCharacterCard(token, testCardData)
    assert.Equal(t, 200, uploadResp.StatusCode)
    
    // 3. 发送消息
    msgResp := bot.SendMessage("Hello", privateChat)
    assert.NotEmpty(t, msgResp)
    
    // 4. 验证世界书触发
    history := db.GetHistory(privateChat)
    assert.Contains(t, history, worldBookContent)
    
    // 5. 测试分享
    shareResp := bot.SendCommand("/share", privateChat)
    assert.Contains(t, shareResp, "telegra.ph")
}
```

### 性能测试

```go
func BenchmarkRequestBuilder_BuildRequest(b *testing.B) {
    builder := setupRequestBuilder()
    ctx := &BuildContext{
        History: generateLargeHistory(1000),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := builder.BuildRequest(ctx)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkWorldBook_TriggerEntries(b *testing.B) {
    manager := setupWorldBookManager()
    messages := generateMessages(100)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := manager.TriggerEntries(bookID, messages)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 部署考虑

### 环境变量

新增的环境变量：

```bash
# 上下文管理
MAX_CONTEXT_LENGTH=8000        # 最大上下文长度（tokens）
SUMMARY_THRESHOLD=0.8          # 触发摘要的阈值（0.0-1.0）
MIN_RECENT_PAIRS=2             # 保留的最小最近消息对数

# 管理器配置
MANAGER_PORT=8081              # 管理器 Web 服务端口
MANAGER_ENABLED=true           # 是否启用管理器

# Telegraph 配置
TELEGRAPH_ENABLED=true         # 是否启用 Telegraph 分享功能
```

### 数据库迁移

使用 GORM 的 AutoMigrate 功能：

```go
func migrateDatabase(db *gorm.DB) error {
    return db.AutoMigrate(
        &CharacterCard{},
        &WorldBook{},
        &WorldBookEntry{},
        &Preset{},
        &RegexPattern{},
        &LoginToken{},
    )
}
```

### 向后兼容性

1. **现有历史记录**: 自动迁移现有历史记录，添加 role 和 timestamp 字段
2. **配置兼容**: 保留 SYSTEM_INIT_MESSAGE 作为后备，如果没有角色卡则使用
3. **API 兼容**: 保持现有 Telegram 命令的兼容性

### 监控和日志

```go
// 关键指标
type Metrics struct {
    RequestsBuilt      prometheus.Counter
    SummariesTriggered prometheus.Counter
    WorldBookHits      prometheus.Counter
    APIErrors          prometheus.Counter
    AverageContextSize prometheus.Histogram
}

// 结构化日志
log.WithFields(log.Fields{
    "user_id":     userID,
    "session_id":  sessionID,
    "action":      "build_request",
    "messages":    len(messages),
    "world_books": len(triggeredEntries),
}).Info("Request built successfully")
```

## 安全考虑

### 输入验证

1. **文件上传**: 验证文件大小、格式、内容
2. **正则模式**: 验证正则表达式的安全性，防止 ReDoS 攻击
3. **SQL 注入**: 使用 GORM 的参数化查询
4. **XSS**: 对 Telegraph 内容进行 HTML 转义

### 认证和授权

1. **令牌安全**: 使用加密安全的随机数生成器
2. **令牌存储**: 在数据库中存储令牌的哈希值
3. **HTTPS**: 强制管理器使用 HTTPS
4. **CORS**: 配置适当的 CORS 策略

### 速率限制

```go
type RateLimiter struct {
    requests map[int64]*rate.Limiter
    mu       sync.RWMutex
}

func (r *RateLimiter) Allow(userID int64) bool {
    r.mu.RLock()
    limiter, exists := r.requests[userID]
    r.mu.RUnlock()
    
    if !exists {
        r.mu.Lock()
        limiter = rate.NewLimiter(rate.Every(time.Second), 10)
        r.requests[userID] = limiter
        r.mu.Unlock()
    }
    
    return limiter.Allow()
}
```

## 未来扩展

### 可能的增强功能

1. **多角色卡支持**: 允许在同一会话中切换角色
2. **角色卡版本控制**: 跟踪角色卡的修改历史
3. **世界书继承**: 支持世界书之间的继承关系
4. **高级摘要**: 使用更智能的摘要算法
5. **导出功能**: 导出对话为各种格式（JSON、Markdown、PDF）
6. **协作功能**: 允许用户分享和导入他人的角色卡和世界书

### 架构演进

1. **微服务化**: 将管理器和 bot 分离为独立服务
2. **缓存层**: 添加 Redis 缓存以提高性能
3. **消息队列**: 使用消息队列处理异步任务（摘要、清理）
4. **CDN**: 使用 CDN 存储和分发角色卡头像

