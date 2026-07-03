## Why

当前 `joy tpl` 仅有 `exec` 子命令用于执行模板，没有便捷的编辑模板配置参数（`setting.toml`）的方式。用户需要手动定位并编辑 `~/.seneschal/tpl/<name>/setting.toml` 文件，流程繁琐。添加 `edit` 子命令，与现有编辑器集成（`$VISUAL > $EDITOR > vi`），可以显著改善模板配置体验。

## What Changes

- **新增** `joy tpl edit <template_name>` 子命令，打开指定模板的 `setting.toml` 文件到编辑器
- 复用已有的 `pkg/util/editor.go:EditFile()` 实现编辑器调用
- 复用已有的 `internal/command/file/` 路径解析逻辑（`joyTplGetTemplateInfo`、`config.TplDir`、`config.TplSettingName`）

## Capabilities

### New Capabilities
- `tpl-params-edit`: 通过 `joy tpl edit <name>` 交互式编辑模板配置参数（`setting.toml`）

### Modified Capabilities
- （无）

## Impact

- `cmd/joy.go`: 新增 `joyTplEditCmd`（及对应 `init()` 注册），约 30-50 行
- `pkg/util/editor.go`: 无变动（复用 `EditFile`）
- 无外部依赖变更
