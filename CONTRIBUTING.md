# 贡献指南

感谢您对本项目的关注！我们欢迎各种形式的贡献。

## 如何贡献

### 报告 Bug

如果您发现了 Bug，请创建一个 Issue，并包含以下信息：

1. Bug 的详细描述
2. 复现步骤
3. 预期行为
4. 实际行为
5. 环境信息（Go 版本、操作系统等）
6. 相关日志或截图

### 提出新功能

如果您有新功能的想法，请创建一个 Issue，并描述：

1. 功能的用途和价值
2. 预期的实现方式
3. 可能的替代方案

### 提交代码

1. **Fork 项目**

2. **创建分支**
```bash
git checkout -b feature/your-feature-name
```

3. **编写代码**
   - 遵循 Go 代码规范
   - 添加必要的注释
   - 编写单元测试
   - 确保所有测试通过

4. **提交代码**
```bash
git add .
git commit -m "feat: add your feature description"
```

提交信息格式：
- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `style:` 代码格式调整
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建或辅助工具的变动

5. **推送到 GitHub**
```bash
git push origin feature/your-feature-name
```

6. **创建 Pull Request**

## 开发环境设置

### 前置要求

- Go 1.21+
- Git
- Make（可选）

### 设置步骤

1. **克隆项目**
```bash
git clone <your-fork-url>
cd go_version
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，填入必要的配置
```

4. **运行测试**
```bash
make test
```

5. **运行程序**
```bash
make run
```

## 代码规范

### Go 代码风格

- 使用 `gofmt` 格式化代码
- 使用 `go vet` 检查代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- 使用有意义的变量和函数名
- 添加必要的注释，特别是导出的函数和类型

### 测试规范

- 为新功能编写单元测试
- 测试覆盖率应保持在 80% 以上
- 使用表驱动测试（table-driven tests）
- 测试文件命名：`*_test.go`

示例：
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case1", "input1", "output1", false},
        {"case2", "input2", "output2", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 文档规范

- 为导出的函数、类型、常量添加文档注释
- 文档注释以类型或函数名开头
- 更新 README.md 和相关文档

示例：
```go
// Config represents the application configuration.
// It contains all settings loaded from environment variables.
type Config struct {
    // Port is the HTTP server port
    Port int
}

// LoadConfig loads configuration from environment variables.
// It returns an error if required variables are missing.
func LoadConfig() (*Config, error) {
    // implementation
}
```

## 项目结构

```
go_version/
├── cmd/bot/           # 程序入口
├── internal/          # 内部包
│   ├── config/        # 配置管理
│   ├── storage/       # 存储层
│   ├── telegram/      # Telegram 集成
│   ├── agent/         # AI Agent
│   └── ...
├── doc/               # 文档
└── ...
```

## Pull Request 检查清单

在提交 PR 之前，请确保：

- [ ] 代码已格式化（`make fmt`）
- [ ] 通过代码检查（`make vet`）
- [ ] 所有测试通过（`make test`）
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 提交信息清晰明确
- [ ] PR 描述详细说明了改动内容

## 获取帮助

如果您有任何问题，可以：

1. 查看现有的 Issues
2. 创建新的 Issue
3. 在 PR 中提问

## 行为准则

- 尊重所有贡献者
- 保持友好和专业
- 接受建设性的批评
- 关注项目的最佳利益

感谢您的贡献！
