# 本地编辑器编辑工具函数设计

## 概述

在 `pkg/util` 包中新增一个工具函数 `EditFile`，用于调用用户本地编辑器（vi/nano/vim/code 等）交互式编辑文件，覆盖两种使用场景：编辑已有文件和预填内容后再编辑。

## API

```go
package util

// EditFile opens a file in the user's preferred editor and blocks until the
// editor exits, then returns the file's contents.
//
//   path: 文件路径（必填），文件不存在时编辑器会创建新文件
//   content: 可选，不为 nil 时先写入 path 再打开编辑器
//
// 编辑器优先级：$VISUAL > $EDITOR > PATH 中查找 (vi > nano > vim > emacs > code --wait) > vi
func EditFile(path string, content []byte) ([]byte, error)
```

## 文件

- `pkg/util/editor.go` — 实现
- `pkg/util/editor_test.go` — 测试

## 实现细节

### 编辑器发现 (`detectEditor()`)

1. 依次读取 `$VISUAL`、`$EDITOR` 环境变量，非空即用
2. 在 `PATH` 中查找 `vi`、`nano`、`vim`、`emacs`、`code --wait`，找到即用
3. 兜底返回 `vi`

### 参数拆分 (`splitEditorArgs()`)

编辑器值可能包含参数（如 `code --wait`），按空格拆分为 `[cmd, arg1, arg2, ..., path]` 传给 `exec.Command`。

### 执行流程

```
1. content != nil → WriteFile(path, content, 0644)
2. editor = detectEditor()
3. args = splitEditorArgs(editor, path)
4. cmd = exec.Command(args[0], args[1:]...)
5. cmd.Stdin/Stdout/Stderr = os.Stdin/Stdout/Stderr
6. cmd.Run()  // 阻塞等待退出
7. os.ReadFile(path)  // 无论编辑器退出码，始终读回文件
```

### 错误处理

| 场景 | 行为 |
|------|------|
| path 为空 | 返回 `ErrEmptyPath` |
| 编辑器找不到 | 返回明确错误，列出已查找的编辑器列表 |
| content 写入失败 | 返回写入原始错误 |
| 编辑器非零退出 | 仍尝试读回文件内容 |
| 读回文件失败 | 返回读取错误 |

### 常量与错误变量

```go
var ErrEmptyPath = errors.New("file path is empty")

// detectEditor 中使用的编辑器查找顺序
var editorPriority = []string{"vi", "nano", "vim", "emacs", "code --wait"}
```

## 测试策略

1. 编辑器发现测试：mock `os.Getenv` 和 `exec.LookPath` 的返回，验证优先级
2. 参数拆分测试：空格拆分、无参编辑器、含参编辑器
3. `EditFile` 集成测试：用 `echo` 或 `cat > file` 模拟编辑器，验证文件内容被正确读回
4. 错误路径测试：空 path、编辑器不存在的友好错误
