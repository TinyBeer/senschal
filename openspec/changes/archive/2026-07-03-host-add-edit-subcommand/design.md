## Context

当前 `host` 命令体系支持添加、查看、部署、上传下载等操作，但缺少编辑已有配置的能力。用户修改 SSH 配置需要手动定位 `~/.seneschal/conf/ssh/<alias>.toml` 文件并用编辑器打开，体验不够流畅。

新增 `host edit <alias>` 子命令，复用项目已有的 `pkg/util.EditFile` 在终端编辑器中直接打开配置，编辑退出后自动校验。

## Goals / Non-Goals

**Goals:**
- 通过 `host edit <alias>` 在用户编辑器中打开对应主机的 TOML 配置文件
- 保存退出后自动重新读取并校验配置完整性
- 最小化改动，复用现有模式（Cobra 子命令注册、SSH 配置路径定位）

**Non-Goals:**
- 不涉及配置字段的交互式引导修改（那是 `add` 的职责），`edit` 只是打开编辑器供用户自由编辑
- 不做字段级别的校验（viper 反序列化时的缺失字段按默认值处理即可）
- 不改动 `pkg/util.EditFile` 本身

## Decisions

1. **通过 alias 定位配置文件路径** — 不使用 `config.GetSSHConfigMap()` 先读取再反查文件名的间接方式。直接用 `config.SSHConfigDir/<alias>.toml` 拼接路径。如果文件不存在，给出明确错误提示
2. **编辑后执行 `readSSHConfigFromToml` 校验** — 编辑退出后调用内部 `readSSHConfigFromToml` 尝试重新解析 TOML 文件。如果解析失败（格式错误、字段类型不匹配等），给出错误信息并提示用户文件未被保存
3. **初始内容预填充** — 不填充初始内容。直接打开已有文件。因为编辑的是已存在配置，不是新建

## Risks / Trade-offs

- **TOML 格式损坏** — 如果用户保存了无效的 TOML，会导致解析失败。设计上用 `readSSHConfigFromToml` 解析并返回错误，用户可以重新 `edit` 修复
- **编辑器退出非零** — `EditFile` 已处理该情况：即使非零退出也读取文件内容，通过 `ErrEditorExitedNonZero` 包装返回。在 edit 命令中如果读取成功但编辑器非零退出，slog.Warn 提示但不阻止
