# CONTRIBUTING.md - 贡献指南

感谢你对 Distill 项目的关注！以下是参与贡献的规范和流程。

## 目录

- [开发环境](#开发环境)
- [项目结构](#项目结构)
- [贡献流程](#贡献流程)
- [分支命名](#分支命名)
- [Commit 规范](#commit-规范)
- [代码规范](#代码规范)
- [Pull Request 规范](#pull-request-规范)
- [Git Hooks](#git-hooks)

---

## 开发环境

**要求：**

- Go >= 1.26
- Git（推荐配置 Git Hooks）

**初始化：**

```bash
git clone https://github.com/ZouZhao321/distill.git
cd Distill
go mod download
go build -o distill .
```

**常用命令（Makefile）：**

```bash
make build   # 编译
make test    # 运行全部测试（带覆盖率和详细输出）
make lint    # 静态检查
make clean   # 清理二进制文件
```

## 项目结构

```
Distill/
├── cmd/              # CLI 命令入口（cobra 子命令）
├── internal/
│   └── core/
│       ├── domain/   # 领域实体、错误类型、配置结构
│       └── usecase/  # 业务用例（核心逻辑）
├── scripts/
│   └── hooks/        # Git Hooks（pre-commit、commit-msg）
├── doc/              # 设计文档
├── go.mod
├── Makefile
└── main.go           # 程序入口
```

项目采用 **Clean Architecture**，核心业务逻辑位于 `internal/core/`，CLI 层位于 `cmd/`。

## 贡献流程

1. **先开 Issue** — 在 GitHub 上创建 Issue，描述要解决的问题或新功能建议，等待讨论确认后再动手
2. **Fork 仓库** — Fork 到自己的账号下
3. **创建分支** — 从最新的 `main` 创建功能分支（见[分支命名](#分支命名)）
4. **开发与测试** — 编写代码和测试，确保本地检查全部通过
5. **提交代码** — 遵循 [Commit 规范](#commit-规范)
6. **推送并创建 PR** — 推送到你的 Fork，向主仓库发起 Pull Request

> 如果你是项目成员（有直接 push 权限），可以跳过 Fork，直接在仓库内创建分支。

## 分支命名

格式：`<type>/<issue-id>-<简短描述>`

```
feat/15-add-user-auth
fix/3-log-output-to-file
docs/8-update-readme
chore/add-merge-template
```

**type 对应关系：**

| type | 用途 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档变更 |
| `refactor` | 代码重构 |
| `test` | 测试相关 |
| `chore` | 构建、配置、杂项 |

## Commit 规范

遵循 **Conventional Commits** 格式：

```
<type>(<scope>): <subject>
```

- **type**：`feat`、`fix`、`docs`、`style`、`refactor`、`perf`、`test`、`build`、`ci`、`chore`、`revert`
- **scope**（可选）：影响的模块，如 `core`、`cli`、`init`、`log`
- **subject**：中文描述，说明做了什么，不超过 100 字符

**示例：**

```
feat(core): 实现 AddAssetUseCase
fix(cli): 修复 checkout 覆盖策略
chore: 更新 .gitignore
```

**注意：**

- 不要在 commit message 中写开发方法（TDD）、测试数量等内部细节
- 代码注释使用中文，包注释以 `// Package xxx` 开头

Git Hooks 会自动校验 commit 格式，不符合规范将无法提交。

## 代码规范

- **格式化**：使用 `gofmt`（Git Hooks 会自动检查）
- **静态检查**：使用 `go vet`（Git Hooks 会自动检查）
- **注释语言**：中文
- **换行符**：Go 文件强制 LF（`.gitattributes` 已配置）
- **TOML 路径**：Windows 路径在 TOML 配置中必须使用正斜杠（`C:/Users/...`），反斜杠会被误读为 Unicode 转义

## Pull Request 规范

**标题格式：** 与 Commit 规范一致

```
feat(core): 实现资产搜索功能
fix(log): 日志写入文件不再仅输出到控制台
```

**描述模板：**

```markdown
## 变更说明
<!-- 一两句话说清楚这个 PR 做了什么 -->

## 动机
<!-- 为什么要做这个改动？关联哪个 Issue -->
Fixes #3

## 改动内容
<!-- 具体改了哪些文件/模块，关键设计决策 -->

## 测试方式
<!-- 怎么验证这个改动是正确的 -->

## Checklist
- [ ] 本地编译通过
- [ ] 单元测试通过
- [ ] 无 go vet 警告
```

**提交前检查清单：**

- [ ] 从最新 `main` 分支创建，已 rebase
- [ ] 所有测试通过（`make test`）
- [ ] 静态检查通过（`make lint`）
- [ ] 无调试代码、临时文件
- [ ] PR 只包含本次变更，不混入无关改动
- [ ] 如果有关联 Issue，在描述中用 `Fixes #id` 关联

## Git Hooks

项目通过 `core.hooksPath` 共享以下 Hooks，Clone 后需手动配置：

```bash
git config core.hooksPath scripts/hooks
```

| Hook | 作用 |
|------|------|
| `pre-commit` | `gofmt` → `go vet` → `go test -short`，任一失败则阻止提交 |
| `commit-msg` | 校验 Commit Message 是否符合 Conventional Commits 格式 |
