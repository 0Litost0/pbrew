# pbrew

**pbrew** 是一个 AI 组件管理工具，用于在不同 AI Agent 之间同步、安装、更新 Rule 和 Skill 等组件。

## 项目概述

日常使用 AI Agent（如 Claude Code、CodeBuddy Code）时，用户在不同设备、不同项目间需要重复配置相同的规则与技能。pbrew 将这些能力抽象为**原子化的 AI 组件**，通过统一的数据源实现一键安装与跨 Agent 适配：

- **Rule**（规则）— 如 Git 提交规范、API 开发约束
- **Skill**（技能）— 如自动生成 commit message、代码审查流程

核心特性：

- **跨 Agent 适配**：内置 Claude Code / CodeBuddy Code 的格式差异映射，安装时自动转换目录结构与变量名
- **Canonical 中转架构**：新增 Agent 只需编写一对 Adapter，无需与其他 Agent 两两适配
- **数据驱动路径**：所有安装路径由 `AgentProfile` 声明式驱动，零 if/else 分支逻辑
- **原子写入**：全局采用 write-temp-then-rename 策略，崩溃安全

### MVP 范围

| 项目 | 支持 |
|------|------|
| 命令 | `config`、`list`、`install`、`update`、`uninstall` |
| 数据源 | Git 仓库（浅克隆）、本地目录 |
| Agent | Claude Code、CodeBuddy Code |
| 组件类型 | Rule、Skill |

## 快速上手

### 安装

确保 `$GOPATH/bin`（默认 `~/go/bin`）已在 `$PATH` 中，然后：

```bash
go install github.com/user/pbrew@latest
```

即可直接使用 `pbrew` 命令。

> 也可从源码构建，详见下方[本地构建](#本地构建)。

### 配置数据源

首次使用需绑定一个组件数据源（Git 仓库或本地目录）：

```bash
# Git 仓库
pbrew config set source https://github.com/your-org/ai-components.git

# 本地目录
pbrew config set source /path/to/local/repo

# 查看配置
pbrew config list
```

### 查询组件

```bash
# 列出数据源中所有可用组件
pbrew list

# 列出指定 Agent 下已安装的组件及版本状态
pbrew list --agent claude --scope user
```

### 安装组件

```bash
# 安装 Rule 到 Claude Code（项目级，默认）
pbrew install git-conventions --agent claude

# 安装 Skill 到 CodeBuddy Code（用户级）
pbrew install commit --agent codebuddy --scope user

# 跳过交互确认（CI/CD 场景）
pbrew install git-conventions --agent claude --force
```

### 更新组件

```bash
# 更新指定组件
pbrew update git-conventions --agent claude --scope user

# 更新全部已安装组件
pbrew update --all --agent claude --scope user
```

### 卸载组件

```bash
pbrew uninstall git-conventions --agent claude --scope user
```

## 本地构建

### 环境依赖

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| Go | ≥ 1.23 | 编译与运行 |
| Git | 任意稳定版 | 仅使用 Git 数据源时需要 |

> pbrew 编译为静态二进制，运行时不需要 Go 环境。

### 安装与构建

```bash
git clone git@github.com:0Litost0/pbrew.git
cd pbrew

# 编译并安装到 $GOPATH/bin（推荐，无需手动 cp）
go install .

# 或者仅在当前目录生成二进制（不安装）
go build -o pbrew .

# 跨平台发布构建
GOOS=linux   GOARCH=amd64 go build -o dist/pbrew-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o dist/pbrew-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o dist/pbrew-windows-amd64.exe .
```

> **提示**：`go install .` 会编译并将二进制放入 `$GOPATH/bin`（默认 `~/go/bin`），确保该目录在 `$PATH` 中即可直接使用 `pbrew`。

### 运行测试

```bash
# 单元测试
go test ./internal/...

# E2E 测试（testscript）
go test ./e2e/ -v -timeout 120s

# 全部测试
go test ./...
```

### Go 依赖

项目遵循最小化依赖原则，直接依赖仅 5 个：

| 依赖 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `gopkg.in/yaml.v3` | YAML frontmatter 解析 |
| `github.com/Masterminds/semver/v3` | 语义化版本比较 |
| `github.com/rogpeppe/go-internal` | E2E 测试（testscript） |
| `github.com/stretchr/testify` | 测试断言 |

## 目录结构

```
pbrew/
├── main.go                       # 入口：初始化 cobra 根命令并执行
├── go.mod / go.sum               # 依赖管理
├── cmd/                          # CLI 命令定义（薄层，只做参数解析和编排）
│   ├── root.go                   # 根命令、全局 flag（--verbose）
│   ├── config.go                 # pbrew config set / get / list
│   ├── list.go                   # pbrew list
│   ├── install.go                # pbrew install
│   ├── update.go                 # pbrew update
│   ├── uninstall.go              # pbrew uninstall
│   └── helpers.go                # 命令层共享工具函数
├── internal/                     # 核心业务逻辑（不可被外部导入）
│   ├── model/                    # 纯数据结构：Component、InstallRecord、Config、AgentProfile
│   ├── agent/                    # Agent 注册表：内置 claude / codebuddy Profile + 查询接口
│   ├── config/                   # 配置管理：读写 ~/.pbrew/config.json
│   ├── source/                   # 数据源抽象：Source 接口 + LocalSource / GitSource 实现
│   ├── registry/                 # 组件发现：遍历目录 + _index.json 索引加速
│   ├── adapter/                  # 适配引擎：Canonical 中转格式 + 双向 Adapter + 路径替换
│   ├── installer/                # 编排层：Install / Uninstall / Update 全流程 + 原子写入
│   └── platform/                 # 平台工具：项目根目录查找、Agent 目标路径计算
├── testdata/                     # 共享测试数据
│   └── fixtures/
│       └── repo/                 # 组件仓库示例（含 _index.json、Rule、Skill）
├── e2e/                          # E2E 测试（testscript）
│   ├── main_test.go              # 编译二进制 + 运行脚本 + 自定义命令
│   └── testdata/script/          # .txt 测试脚本
└── feature/                      # 需求文档
    └── init/                     # 需求分析、设计文档、任务清单
```
