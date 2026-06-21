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
| `sourceAgent` | 组件源格式标识（如 `claude`、`codebuddy`），默认 `claude` |

> **`sourceAgent` 字段说明**：虽然在 MVP 中仓库默认使用 Claude Code 格式，但预留此字段以便未来支持 CodeBuddy 格式的组件源。安装时 pbrew 可根据 `sourceAgent` 与目标 Agent 决定适配方向（如 `sourceAgent: codebuddy` + 目标 `claude` 时执行反向适配）。MVP 阶段此字段可省略，默认值为 `claude`。

### 1.3 目标用户

pbrew 同时面向**个人开发者**和**团队/组织**：

- **个人开发者**：在不同设备、不同项目间同步自己常用的规则和 Skill；
- **团队/组织**：统一维护一套团队规范组件，成员通过 `pbrew install` 快速同步。

MVP 阶段不考虑权限控制（谁能 publish）和强制版本策略，这些留待后续迭代。

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

pbrew 在安装组件后，会在对应 Agent 目录下创建 `.pbrew/` 子目录，其中放置 `_config.json` 记录文件，记录已安装组件的名称、版本、来源等信息。

- 该文件是 `pbrew update` 判断组件是否需要更新的基准；
- pbrew 默认用户不会手动破坏此文件；
- 不对此文件做额外的校验或防护机制。

> **`.pbrew/` 目录约定**：该目录下仅允许放置 `.json` 文件，禁止放置 `.md` 文件——因为 CodeBuddy Code 会递归发现 Agent 目录下的所有 `.md` 文件并当作规则加载，放入 `.md` 文件会导致误加载。

### 4.4 单一数据源

用户绑定一个数据源（Git 仓库或本地目录）作为组件来源。`pbrew list`、`install`、`update` 均基于该数据源操作。

### 4.5 项目根目录的确定

当 `--scope project`（默认值）时，pbrew 需要确定"项目根目录"。策略为：**从当前工作目录向上查找，取最近的包含 `.git/` 的目录作为项目根目录**。这与 Claude Code 和 CodeBuddy Code 的行为一致——它们都以 Git 仓库边界作为项目边界。

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
pbrew install <component-name...> --agent <agent-id> [--scope <scope>]
```

- `<component-name...>`：一个或多个组件名称
- `--agent`：**必填**，目标 Agent（如 `claude`、`codebuddy`）
- `--scope`：可选，安装范围，默认 `project`。支持以下值：
  - `user`：用户级（如 `~/.claude/` 或 `~/.codebuddy/`）
  - `project`：项目级（如 `./.claude/` 或 `./.codebuddy/`），默认值

> **一次安装只支持一个 Agent**。如果需要将同一组件安装到多个 Agent，需分别执行命令。

安装流程：

1. 从数据源读取组件目录及 `.pbrew.json`
2. 根据目标 Agent 查找内置适配规则，对组件文件执行转换
3. 将转换后的文件写入目标 Agent 对应的安装路径
4. 在该 Agent 目录下的 `.pbrew/_config.json` 中更新安装记录，记录内容包括组件名称、版本、来源以及**已安装文件的路径信息**

> **文件路径记录说明**：
> - Rule 组件：记录已安装的文件路径（Rule 可以是单个 `.md` 文件，也可以是子目录下的 `.md` 文件，两个 Agent 均支持递归发现 `rules/` 下的子目录）
> - Skill 组件：记录已安装的 Skill 子目录路径（卸载时直接删除整个子目录）

### 5.4 AI 组件的更新（pbrew update）

```bash
# 更新指定组件
pbrew update <component-name...> --agent <agent-id> [--scope <scope>]

# 更新指定范围内所有已安装组件
pbrew update --all --agent <agent-id> [--scope <scope>
```

更新流程：

1. 读取目标 Agent 目录下 `.pbrew/_config.json` 中的安装记录
2. 对比数据源中组件的 `version` 与已安装版本
3. 如有新版本，重新执行安装流程覆盖旧文件并更新记录

> **MVP 版本只对比 `version` 字段**，不做 `hash` 对比。`hash` 对比留作后续迭代，用于检测用户本地手动修改等场景。

### 5.5 AI 组件的卸载（pbrew uninstall）

```bash
pbrew uninstall <component-name...> --agent <agent-id> --scope <scope>
```

卸载流程：

1. 用户指定要卸载的组件名称
2. 读取 `.pbrew/_config.json` 安装记录中该组件的文件路径信息
3. 根据记录删除对应的文件：
   - Rule：删除记录中列出的具体 `.md` 文件，并清理产生的空目录
   - Skill：删除整个 Skill 子目录
4. 从 `.pbrew/_config.json` 安装记录中移除该组件条目

### 5.6 AI 组件的发布（pbrew publish）

```bash
pbrew publish --name <component-name> --dir <component-dir>
```

- `--name`：期望发布的组件名称
- `--dir`：组件所在目录

发布流程：

1. 读取组件目录，根据组件类型和内置适配规则，将组件格式转换为仓库默认格式（Claude Code）
2. 自动生成版本号：
   - 首次发布：`0.1.0`
   - 更新发布：`patch +1`（如 `1.0.3` → `1.0.4`）
3. 生成 `.pbrew.json` 元数据文件
4. 将组件写入已绑定的数据源目录
5. **如果数据源中已存在同名且同版本号的组件，直接 publish 失败，要求用户手动修改版本号后重试**
6. 输出后续 git 提交命令供用户手动执行

## 6. 组件仓库结构

### 6.1 推荐结构

```
repo/
├── _index.json                       # 组件索引，加速 list 查询
├── rules/
│   ├── git-conventions/
│   │   ├── .pbrew.json               # 组件元数据
│   │   └── git-conventions.md        # 规则正文
│   └── api-style/
│       ├── .pbrew.json
│       └── ...
└── skills/
    ├── commit/
    │   ├── .pbrew.json
    │   ├── SKILL.md
    │   ├── scripts/
    │   └── references/
    └── review/
        ├── .pbrew.json
        └── ...
```

- **按 type 分目录**（rules/、skills/）：直观，list 时可按类型筛选
- **`_index.json` 根索引**：内容是所有组件的摘要（name、type、version、description、path），`pbrew list` 只读这一个文件，不用遍历子目录
- MVP 阶段索引可手动维护，后续可在 publish 时自动生成

`_index.json` 示例：

```json
{
  "version": 1,
  "components": [
    {
      "name": "git-conventions",
      "type": "rule",
      "version": "1.0.0",
      "description": "Git 提交与分支规范",
      "path": "rules/git-conventions"
    },
    {
      "name": "commit",
      "type": "skill",
      "version": "0.2.0",
      "description": "生成规范的 commit message",
      "path": "skills/commit"
    }
  ]
}
```

### 6.2 自定义目录结构

pbrew 支持用户根据不同场景自定义仓库的目录结构。程序会**遍历组件仓库下的所有目录**，递归寻找所有包含有效 `.pbrew.json` 的目录，将其识别为合法的 AI 组件。`_index.json` 是可选的加速优化，不存在时退化为全量遍历。

## 7. 适配规则补充约定

### 7.1 Rule Frontmatter 转换的精细化

Claude → CodeBuddy 方向转换时，`alwaysApply` 的设置不应一律为 `true`，需根据源文件 frontmatter 判断：

| 源文件（Claude 格式） | 转换后（CodeBuddy 格式） | 说明 |
|----------------------|-------------------------|------|
| 无 frontmatter | `alwaysApply: true` | 原规则始终加载 |
| 有 `paths` 字段 | `alwaysApply: false` | 原规则为条件触发 |
| 已有 `alwaysApply` | 保持原值 | 无需改动 |

同时始终补充 `enabled: true`。

### 7.2 Rule 内容中 `@` 引用路径的替换

规则正文中的 `@path/to/file` 导入语法，如果路径中包含 `.claude/` 前缀，需要在转换为 CodeBuddy 格式时替换为 `.codebuddy/`。

### 7.3 Skill `description` 超长截断

当转换为 CodeBuddy 格式时，如果 Skill 的 `description` 超过 1024 字符，直接截断至 1024 字符。

### 7.4 安装冲突策略

如果目标路径已存在同名文件（用户手动创建的），MVP 采用**覆盖**策略，但在覆盖前提示用户确认。后续迭代可增加更细粒度的冲突处理（跳过、备份后覆盖等）。

## 8. MVP 范围

### 8.1 命令

| 命令 | 优先级 |
|------|--------|
| `pbrew config set/get/list` | P0 |
| `pbrew list` | P0 |
| `pbrew install` | P0 |
| `pbrew update` | P0 |
| `pbrew uninstall` | P0 |
| `pbrew publish` | P1（MVP 之后） |

### 8.2 数据源

MVP 支持以下数据源类型：

- Git 仓库（通过 URL 指定）
- 本地目录

### 8.3 Agent 支持

MVP 支持以下 Agent：

- Claude Code
- CodeBuddy Code

### 8.4 组件类型

MVP 支持以下组件类型：

- Rule（规则）
- Skill

> Hook 组件的类型定义（目录结构、文件格式等）尚未明确，MVP 阶段不支持，待后续迭代补充需求后再实现。

## 9. 已知限制与后续迭代项

| 项目 | 当前状态 | 后续计划 |
|------|---------|---------|
| 权限控制 | 无，任何用户均可 publish | 团队/组织场景下增加权限管理 |
| `hash` 对比更新 | 仅对比 `version` | 增加 `hash` 对比，检测本地手动修改 |
| 并发安装安全 | 无文件锁保护 | 增加文件锁，防止并发安装导致安装记录损坏 |
| CodeBuddy → Claude 反向适配 | 不支持 | 在 `.pbrew.json` 中通过 `sourceAgent` 支持 |
| Hook 组件 | 类型定义未明确 | 补充 Hook 格式差异对比与适配规则 |
| 安装冲突策略 | 覆盖前提示确认 | 增加跳过、备份后覆盖等策略 |
| Git 数据源认证 | 未定义 | 支持私有仓库的 SSH key / token |
| 版本锁定/范围 | 仅支持最新版本 | 支持指定版本号安装（如 `pbrew install rule@1.0.0`） |
| `_index.json` 自动生成 | 手动维护 | publish 时自动更新索引 |

## 10. 技术要求

- 项目需集成端到端（E2E）测试能力，确保各命令功能可及时验证；
- 后续迭代中通过 E2E 测试保障功能稳定性。
