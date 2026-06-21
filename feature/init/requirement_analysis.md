# pbrew — AI 组件管理工具需求分析

## 1. 关键词

### 1.1 AI 组件的定义

所谓 AI 组件，是指用户在使用 AI Agent 的过程中所依赖的一些规则、Skill 或 Hook 等能力，我们将其统称为 AI 组件。

### 1.2 AI 组件的原子化

每个 AI 组件是**原子化、自闭合**的——即每个组件拥有一个独立的目录，目录内包含组件的所有文件以及一个 `.pbrew.json` 元数据文件。

`.pbrew.json` 声明当前组件的基本信息：

| 字段 | 说明 |
|------|------|
| `name` | 组件名称，全局唯一标识 |
| `version` | 语义化版本号 |
| `type` | 组件类型：`rule` / `skill` / `hook` |
| `description` | 组件简介 |
| `hash` | 组件内容哈希（用于完整性校验与变更检测） |
| `updatedAt` | 最后修改时间 |

## 2. 项目背景

用户日常在使用 AI Agent 的过程中，可能会需要在不同的设备、甚至不同的项目工程中使用相同的 AI 组件。在这种情况下，用户希望能够方便地同步、安装、更新自己平时顺手的规则、Skill、Hook 等能力。

例如：当用户操作一台新设备时，就希望能够快速将常用的规则、Skill、Hook 等能力迁移到新设备。

## 3. 使用场景

目前市面上有多款主流的 AI Agent，如 CodeBuddy Code、Claude Code 等。不同 Agent 对于同一类组件的文件命名、目录结构、配置字段可能有差异。

因此，pbrew 需要同时支持多种主流 AI Agent。MVP 版本优先支持 **CodeBuddy Code** 和 **Claude Code**。

## 4. 核心设计决策

### 4.1 组件仓库默认使用 Claude Code 格式

组件仓库中的所有 Rule / Skill 文件默认遵循 Claude Code 的格式规范。原因：

- Claude Code 格式更早、社区内容更丰富；
- Claude → CodeBuddy 的转换是机械性替换（文件名、路径前缀、变量名），逻辑清晰、不易出错；
- 反方向（CodeBuddy → Claude）会丢失 `alwaysApply`、`enabled` 等 Claude 不支持的字段。

### 4.2 Agent 适配规则作为 pbrew 内置知识

不同 Agent 针对同一类组件（如 Rule 或 Skill）的差异是**固定的映射关系**，不属于单个组件的个性信息。因此：

- 适配规则**内置在 pbrew 中**，按组件类型（rule / skill / hook）分别定义；
- `.pbrew.json` 中**不包含** Agent 特定字段，只保留组件自身的元数据；
- 安装时 pbrew 根据目标 Agent 查找对应的适配规则，对源文件执行转换后写入目标路径。

### 4.3 安装记录采用"信任用户"原则

pbrew 在安装组件后，会在对应 scope × agent 的目录下放置一个 `.pbrew.json` 记录文件，记录已安装组件的名称、版本、来源等信息。

- 该文件是 `pbrew update` 判断组件是否需要更新的基准；
- pbrew 默认用户不会手动破坏此文件；
- 不对此文件做额外的校验或防护机制。

### 4.4 单一数据源

用户绑定一个数据源（Git 仓库或本地目录）作为组件来源。`pbrew list`、`install`、`update` 均基于该数据源操作。

## 5. 基本使用

### 5.1 数据源配置（pbrew config）

用户首次使用 pbrew 时需设置数据源。配置方式参考 `git config`：

```bash
# 设置数据源
pbrew config set source <git-repo-url | local-dir-path>

# 查看配置
pbrew config get source

# 列出所有配置
pbrew config list
```

配置文件存储在 `~/.pbrew/config`（INI 格式）。

执行 `list` / `install` / `update` 前会检测 `source` 是否已设置，未设置则提示用户先配置。

### 5.2 AI 组件的查询（pbrew list）

```bash
pbrew list
```

列出已绑定数据源下的所有可用 AI 组件，显示组件名称、类型、版本、简介等信息。

### 5.3 AI 组件的安装（pbrew install）

```bash
pbrew install <component-name...> --agent <agent-id> --scope <scope>
```

- `<component-name...>`：一个或多个组件名称
- `--agent`：目标 Agent（如 `claude`、`codebuddy`）
- `--scope`：安装范围，支持以下值：
  - `user`：用户级（如 `~/.claude/` 或 `~/.codebuddy/`）
  - `project`：项目级（如 `./.claude/` 或 `./.codebuddy/`）

安装流程：

1. 从数据源读取组件目录及 `.pbrew.json`
2. 根据目标 Agent 查找内置适配规则，对组件文件执行转换
3. 将转换后的文件写入目标 Agent 对应的安装路径
4. 在该 scope × agent 目录下更新 `.pbrew.json` 安装记录

### 5.4 AI 组件的更新（pbrew update）

```bash
# 更新指定组件
pbrew update <component-name...> --agent <agent-id> --scope <scope>

# 更新指定范围内所有已安装组件
pbrew update --all --agent <agent-id> --scope <scope>
```

更新流程：

1. 读取目标 scope × agent 目录下的 `.pbrew.json` 安装记录
2. 对比数据源中组件的版本/哈希与已安装版本
3. 如有新版本，重新执行安装流程覆盖旧文件并更新记录

### 5.5 AI 组件的卸载（pbrew uninstall）

```bash
pbrew uninstall <component-name...> --agent <agent-id> --scope <scope>
```

卸载流程：

1. 读取目标 scope × agent 目录下的 `.pbrew.json` 安装记录
2. 删除该组件对应的已安装文件
3. 从 `.pbrew.json` 安装记录中移除该组件条目

### 5.6 AI 组件的发布（pbrew publish）

```bash
pbrew publish --name <component-name> --dir <component-dir>
```

- `--name`：期望发布的组件名称
- `--dir`：组件所在目录

发布流程：

1. 读取组件目录，根据组件类型和内置适配规则，将组件格式转换为仓库默认格式（Claude Code）
2. 自动生成版本号
3. 生成 `.pbrew.json` 元数据文件
4. 将组件写入已绑定的数据源目录
5. 输出后续 git 提交命令供用户手动执行

## 6. MVP 范围

### 6.1 命令

| 命令 | 优先级 |
|------|--------|
| `pbrew config set/get/list` | P0 |
| `pbrew list` | P0 |
| `pbrew install` | P0 |
| `pbrew update` | P0 |
| `pbrew uninstall` | P0 |
| `pbrew publish` | P1（MVP 之后） |

### 6.2 数据源

MVP 支持以下数据源类型：

- Git 仓库（通过 URL 指定）
- 本地目录

### 6.3 Agent 支持

MVP 支持以下 Agent：

- Claude Code
- CodeBuddy Code

### 6.4 组件类型

MVP 支持以下组件类型：

- Rule（规则）
- Skill

## 7. 技术要求

- 项目需集成端到端（E2E）测试能力，确保各命令功能可及时验证；
- 后续迭代中通过 E2E 测试保障功能稳定性。
