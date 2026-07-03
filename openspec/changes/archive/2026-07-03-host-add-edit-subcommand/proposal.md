## Why

当前 `host` 命令支持 `add`、`list`、`check`、`deploy`、`up`、`down`，但没有编辑已有配置的功能。用户需要修改 SSH 配置时只能手动找到文件编辑，体验割裂。新增 `edit` 子命令，复用 `pkg/util.EditFile` 在终端编辑器中直接修改配置，编辑后自动加载校验。

## What Changes

- 新增 `host edit <alias>` 子命令
- 通过 `pkg/util.EditFile` 在用户编辑器中打开对应主机的 TOML 配置文件
- 编辑保存退出后，重新读取并校验配置完整性（非空校验）
- 支持 `--alias` 自动补全选器（当前已有的模式，不需要改动）
- `host` 命令的 example 字段同步更新

## Capabilities

### New Capabilities
- `host-edit`: 编辑已添加主机的 SSH 配置。通过 alias 定位配置文件路径，在用户编辑器中打开 TOML 文件，保存退出后自动校验配置合法性

### Modified Capabilities

<!-- None -->

## Impact

- `cmd/host.go`：新增 `hostEditCmd`，注册到 `hostCmd`，并在 `init()` 中添加 `hostCmd.AddCommand(hostEditCmd)`
- 依赖 `pkg/util.EditFile`（已有实现）
- 依赖 `config.SSHConfigDir` 和 `config.GetSSHConfigMap()` 定位文件路径
- 无外部依赖变更，无 breaking changes
