# imgrab 需求规范与技术选型

## 1. 项目概述

imgrab 是一个用 Go 1.25.7 开发的 Docker 镜像拉取 CLI 工具，提供镜像搜索、拉取、保存、导入等功能，支持 Docker Hub 和私有仓库。

## 2. 功能需求分析

### 2.1 镜像拉取功能

| 功能点 | 描述 | 优先级 |
|--------|------|--------|
| 多仓库支持 | 从 Docker Hub 或私有仓库拉取镜像 | P0 |
| 架构选择 | 默认与当前环境架构一致，支持手动指定 | P0 |
| 版本指定 | 使用 `@` 格式指定镜像版本 | P0 |
| 保存到指定目录 | 允许用户指定 tar 文件保存位置 | P0 |
| 自动导入 Docker | 拉取后通过 docker 命令导入（需手动指定） | P1 |
| 进度条展示 | 拉取过程中显示进度条 | P0 |
| 分层并行拉取 | 支持镜像分层并行下载 | P1 |
| 默认行为 | 默认只下载不导入 | P0 |

### 2.2 认证功能

| 功能点 | 描述 | 优先级 |
|--------|------|--------|
| 登录认证 | 支持用户名/密码登录 Docker Hub 或私有仓库 | P0 |

### 2.3 搜索功能（TUI）

| 功能点 | 描述 | 优先级 |
|--------|------|--------|
| 镜像搜索 | 搜索并展示镜像列表 | P0 |
| 版本列表 | 按版本列出镜像 | P0 |
| 选择下载 | 选中版本后弹窗确认下载，可选择架构 | P0 |

## 3. 技术选型

### 3.1 核心依赖库

| 库名 | 用途 | 选型理由 |
|------|------|----------|
| `github.com/google/go-containerregistry` | 镜像拉取、操作 | Google 官方库，功能完善，支持认证、分层拉取等 |
| `github.com/charmbracelet/bubbletea` | TUI 框架 | 现代化 TUI 库，生态完善，适合构建交互式界面 |
| `github.com/charmbracelet/bubbles` | TUI 组件 | 提供进度条、列表等常用组件 |
| `github.com/charmbracelet/lipgloss` | TUI 样式 | 强大的终端样式库 |
| `github.com/spf13/cobra` | CLI 框架 | 行业标准的 Go CLI 框架 |
| `github.com/schollz/progressbar/v3` | 进度条 | 简单易用的进度条库 |

### 3.2 技术架构

```
imgrab/
├── cmd/                    # CLI 命令入口
│   ├── root.go            # 根命令
│   ├── pull.go            # 拉取命令
│   ├── search.go          # 搜索命令
│   └── login.go           # 登录命令
├── internal/
│   ├── registry/          # 镜像仓库交互
│   │   ├── client.go      # 仓库客户端
│   │   ├── auth.go        # 认证相关
│   │   ├── pull.go        # 拉取逻辑
│   │   └── search.go      # 搜索逻辑
│   ├── tui/               # TUI 界面
│   │   ├── app.go         # 主应用
│   │   ├── search.go      # 搜索界面
│   │   ├── confirm.go     # 确认弹窗
│   │   └── progress.go    # 进度展示
│   └── docker/            # Docker 相关操作
│       └── import.go      # 镜像导入
├── go.mod
└── go.sum
```

### 3.3 go-containerregistry 能力验证

该库可以满足需求：
- ✅ 支持 Docker Hub 和私有仓库
- ✅ 支持认证（username/password）
- ✅ 支持架构选择
- ✅ 支持分层拉取
- ✅ 支持保存为 tar 文件
- ✅ 可获取下载进度信息

## 4. 非功能性需求

- **性能**: 镜像分层并行拉取，提升下载速度
- **用户体验**: TUI 界面友好，进度条清晰
- **可维护性**: 代码结构清晰，模块化设计
