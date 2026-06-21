# Agent AI 组件格式差异对比：Claude Code vs CodeBuddy Code

本文档梳理 Claude Code 与 CodeBuddy Code 在 Rule 和 Skill 两种 AI 组件上的格式差异，作为 PBrew 适配规则的设计依据。

---

## 1. Rule（规则）组件差异

### 1.1 文件命名

| 维度 | Claude Code | CodeBuddy Code |
|------|-------------|----------------|
| 主文件名 | `CLAUDE.md` | `CODEBUDDY.md` |
| 本地私有文件名 | `CLAUDE.local.md` | `CODEBUDDY.local.md` |

### 1.2 目录结构

| 维度 | Claude Code | CodeBuddy Code |
|------|-------------|----------------|
| 项目级配置目录 | `.claude/` | `.codebuddy/` |
| 规则文件目录 | `.claude/rules/*.md` | `.codebuddy/rules/*.md` |
| 用户级目录 | `~/.claude/` | `~/.codebuddy/` |
| 用户级规则目录 | `~/.claude/rules/*.md` | `~/.codebuddy/rules/*.md` |

### 1.3 条件加载（Frontmatter）

两者均使用 YAML frontmatter 控制规则的加载行为，但字段有差异：

**Claude Code：**

```yaml
---
paths: src/api/**/*.ts
---
```

仅支持 `paths` 字段。规则默认始终加载；有 `paths` 时变为条件触发。

**CodeBuddy Code：**

```yaml
---
description: API 开发规则
alwaysApply: true
enabled: true
paths: src/api/**/*.ts
updatedAt: 2026-01-13T12:00:00Z
provider: some-plugin
---
```

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `description` | string | — | 规则简介，IDE 用于代理请求匹配 |
| `alwaysApply` | boolean | `true` | 是否始终注入上下文 |
| `enabled` | boolean | `true` | `false` 时规则完全不加载 |
| `paths` | string/string[] | — | 触发规则的 glob 模式 |
| `updatedAt` | string | — | 时间戳，IDE 编辑时自动管理 |
| `provider` | string | — | 来源标识符 |

**规则类型判定逻辑对比：**

| alwaysApply | paths | Claude Code 行为 | CodeBuddy Code 行为 |
|-------------|-------|-------------------|---------------------|
| true（默认）/ 无 frontmatter | 任意 | 始终加载 | 始终加载 |
| false | 有值 | 条件触发 | 条件触发（仅 CLI）|
| false | 无 | 条件触发 | **不加载** |

> 注意：CodeBuddy Code 在 `alwaysApply: false` 且无 `paths` 时规则不会加载，而 Claude Code 仍可作为条件规则触发。这是一个行为差异。

### 1.4 作用域层级

| 层级 | Claude Code | CodeBuddy Code |
|------|-------------|----------------|
| 托管策略（Managed Policy） | 支持，由组织管理员统一推送 | 无此层 |
| 用户级 | `~/.claude/CLAUDE.md` + `~/.claude/rules/` | `~/.codebuddy/CODEBUDDY.md` + `~/.codebuddy/rules/` |
| 项目级 | `./CLAUDE.md` 或 `./.claude/CLAUDE.md` | `./CODEBUDDY.md` 或 `./.codebuddy/CODEBUDDY.md` |
| 项目本地 | `./CLAUDE.local.md` | `./CODEBUDDY.local.md` |

### 1.5 导入语法

两者均支持 `@path/to/file` 导入语法，差异很小：

| 维度 | Claude Code | CodeBuddy Code |
|------|-------------|----------------|
| 最大递归深度 | 4 级 | 5 级 |
| 路径类型 | 相对路径、绝对路径 | 相对路径、绝对路径 |

### 1.6 AGENTS.md 兼容

| 维度 | Claude Code | CodeBuddy Code |
|------|-------------|----------------|
| 回退加载 `AGENTS.md` | 需手动添加 `@AGENTS.md` 或创建 symlink | 自动回退：当 `CODEBUDDY.md` 不存在时加载 `AGENTS.md` |

---

## 2. Skill 组件差异

### 2.1 目录结构

两者 Skill 目录结构高度一致：

```
skill-name/
├── SKILL.md          # 核心定义文件（两者相同）
├── scripts/          # 可选脚本目录
├── references/       # 可选参考文档
└── assets/           # 可选资源文件
```

### 2.2 安装路径

| Scope | Claude Code | CodeBuddy Code |
|-------|-------------|----------------|
| 用户级 | `~/.claude/skills/<name>/SKILL.md` | `~/.codebuddy/skills/<name>/SKILL.md` |
| 项目级 | `.claude/skills/<name>/SKILL.md` | `.codebuddy/skills/<name>/SKILL.md` |
| 插件级 | `<plugin>/skills/<name>/SKILL.md` | `<plugin>/skills/<name>/SKILL.md` |

### 2.3 SKILL.md Frontmatter 字段

**Claude Code：**

```yaml
---
description: 生成 Git commit message
when_to_use: 需要创建规范的 commit message 时使用
allowed-tools: [Bash, Read, Write]
disallowed-tools: [Edit]
disable-model-invocation: false
user-invocable: true
context: fork
agent: commit-agent
model: default
---
```

**CodeBuddy Code：**

```yaml
---
description: 生成 Git commit message
allowed-tools: [Bash, Read, Write]
disable-model-invocation: false
user-invocable: true
context: fork
agent: commit-agent
model: default
---
```

| 字段 | Claude Code | CodeBuddy Code | 差异说明 |
|------|-------------|----------------|----------|
| `description` | 支持，无长度限制 | 支持，1024 字符限制 | CodeBuddy 有长度限制 |
| `when_to_use` | 支持 | **不支持** | CodeBuddy 无此字段 |
| `allowed-tools` | 支持 | 支持 | 一致 |
| `disallowed-tools` | 支持 | **不支持** | CodeBuddy 无此字段 |
| `disable-model-invocation` | 支持 | 支持 | 一致 |
| `user-invocable` | 支持 | 支持 | 一致 |
| `context` / `agent` / `model` | 支持（Fork 模式） | 支持（Fork 模式） | 一致 |

### 2.4 变量占位符

| 变量 | Claude Code | CodeBuddy Code |
|------|-------------|----------------|
| 命令行参数 | `$ARGUMENTS`, `$ARGUMENTS[0]` | `$ARGUMENTS`, `$ARGUMENTS[0]` |
| Skill 目录 | `${CLAUDE_SKILL_DIR}` | `${CODEBUDDY_SKILL_DIR}`（兼容 `${CLAUDE_SKILL_DIR}` 别名） |
| 工作目录 | `${CLAUDE_PROJECT_DIR}` | `${CODEBUDDY_PROJECT_DIR}`（兼容 `${CLAUDE_PROJECT_DIR}` 别名） |

> CodeBuddy Code 原生使用 `${CODEBUDDY_*}` 前缀，但同时兼容 `${CLAUDE_*}` 别名。从 Claude → CodeBuddy 方向转换时，变量名可直接保留（因兼容别名），也可选择替换为原生前缀。

### 2.5 内联 Shell 语法

两者均支持 `` !`command` `` 内联 Shell 执行语法，行为一致无差异。

---

## 3. 适配规则摘要

基于以上差异，PBrew 内置的适配规则（Claude → CodeBuddy 方向）如下：

### 3.1 Rule 适配

| 转换项 | 操作 |
|--------|------|
| 文件名 | `CLAUDE.md` → `CODEBUDDY.md`；`CLAUDE.local.md` → `CODEBUDDY.local.md` |
| 目录名 | `.claude/` → `.codebuddy/` |
| 路径前缀 | `~/.claude/` → `~/.codebuddy/` |
| Frontmatter | 补充 `alwaysApply: true`（如原文件无 frontmatter 或无此字段） |
| Frontmatter | 补充 `enabled: true`（同理） |

### 3.2 Skill 适配

| 转换项 | 操作 |
|--------|------|
| 安装路径 | `.claude/skills/` → `.codebuddy/skills/` |
| 路径前缀 | `~/.claude/skills/` → `~/.codebuddy/skills/` |
| Frontmatter | 移除 `when_to_use` 字段 |
| Frontmatter | 移除 `disallowed-tools` 字段 |
| 变量替换 | `${CLAUDE_SKILL_DIR}` → `${CODEBUDDY_SKILL_DIR}` |
| 变量替换 | `${CLAUDE_PROJECT_DIR}` → `${CODEBUDDY_PROJECT_DIR}` |
| 变量替换 | `$ARGUMENTS` 无需替换（两者一致） |

### 3.3 注意事项

- 反方向（CodeBuddy → Claude）转换在 MVP 中不需要，因为仓库默认使用 Claude 格式；
- CodeBuddy 独有的 `alwaysApply: false` + 无 `paths` 导致规则不加载的行为，在从 Claude 格式转换时不会触发（Claude 格式的规则要么有 frontmatter 的 `paths`，要么始终加载）；
- 如果未来需要支持 CodeBuddy 格式的组件源，应在 `.pbrew.json` 中标记 `sourceAgent` 字段，并补充 CodeBuddy → Claude 的反向适配规则。
