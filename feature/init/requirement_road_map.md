# pbrew — MVP 任务清单

## 阶段 0：项目脚手架

- [ ] 初始化 Go Module（`go mod init github.com/user/pbrew`）
- [ ] 创建目录结构（cmd/、internal/model/、internal/agent/ 等）
- [ ] 入口 main.go + cobra 根命令 + `--verbose` 全局 flag

## 阶段 1：基础数据层（无外部依赖）

- [ ] `internal/model/component.go` — Component、ComponentType、SourceAgent 定义
- [ ] `internal/model/record.go` — InstallRecord、InstalledFile、RecordStore、HookHandlerRecord 定义
- [ ] `internal/model/config.go` — Config 结构定义
- [ ] `internal/model/agent.go` — AgentProfile、FrontmatterConstraints、FieldConstraint 定义
- [ ] `internal/agent/registry.go` — 内置 claude / codebuddy AgentProfile 定义 + Get / List 查询接口

## 阶段 2：基础设施层

- [ ] `internal/config/config.go` — 读写 `~/.pbrew/config.json`，支持 Load / Save / Get / Set（点分路径映射）
- [ ] `internal/platform/paths.go` — FindProjectRoot（向上查找 .git/）、AgentRoot、InstallTargetDir、RecordFilePath

## 阶段 3：数据源层

- [ ] `internal/source/source.go` — Source 接口定义（Root / ReadDir / Exists / WriteDir / Close）
- [ ] `internal/source/local.go` — LocalSource 实现
- [ ] `internal/source/git.go` — GitSource 实现（shallow clone `--depth 1` 到 `/tmp/pbrew-<random>/`，Close 清理临时目录）

## 阶段 4：组件发现

- [ ] `internal/registry/index.go` — `_index.json` 解析
- [ ] `internal/registry/registry.go` — ListComponents（优先读索引，退化全量遍历）、FindComponent、LoadComponent；校验组件目录扁平性（不允许嵌套子目录）

## 阶段 5：适配引擎

- [ ] `internal/adapter/canonical.go` — CanonicalComponent 及各子结构定义
- [ ] `internal/adapter/adapter.go` — Adapter 接口 + NewAdapter 工厂 + Transform 管道函数
- [ ] `internal/adapter/frontmatter.go` — Frontmatter 解析 / 修改 / 序列化（基于 yaml.v3 Node API，保证 round-trip 保真）
- [ ] `internal/adapter/path_replace.go` — 路径与变量名替换（.claude/ ↔ .codebuddy/、${CLAUDE_*} ↔ ${CODEBUDDY_*}）
- [ ] `internal/adapter/claude_adapter.go` — ClaudeAdapter：ToCanonical（Rule + Skill）+ FromCanonical（Rule + Skill）
- [ ] `internal/adapter/codebuddy_adapter.go` — CodeBuddyAdapter：ToCanonical（Rule + Skill）+ FromCanonical（Rule + Skill）；AlwaysApply:false + 无 Paths 时输出 warning

## 阶段 6：编排层

- [ ] `internal/installer/install.go` — Install 流程：查找组件 → 加载源文件 → Transform → 检查冲突（无 --force 时交互确认）→ 写入文件 → 更新 _config.json
- [ ] `internal/installer/uninstall.go` — Uninstall 流程：读取记录 → 删除文件并清理空目录 → 移除记录
- [ ] `internal/installer/update.go` — Update 流程：对比 version → 调用 Install 覆盖

## 阶段 7：CLI 命令层

- [ ] `cmd/config.go` — pbrew config set / get / list
- [ ] `cmd/list.go` — pbrew list（表格输出）
- [ ] `cmd/install.go` — pbrew install \<name...\> --agent \<agent\> [--scope \<scope\>] [--force]
- [ ] `cmd/update.go` — pbrew update [name...] / --all --agent \<agent\> [--scope \<scope\>]
- [ ] `cmd/uninstall.go` — pbrew uninstall \<name...\> --agent \<agent\> --scope \<scope\>

## 阶段 8：测试

- [ ] `internal/adapter/` 单元测试（≥ 90% 覆盖率）：各方向 Rule/Skill 转换、Frontmatter 解析 round-trip、路径替换
- [ ] `internal/agent/` 单元测试（≥ 90% 覆盖率）：注册表查询
- [ ] `internal/model/` 单元测试：序列化 / 反序列化
- [ ] `internal/platform/` 单元测试：路径计算（数据驱动）
- [ ] `internal/config/` 集成测试：配置读写
- [ ] `internal/registry/` 集成测试：组件发现（含扁平性校验）、_index.json 解析
- [ ] `internal/source/` 集成测试：LocalSource / GitSource
- [ ] `internal/installer/` 集成测试（≥ 80% 覆盖率）：安装 / 卸载 / 更新 编排
- [ ] E2E 测试基础设施：testscript 入口 `e2e/main_test.go`
- [ ] E2E：pbrew config set / get / list
- [ ] E2E：pbrew list
- [ ] E2E：pbrew install rule --agent
- [ ] E2E：pbrew install skill --agent
- [ ] E2E：pbrew update --all --agent
- [ ] E2E：pbrew uninstall --agent --scope
- [ ] E2E：文件系统验证（exists / cmp）
