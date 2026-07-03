## Context

当前 `cmd/joy.go` 中已实现 `joy tpl`（列模板列表）和 `joy tpl exec <name>`（执行模板渲染），模板配置参数存储在 `~/.seneschal/tpl/<name>/setting.toml`。

缺失的环节：用户修改模板参数需手动定位 TOML 文件路径并用文本编辑器打开。本次变更添加 `joy tpl edit <name>` 子命令，方便地复用项目中已有的 `pkg/util/editor.go:EditFile()` 进行编辑。

## Goals / Non-Goals

**Goals:**

- 提供 `joy tpl edit <name>` — 交互式编辑器修改 `setting.toml`
- 复用已有基础设施：`EditFile`、`joyTplListTemplateInfo`、`config.TplDir`、`config.TplSettingName`
- 所有编辑器交互、模板不存在等异常均有清晰错误消息

**Non-Goals:**

- 不添加 `--set`/`--show` 等 CLI flag 模式
- 不修改 `joy tpl exec` 的现有行为
- 不引入新的外部依赖
- 不修改 `setting.toml` 的格式/结构
- 不添加模板参数校验/验证逻辑（仅编辑，不验证）

## Decisions

### Decision 1: 注册为 `joy tpl edit` 子命令

**选择**: 作为 `joy tpl` 的子命令，与 `exec` 同级。

**理由**: 与现有 `joy tpl exec` 保持一致；Cobra 父子命令的 flag 继承和帮助文本自然聚合；路径发现逻辑可复用 `joyTplListTemplateInfo`。

**被拒方案**: 独立 `joy tplconfig` 命令 — 会增加整体命令树的认知负担。

### Decision 2: 复用 `EditFile` 而非自建文件编辑

**选择**: 直接调用 `pkg/util/editor.go:EditFile()` 打开 `setting.toml`。

**理由**: 项目中已有成熟的编辑器调用封装，支持 `$VISUAL > $EDITOR > vi` 回退、阻塞等待、内容读取。无需重新实现。

**被拒方案**: 自建文件读写 + 输出到 stdout 提示用户手动编辑 — 体验远不如直接打开编辑器，且增加维护负担。

## Risks / Trade-offs

- **[低] 编辑器进程阻塞**: `EditFile` 阻塞等待编辑器退出，在无图形环境的 SSH 会话中编辑器可能为 `vi`，符合 CLI 预期。
