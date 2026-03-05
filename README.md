# imgrab

[English](README_EN.md) | 中文

imgrab 是一个用 Go 开发的 Docker 镜像拉取 CLI 工具，提供镜像搜索、拉取、保存、导入等功能，支持 Docker Hub 和私有仓库。

## 安装

```bash
git clone https://github.com/llklkl/imgrab.git
cd imgrab
go build -o imgrab .
```

## 命令说明

### pull - 拉取镜像

从 Docker Hub 或私有仓库拉取镜像并保存为 tar 文件。

```bash
./imgrab pull [image] [flags]
```

**参数：**
- `image`: 镜像名称，格式为 `[registry/]repository[:tag]`

**Flags：**
- `-o, --output string`: 输出目录
- `-a, --arch string`: 架构（默认：当前架构）
- `-i, --import`: 拉取后导入 Docker

**示例：**
```bash
# 拉取 nginx 最新版
./imgrab pull nginx

# 拉取指定版本
./imgrab pull nginx:1.25.3

# 拉取并保存到指定目录
./imgrab pull nginx -o ./images

# 指定架构
./imgrab pull nginx -a arm64

# 拉取后自动导入 Docker
./imgrab pull nginx -i
```

### search - 搜索镜像（TUI 界面）

使用交互式 TUI 界面搜索 Docker Hub 镜像。

```bash
./imgrab search [query]
```

**参数：**
- `query`: 可选，初始搜索关键词

**示例：**
```bash
# 打开搜索界面
./imgrab search

# 直接搜索 nginx
./imgrab search nginx
```

**TUI 操作：**
- 在搜索框输入关键词，按 `Enter` 开始搜索
- 使用方向键 `↑`/`↓` 选择镜像，按 `Enter` 查看版本
- 在版本列表选择版本，按 `Enter` 确认下载
- 在确认弹窗选择架构（←/→ 切换），按 `y`/`Enter` 确认下载
- 按 `Esc` 返回上一级，按 `q`/`Ctrl+C` 退出

### login - 登录仓库

登录 Docker Hub 或私有仓库。

```bash
./imgrab login [flags]
```

**Flags：**
- `-u, --username string`: 用户名
- `-p, --password string`: 密码
- `-r, --registry string`: 仓库地址（默认：Docker Hub）

**示例：**
```bash
# 登录 Docker Hub
./imgrab login -u your_username -p your_password

# 登录私有仓库
./imgrab login -u your_username -p your_password -r registry.example.com
```

## 配置

imgrab 会在用户目录下保存认证凭据（`~/.docker/config.json`）。
