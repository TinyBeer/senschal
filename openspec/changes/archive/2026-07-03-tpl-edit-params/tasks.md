## 1. 新增 `joy tpl edit` 命令框架

- [x] 1.1 在 `cmd/joy.go` 中定义 `joyTplEditCmd`（Cobra 命令结构），注册 `edit` 子命令到 `joyTplCmd`
- [x] 1.2 实现命令基本流程：模板名称校验 → 定位模板路径 → 定位 `setting.toml`

## 2. 交互式编辑功能

- [x] 2.1 调用 `pkg/util/editor.go:EditFile()` 打开 `setting.toml`，等待编辑器退出
- [x] 2.2 编辑器退出后打印确认消息（包含模板名称和路径）

## 3. 错误处理与边界情况

- [x] 3.1 模板名称不存在 → 报错 "template '<name>' not found"
- [x] 3.2 模板没有 `setting.toml` → 报错 "setting file not found"
- [x] 3.3 无参数调用 `joy tpl edit` → 显示 usage 并返回非零退出码
- [x] 3.4 编辑器退出非零 → 返回 `ErrEditorExitedNonZero`

## 4. 测试

- [x] 4.1 为 `joyTplEditCmd` 添加单元测试（mock 编辑器、mock 文件系统）
- [x] 4.2 验证模板不存在和 setting 不存在的错误路径
- [x] 4.3 验证编辑后文件内容正确更新
