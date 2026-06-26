# pbrew — MVP 任务清单

## 阶段 0：项目脚手架

- [x] 初始化 Go Module（`go mod init github.com/user/pbrew`）
- [x] 创建目录结构（cmd/、internal/model/、internal/agent/ 等）
- [x] 入口 main.go + cobra 根命令 + `--verbose` 全局 flag

## 阶段 1：基础数据层（无外部依赖）

- [x] `internal/model/component.go` — Component、ComponentType、SourceAgent 定义
- [x] `internal/model/record.go` — InstallRecord、InstalledFile、RecordStore、HookHandlerRecord 定义
- [x] `internal/model/config.go` — Config 结构定义（扁平 `source` 字段）
- [x] `internal/model/agent.go` — AgentProfile、FrontmatterConstraints、FieldConstraint 定义
- [x] `internal/agent/registry.go` — 内置 claude / codebuddy AgentProfile 定义 + Get / List 查询接口

## 阶段 2：基础设施层

- [x] `internal/config/config.go` — 读写 `~/.pbrew/config.json`，支持 Load / Save / Get / Set（顶层 key 映射）
- [x] `internal/platform/paths.go` — FindProjectRoot（向上查找 .git/）、AgentRoot、InstallTargetDir、RecordFilePath

## 阶段 3：数据源层

- [x] `internal/source/source.go` — Source 接口定义（Root / ReadDir / Exists / WriteDir / Close）
- [x] `internal/source/local.go` — LocalSource 实现
- [x] `internal/source/git.go` — GitSource 实现（shallow clone `--depth 1` 到 `/tmp/pbrew-<random>/`，Close 清理临时目录）

## 阶段 4：组件发现

- [x] `internal/registry/index.go` — `_index.json` 解析
- [x] `internal/registry/registry.go` — ListComponents（优先读索引，退化全量遍历）、FindComponent、LoadComponent；校验 Rule 组件目录扁平性（如存在子目录则报错，Skill 组件允许子目录）

## 阶段 5：适配引擎

- [x] `internal/adapter/canonical.go` — CanonicalComponent 及各子结构定义
- [x] `internal/adapter/adapter.go` — Adapter 接口 + NewAdapter 工厂 + Transform 管道函数
- [x] `internal/adapter/frontmatter.go` — Frontmatter 解析 / 修改 / 序列化（基于 yaml.v3 Node API，保证 round-trip 保真）；仅供 Skill 适配使用
- [x] `internal/adapter/path_replace.go` — 路径与变量名替换（.claude/ ↔ .codebuddy/、${CLAUDE_*} ↔ ${CODEBUDDY_*}）
- [x] `internal/adapter/claude_adapter.go` — ClaudeAdapter：ToCanonical（Rule 原样传递 + Skill 解析 frontmatter）+ FromCanonical（Rule 原样输出 + Skill 写入 frontmatter）
- [x] `internal/adapter/codebuddy_adapter.go` — CodeBuddyAdapter：ToCanonical（Rule 原样传递 + Skill 解析 frontmatter）+ FromCanonical（Rule 原样输出 + Skill 跳过 when_to_use/disallowed-tools + 截断 description 至 1024 字符）

## 阶段 6：编排层

- [x] `internal/installer/install.go` — Install 流程：查找组件 → 加载源文件 → Transform → 检查冲突（无 --force 时交互确认）→ 写入文件 → 更新 _config.json（write-temp-then-rename 原子写入）
- [x] `internal/installer/uninstall.go` — Uninstall 流程：读取记录 → Rule 删除文件并清理空目录，Skill 删除整个安装子目录 → 移除记录（write-temp-then-rename 原子写入）
- [x] `internal/installer/update.go` — Update 流程：语义化版本对比（major.minor.patch 有序比较）→ 调用 Install 覆盖

## 阶段 7：CLI 命令层

- [x] `cmd/config.go` — pbrew config set / get / list
- [x] `cmd/list.go` — pbrew list [--agent \<agent\>] [--scope \<scope\>]；带 --agent 时列出已安装组件及版本状态
- [x] `cmd/install.go` — pbrew install \<name...\> --agent \<agent\> [--scope \<scope\>] [--force]
- [x] `cmd/update.go` — pbrew update [name...] / --all --agent \<agent\> [--scope \<scope\>]
- [x] `cmd/uninstall.go` — pbrew uninstall \<name...\> --agent \<agent\> --scope \<scope\>

## 阶段 8：测试

- [ ] `internal/adapter/` 单元测试（≥ 90% 覆盖率）：各方向 Rule/Skill 转换（Rule 不修改 frontmatter）、Frontmatter 解析 round-trip、路径替换
- [ ] `internal/agent/` 单元测试（≥ 90% 覆盖率）：注册表查询
- [ ] `internal/model/` 单元测试：序列化 / 反序列化
- [ ] `internal/platform/` 单元测试：路径计算（数据驱动）
- [ ] `internal/config/` 集成测试：配置读写（扁平 source 字段）
- [ ] `internal/registry/` 集成测试：组件发现（含 Rule 扁平性校验、Skill 允许子目录）、_index.json 解析
- [ ] `internal/source/` 集成测试：LocalSource / GitSource
- [ ] `internal/installer/` 集成测试（≥ 80% 覆盖率）：安装 / 卸载 / 更新 编排，验证 atomic write
- [ ] E2E 测试基础设施：testscript 入口 `e2e/main_test.go`
- [ ] E2E：pbrew config set / get / list
- [ ] E2E：pbrew list（数据源组件列表）
- [ ] E2E：pbrew list --agent（已安装组件列表 + 版本状态）
- [ ] E2E：pbrew install rule --agent
- [ ] E2E：pbrew install skill --agent
- [ ] E2E：pbrew update --all --agent
- [ ] E2E：pbrew uninstall --agent --scope
- [ ] E2E：文件系统验证（exists / cmp）
- [ ] E2E：语义化版本比较正确性验证
