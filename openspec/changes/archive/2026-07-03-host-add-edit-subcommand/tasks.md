## 1. 新增 `host edit` 子命令

- [x] 1.1 在 `cmd/host.go` 中定义 `hostEditCmd` cobra 命令（`Use`, `Short`, `Example`, `Args`, `RunE`）
- [x] 1.2 在 `init()` 中添加 `hostCmd.AddCommand(hostEditCmd)` 注册子命令
- [x] 1.3 更新 `hostCmd` 的 `Example` 字段，追加 `edit` 到提示列表

## 2. 实现编辑逻辑

- [x] 2.1 `RunE` 中基于 `<alias>` 拼接配置文件路径 `config.SSHConfigDir/<alias>.toml`
- [x] 2.2 检查文件存在性，不存在则返回 "host alias not found" 错误
- [x] 2.3 调用 `util.EditFile(path, nil)` 打开编辑器（不填充初始内容，编辑已有文件）
- [x] 2.4 编辑器退出后，读取编辑结果并通过 `ReadSSHConfigFromToml` 校验 TOML 合法性
- [x] 2.5 编辑成功时 slog.Info 记录；编辑器非零退出时 slog.Warn 提示但不阻塞

## 3. 测试

- [x] 3.1 为 `host edit` 命令编写测试用例（文件存在、文件不存在、TOML 无效等场景）
