# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 概述

**govm** 是一个用 Go 编写的 Go 版本管理器。它允许用户安装、管理和切换不同的 Go 版本。该工具从官方 golang.org 镜像下载 Go 发行版，使用 SHA256 验证，并通过符号链接进行管理。

## 高层架构

### 应用程序结构

```
govm/
├── cmd/                    # 命令实现
│   ├── cmd.go             # 主 CLI 设置和命令路由
│   ├── govm.go            # 自我升级功能
│   ├── install.go         # 安装 Go 版本
│   ├── use.go             # 切换活动 Go 版本
│   ├── list.go            # 列出可用/已安装版本
│   ├── uninstall.go       # 移除已安装版本
│   ├── cache.go           # 缓存管理
│   ├── exec.go            # 使用特定 Go 版本执行命令
│   ├── update.go          # 从 golang.org 更新版本列表
│   ├── upgrade.go         # 升级 govm 自身
│   ├── hold.go            # 保持/取消保持版本
│   └── types.go           # 类型定义
├── cmds/govm/             # 主入口点
│   └── main.go            # 调用 cmd.Run()
├── config/                # 配置管理
│   └── config.go          # 配置结构体和 YAML 处理
├── pkg/                   # 共享包
│   ├── version/           # 版本解析和比较
│   ├── utils/             # 工具函数
│   └── ...                # 其他包
├── types/                 # 类型定义
├── .golangci.yml          # Linter 配置
├── .goreleaser.yaml       # 发布配置
├── go.mod                 # Go 模块依赖
├── Makefile               # 构建和开发命令
└── README.md              # 项目文档
```

### 关键组件

1. **CLI 框架**: 使用 `urfave/cli/v3` 作为命令行界面
2. **配置**: 基于 YAML 的配置，位于 `~/.govm/conf.yaml`
3. **版本管理**: 自定义版本解析器，支持带 beta/rc 后缀的语义化版本
4. **下载机制**: 从 `https://go.dev/dl/` 下载并进行 SHA256 验证
5. **符号链接管理**: 使用 `~/.govm/go` 符号链接管理活动版本
6. **GitHub 集成**: 通过 GitHub API 检查更新和自我升级

### 数据位置

- **配置**: `~/.govm/conf.yaml`
- **已安装版本**: `~/.govm/.install/`
- **缓存**: `~/.govm/.cache/`
- **活动版本符号链接**: `~/.govm/go`

## 构建和开发命令

### 构建
```bash
# 构建二进制文件
make build

# 或直接使用 go
go build -ldflags "-s -w -X github.com/serious-snow/govm/cmd.Version=dev" -trimpath -o ./ ./cmds/govm/
```

### Lint 和格式化
```bash
# 运行 linter (需要 mise 或 golangci-lint)
make lint

# 格式化代码
make fmt
```

### 测试
```bash
# 运行测试 (如果有)
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...
```

### 开发
```bash
# 安装用于开发
go install ./cmds/govm/

# 直接运行
go run ./cmds/govm/ [command] [args]
```

## 核心命令

CLI 提供以下主要命令：

| 命令 | 描述 |
|---------|-------------|
| `install` | 下载并安装 Go 版本 |
| `use` | 切换到特定 Go 版本 (创建符号链接) |
| `list` | 列出可用或已安装版本 |
| `uninstall` | 移除已安装的 Go 版本 |
| `unuse` | 停用当前使用的 Go 版本 |
| `cache` | 管理缓存 (目录, 清除, 大小) |
| `exec` | 使用特定 Go 版本运行命令 |
| `update` | 从 golang.org 更新可用版本列表 |
| `upgrade` | 从 GitHub 发布升级 govm 自身 |
| `hold` | 保持版本 (防止卸载) |
| `unhold` | 取消保持版本 |

## 版本管理

### 版本格式
版本遵循语义化版本控制，可选 beta/rc 后缀：
- `1.21.0` - 标准发布
- `1.21.0-rc2` - 候选版本
- `1.21.0-beta1` - Beta 版本

### 版本比较
`pkg/version/version.go` 中的自定义版本解析器处理：
- 空版本 (视为最高)
- RC 版本 (视为低于稳定版)
- Beta 版本 (视为低于 RC)
- 标准 semver 比较 (major.minor.patch)

### 安装过程
1. 检查版本是否已安装
2. 从 `https://go.dev/dl/` 下载
3. 验证 SHA256 校验和
4. 解压到 `~/.govm/.install/go<version>/`
5. 在 `~/.govm/go` 创建指向活动版本的符号链接

## 配置

### 配置文件位置
`~/.govm/conf.yaml`

### 配置结构
```yaml
cachePath: ~/.govm/.cache
installPath: ~/.govm/.install
autoSetEnv: true
```

### 环境变量
- `GOVM_CACHE_PATH`: 覆盖缓存目录
- `GOVM_INSTALL_PATH`: 覆盖安装目录

## 已知问题

目前代码库中没有已知的严重问题。最近的修复包括：

1. **资产选择逻辑** (commit 9f0b7e5): 修复了在循环中使用 `asset.GetName()` 而不是 `v.GetName()` 的问题
2. **可执行文件路径获取** (commit 546893b): 改进了升级过程中获取真实可执行文件路径的方法，添加了符号链接解析功能

## 开发工作流

### 添加新命令
1. 在 `cmd/` 目录中添加命令实现
2. 在 `cmd/cmd.go` 的 `Run()` 函数中注册命令
3. 在 `types/` 中更新类型 (如果需要)
4. 在相应的 `*_test.go` 文件中添加测试

### 版本解析更改
版本解析器在 `pkg/version/version.go` 中。此处的更改会影响：
- 版本比较逻辑
- 排序顺序
- 与可用版本的匹配

### 下载机制
下载逻辑使用：
- `pkg/utils/httpc` 进行 HTTP 请求
- 来自 golang.org 的 SHA256 验证
- .tar.gz 文件的解压缩

### 自我升级过程
`cmd/govm.go` 中的 `upgradeGOVM()` 函数：
1. 检查 GitHub 发布 API
2. 查找匹配当前操作系统/架构的资产
3. 下载到临时目录
4. 解压并替换当前可执行文件
5. 使用原子重命名确保安全

## 测试策略

### 单元测试
- 版本解析和比较
- 配置加载
- SHA256 验证

### 集成测试
- 安装/使用/卸载工作流
- 缓存管理
- 自我升级过程

### 手动测试
```bash
# 测试安装
go run ./cmds/govm/ install 1.21.0

# 测试使用
go run ./cmds/govm/ use 1.21.0

# 测试列表
go run ./cmds/govm/ list --installed

# 测试执行
go run ./cmds/govm/ exec 1.21.0 go version
```

## 依赖项

`go.mod` 中的关键依赖：
- `github.com/urfave/cli/v3` - CLI 框架
- `github.com/google/go-github/v66` - GitHub API
- `github.com/briandowns/spinner` - 进度指示器
- `github.com/fatih/color` - 终端颜色
- `github.com/manifoldco/promptui` - 交互式提示
- `gopkg.in/yaml.v3` - YAML 解析

## 发布过程

项目使用 GoReleaser (`.goreleaser.yaml`) 进行发布。发布过程：
1. 为多个平台构建二进制文件
2. 创建 GitHub 发布
3. 上传带操作系统/架构命名的资产

## 常见开发任务

### 修复构建错误
```bash
# 检查类型错误
go build ./...

# 运行 linter 检查代码质量问题
make lint
```

### 更新依赖
```bash
# 添加新依赖
go get github.com/example/package

# 更新所有依赖
go get -u ./...
go mod tidy
```

### 调试版本问题
```bash
# 启用调试日志 (如果可用)
GOVM_DEBUG=1 go run ./cmds/govm/ [command]

# 检查版本解析
go run ./cmds/govm/ list --available
```

## 性能考虑

- **缓存**: SHA256 验证被缓存以避免重新下载
- **并行下载**: 当前未实现
- **内存**: 大文件下载使用流式处理
- **符号链接**: 通过符号链接更新实现快速版本切换

## 安全考虑

- **SHA256 验证**: 所有下载都与官方校验和进行验证
- **仅 HTTPS**: 所有下载使用 HTTPS
- **无任意代码执行**: 仅运行 Go 工具链命令
- **文件权限**: 适当设置可执行权限

## 未来改进

潜在增强领域：
1. 并行下载支持
2. 版本别名 (例如 "latest", "stable")
3. 从 go.mod 自动检测版本
4. 与 go.work 文件集成
5. 性能分析和优化
