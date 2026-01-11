# 实现计划: SillyTavern 集成

## 概述

本实现计划将 SillyTavern 核心功能集成到现有 Telegram bot 中。根据代码库检查，大部分实现代码已存在但未提交，关键的数据库模型定义缺失。

## 当前状态分析

### ✅ 已实现（在 staged 区域）
- SillyTavern 管理器代码（character, worldbook, preset, regex）
- 请求构建器和上下文管理器
- Web 管理器 API 和认证
- Telegram 命令（login, share, clear, clear_all）
- Telegraph 集成
- 完整的测试套件

### ❌ 缺失的关键组件
- **数据库模型定义**：CharacterCard, WorldBook, WorldBookEntry, Preset, RegexPattern, LoginToken
- **Storage 接口扩展**：SillyTavern 相关的 CRUD 方法
- **GORM 实现**：数据库操作的具体实现
- **数据库迁移**：AutoMigrate 中缺少 SillyTavern 表

## 任务列表

### 第一阶段：修复数据库层（关键）

- [x] 1. 添加 SillyTavern 数据库模型到 models.go
  - 在 `internal/storage/models.go` 中添加缺失的模型定义
  - CharacterCard - 角色卡模型
  - WorldBook - 世界书模型  
  - WorldBookEntry - 世界书条目模型
  - Preset - 预设模型
  - RegexPattern - 正则模式模型
  - LoginToken - 登录令牌模型
  - 为每个模型添加 GORM 标签、索引和 TableName 方法
  - _需求: 1.1, 1.2, 1.3, 1.4, 4.2_

- [x] 2. 扩展 Storage 接口
  - 在 `internal/storage/types.go` 中添加 SillyTavern 相关方法
  - 角色卡 CRUD：CreateCharacterCard, GetCharacterCard, ListCharacterCards, UpdateCharacterCard, DeleteCharacterCard, GetActiveCharacterCard, ActivateCharacterCard
  - 世界书 CRUD：CreateWorldBook, GetWorldBook, ListWorldBooks, UpdateWorldBook, DeleteWorldBook, GetActiveWorldBook, ActivateWorldBook
  - 世界书条目 CRUD：CreateWorldBookEntry, GetWorldBookEntry, ListWorldBookEntries, UpdateWorldBookEntry, DeleteWorldBookEntry
  - 预设 CRUD：CreatePreset, GetPreset, ListPresets, UpdatePreset, DeletePreset, GetActivePreset, ActivatePreset
  - 正则 CRUD：CreateRegexPattern, GetRegexPattern, ListRegexPatterns, UpdateRegexPattern, DeleteRegexPattern
  - 登录令牌：CreateLoginToken, ValidateLoginToken, DeleteLoginToken, CleanupExpiredTokens
  - _需求: 1.1, 1.2, 1.3, 1.4, 4.2, 4.4_

- [x] 3. 实现 GORM Storage 方法
  - 在 `internal/storage/gorm.go` 中实现所有新接口方法
  - 使用 GORM 的参数化查询防止 SQL 注入
  - 实现事务支持（特别是激活操作）
  - 添加适当的错误处理
  - _需求: 1.1, 1.2, 1.3, 1.4, 4.2, 4.4_

- [x] 4. 更新数据库迁移
  - 在 `internal/storage/gorm.go` 的 NewStorage 函数中
  - 将所有 SillyTavern 模型添加到 AutoMigrate 调用
  - 确保表结构正确创建
  - _需求: 1.1, 1.2, 1.3, 1.4_

- [x] 5. 修改 HistoryItem 结构
  - 在 `internal/storage/types.go` 中为 HistoryItem 添加字段
  - 添加 Timestamp int64 字段（Unix 时间戳）
  - 添加 Truncated bool 字段（标记是否被 /clear 截断）
  - 更新相关的序列化/反序列化逻辑
  - _需求: 6.1, 6.4_

### 第二阶段：验证和测试

- [x] 6. 运行测试验证实现
  - 运行 `go test ./internal/storage/...` 验证存储层
  - 运行 `go test ./internal/sillytavern/...` 验证管理器
  - 运行 `go test ./internal/manager/...` 验证 Web API
  - 运行 `go test ./internal/integration/...` 验证端到端流程
  - 修复任何测试失败
  - _需求: 所有_

- [x] 7. 验证主程序集成
  - 检查 `cmd/bot/main.go` 中的初始化代码
  - 确保所有 SillyTavern 组件正确初始化
  - 确保 HTTP 服务器正确启动
  - 验证命令注册
  - _需求: 1.5, 3.1, 4.1_

- [x] 8. 构建和运行检查
  - 运行 `go build ./...` 确保无编译错误
  - 运行 `go mod tidy` 清理依赖
  - 测试基本功能（如果可能）
  - _需求: 所有_

### 第三阶段：配置和文档

- [x] 9. 验证配置系统
  - 检查 `internal/config/config.go` 中的新配置字段
  - 确保所有环境变量正确定义：
    - MAX_CONTEXT_LENGTH
    - SUMMARY_THRESHOLD  
    - MIN_RECENT_PAIRS
    - MANAGER_PORT
    - MANAGER_ENABLED
    - TELEGRAPH_ENABLED
  - _需求: 2.3, 2.4_

- [x] 10. 更新文档
  - 验证 `doc/MIGRATION.md` 的完整性
  - 验证 `README.md` 中的 SillyTavern 说明
  - 验证 `doc/CONFIG.md` 中的配置说明
  - _需求: 所有_

### 第四阶段：Git 清理和提交

- [-] 11. Git 状态清理
  - 检查所有 staged 文件
  - 确保所有必要的文件都已添加
  - 提交所有更改
  - 推送到远程仓库

## 详细实现指南

### 任务 1: 数据库模型定义

在 `internal/storage/models.go` 文件末尾添加以下模型：

```go
// CharacterCard represents a SillyTavern character card
type CharacterCard struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Owner information
	UserID *int64 `gorm:"index"` // nil for global cards
	
	// Card metadata
	Name   string `gorm:"not null;index"`
	Avatar string `gorm:"type:text"` // Avatar URL or base64
	
	// SillyTavern V2 format data
	Data string `gorm:"type:text;not null"` // JSON format
	
	// Status
	IsActive bool `gorm:"default:false;index"`
}

func (CharacterCard) TableName() string {
	return "character_cards"
}

// WorldBook represents a SillyTavern world book
type WorldBook struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Owner information
	UserID *int64 `gorm:"index"` // nil for global books
	
	// Book metadata
	Name string `gorm:"not null;index"`
	
	// World book data
	Data string `gorm:"type:text;not null"` // JSON format
	
	// Status
	IsActive bool `gorm:"default:false;index"`
}

func (WorldBook) TableName() string {
	return "world_books"
}

// WorldBookEntry represents an entry in a world book
type WorldBookEntry struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Association
	WorldBookID uint `gorm:"not null;index"`
	
	// Entry data
	UID           string `gorm:"not null;uniqueIndex"`
	Keys          string `gorm:"type:text;not null"` // JSON array
	SecondaryKeys string `gorm:"type:text"`          // JSON array
	Content       string `gorm:"type:text;not null"`
	Comment       string `gorm:"type:text"`
	
	// Configuration
	Constant  bool   `gorm:"default:false"`
	Selective bool   `gorm:"default:false"`
	Order     int    `gorm:"default:100"`
	Position  string `gorm:"default:'after_char'"` // before_char, after_char
	Enabled   bool   `gorm:"default:true;index"`
	
	// Extensions
	Extensions string `gorm:"type:text"` // JSON format
}

func (WorldBookEntry) TableName() string {
	return "world_book_entries"
}

// Preset represents a SillyTavern preset
type Preset struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Owner information
	UserID *int64 `gorm:"index"` // nil for global presets
	
	// Preset metadata
	Name    string `gorm:"not null;index"`
	APIType string `gorm:"not null;index"` // openai, anthropic, etc.
	
	// Preset data
	Data string `gorm:"type:text;not null"` // JSON format
	
	// Status
	IsActive bool `gorm:"default:false;index"`
}

func (Preset) TableName() string {
	return "presets"
}

// RegexPattern represents a regex transformation pattern
type RegexPattern struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Owner information
	UserID *int64 `gorm:"index"` // nil for global patterns
	
	// Pattern metadata
	Name string `gorm:"not null;index"`
	
	// Pattern configuration
	Pattern string `gorm:"type:text;not null"` // Regex pattern
	Replace string `gorm:"type:text;not null"` // Replacement text
	Type    string `gorm:"not null"`           // input, output
	Order   int    `gorm:"default:100"`
	Enabled bool   `gorm:"default:true;index"`
}

func (RegexPattern) TableName() string {
	return "regex_patterns"
}

// LoginToken represents a temporary login token for web manager
type LoginToken struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	
	// User information
	UserID int64 `gorm:"not null;uniqueIndex"`
	
	// Token
	Token string `gorm:"not null;uniqueIndex"`
	
	// Expiration
	ExpiresAt time.Time `gorm:"not null;index"`
}

func (LoginToken) TableName() string {
	return "login_tokens"
}
```

### 任务 2 & 3: Storage 接口和实现

需要在 `types.go` 中添加接口方法，在 `gorm.go` 中添加实现。由于篇幅限制，请参考设计文档中的详细接口定义。

### 任务 4: 更新 AutoMigrate

在 `internal/storage/gorm.go` 的 `NewStorage` 函数中，更新 AutoMigrate 调用：

```go
if err := db.AutoMigrate(
	&ChatHistory{},
	&UserConfiguration{},
	&MessageIDs{},
	&GroupAdmins{},
	&CharacterCard{},
	&WorldBook{},
	&WorldBookEntry{},
	&Preset{},
	&RegexPattern{},
	&LoginToken{},
); err != nil {
	return nil, fmt.Errorf("failed to migrate database schema: %w", err)
}
```

## 优先级

**P0 - 阻塞性问题（必须立即修复）：**
- 任务 1: 添加数据库模型定义
- 任务 2: 扩展 Storage 接口
- 任务 3: 实现 GORM Storage 方法
- 任务 4: 更新数据库迁移

**P1 - 高优先级：**
- 任务 5: 修改 HistoryItem 结构
- 任务 6: 运行测试验证

**P2 - 中优先级：**
- 任务 7: 验证主程序集成
- 任务 8: 构建和运行检查

**P3 - 低优先级：**
- 任务 9: 验证配置系统
- 任务 10: 更新文档
- 任务 11: Git 清理

## 注意事项

1. **Git 状态**：所有 SillyTavern 实现文件都在 staged 区域，但数据库模型定义缺失
2. **依赖关系**：任务 1-4 必须按顺序完成，它们是其他所有功能的基础
3. **测试驱动**：现有测试代码已经引用了这些模型，修复后测试应该能通过
4. **向后兼容**：确保不破坏现有的 ChatHistory 等功能

## 估计工作量

- 任务 1: 30 分钟（复制粘贴模型定义）
- 任务 2: 1 小时（定义接口方法）
- 任务 3: 2-3 小时（实现所有 CRUD 操作）
- 任务 4: 10 分钟（更新 AutoMigrate）
- 任务 5: 20 分钟（修改 HistoryItem）
- 任务 6-11: 1-2 小时（测试和验证）

**总计：约 5-7 小时**
