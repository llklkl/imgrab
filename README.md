# imgrab

[English](README_EN.md) | 中文

imgrab 是一个用 Go 开发的 Docker 镜像拉取 CLI 工具，提供镜像搜索、拉取、保存、导入等功能，支持 Docker Hub。

## 特性

- 搜索 Docker Hub 镜像（TUI 界面）
- 默认自动导入到 Docker
- 支持仅下载模式（保存 tar 文件）
- 支持多架构选择

## 安装

```bash
git clone https://github.com/llklkl/imgrab.git
cd imgrab
go build -o imgrab .
```

## 命令说明

### pull - 拉取镜像

从 Docker Hub 或私有仓库拉取镜像。**默认自动导入 Docker**，无需手动操作。

```bash
./imgrab pull [image] [flags]
```

**参数：**
- `image`: 镜像名称，格式为 `[registry/]repository[:tag]`

**Flags：**
- `-d, --download-only`: 仅下载，不导入 Docker
- `-o, --output string`: 输出目录（仅与 `--download-only` 一起使用）
- `-a, --arch string`: 架构（默认：当前架构）

**示例：**
```bash
# 拉取并自动导入 Docker（默认行为）
./imgrab pull nginx

# 拉取指定版本并导入
./imgrab pull nginx:1.25.3

# 仅下载，不导入
./imgrab pull nginx --download-only

# 仅下载到指定目录
./imgrab pull nginx -d -o ./images

# 指定架构
./imgrab pull nginx -a arm64
```

**说明：**
- 默认模式下，镜像会下载到临时目录并自动导入 Docker，完成后自动清理临时文件
- 使用 `--download-only` 可保留 tar 文件到当前目录或指定目录

### search - 搜索镜像（TUI 界面）

使用交互式 TUI 界面搜索 Docker Hub 镜像，支持可视化选择镜像、标签、架构，并选择下载或导入操作。

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

**TUI 操作流程：**

1. **搜索镜像**
   - 在搜索框输入关键词，按 `Enter` 开始搜索
   - 使用 `↑`/`↓` 选择镜像，按 `Enter` 查看版本列表

2. **选择版本**
   - 使用 `↑`/`↓` 选择版本标签
   - 按 `Enter` 查看可用架构

3. **选择架构**
   - 使用 `↑`/`↓` 选择架构
   - 按 `Enter` 进入确认界面

4. **确认操作**
   - 使用 `←`/`→` 切换操作模式：
     - `Download Only` - 仅下载镜像文件
     - `Import to Docker` - 下载并导入到 Docker（默认）
   - 按 `y` 或 `Enter` 确认并开始下载
   - 下载/导入完成后自动退出

5. **进度显示**
   - 实时显示下载进度条、速度和剩余时间
   - 导入完成后自动退出

**快捷键：**
- `Esc` - 返回上一级
- `q` / `Ctrl+C` - 退出应用
