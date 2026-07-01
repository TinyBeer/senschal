# EditFile 工具函数实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 `pkg/util` 包中实现 `EditFile` 函数，支持通过本地编辑器交互式编辑文件

**Architecture:** 编辑器发现（`$VISUAL` > `$EDITOR` > PATH 扫描）→ 参数拆分 → `exec.Command` 连接到当前 TTY 执行 → 读回文件内容。分两个内部函数（`detectEditor`, `splitEditorArgs`）+ 一个公开 API（`EditFile`）。

**Tech Stack:** Go 标准库（`os/exec`, `os`, `strings`, `errors`）

## Global Constraints

- TDD: 先写测试（预期失败），再写刚好通过的最少代码
- 测试框架: `github.com/stretchr/testify`
- 不添加外部依赖，只使用 Go 标准库
- 遵循 CLAUDE.md 编码规范，不做猜测式扩展

---

### Task 1: 内部辅助函数 (detectEditor + splitEditorArgs)

**Files:**
- Create: `pkg/util/editor.go` — 实现
- Create: `pkg/util/editor_test.go` — 测试

**Interfaces:**
- Consumes: 无
- Produces:
  - `func detectEditor() string` — 返回编辑器命令（可能含参数）
  - `func splitEditorArgs(editor, filePath string) []string` — 返回 `[cmd, arg1, ..., filePath]`

- [x] **Step 1: 写 splitEditorArgs 的测试**

写入 `pkg/util/editor_test.go`:

```go
package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitEditorArgs_NoArgs(t *testing.T) {
	args := splitEditorArgs("vim", "/tmp/test.md")
	assert.Equal(t, []string{"vim", "/tmp/test.md"}, args)
}

func TestSplitEditorArgs_WithArgs(t *testing.T) {
	args := splitEditorArgs("code --wait", "/tmp/test.md")
	assert.Equal(t, []string{"code", "--wait", "/tmp/test.md"}, args)
}

func TestSplitEditorArgs_EmptyEditor(t *testing.T) {
	args := splitEditorArgs("", "/tmp/test.md")
	assert.Equal(t, []string{"/tmp/test.md"}, args)
}
```

- [x] **Step 2: 运行测试，验证失败**

```bash
go test ./pkg/util/ -run TestSplitEditorArgs -v
```
Expected: 编译失败，`splitEditorArgs` 未定义

- [x] **Step 3: 写最小实现**

写入 `pkg/util/editor.go`:

```go
package util

import "strings"

func splitEditorArgs(editor, filePath string) []string {
	if editor == "" {
		return []string{filePath}
	}
	parts := strings.Fields(editor)
	return append(parts, filePath)
}
```

- [x] **Step 4: 运行测试，验证通过**

```bash
go test ./pkg/util/ -run TestSplitEditorArgs -v
```
Expected: ALL PASS

- [x] **Step 5: 写 detectEditor 的测试**

追加到 `pkg/util/editor_test.go`:

```go
func TestDetectEditor_FromVISUAL(t *testing.T) {
	t.Setenv("VISUAL", "nano")
	t.Setenv("EDITOR", "vim") // VISUAL 优先，EDITOR 应被忽略
	editor := detectEditor()
	assert.Equal(t, "nano", editor)
}

func TestDetectEditor_FromEDITOR(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "emacs")
	editor := detectEditor()
	assert.Equal(t, "emacs", editor)
}

func TestDetectEditor_FallbackToVi(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	t.Setenv("PATH", "/no/such/path") // 清空 PATH，触发兜底
	editor := detectEditor()
	assert.Equal(t, "vi", editor)
}
```

- [x] **Step 6: 运行测试，验证失败**

```bash
go test ./pkg/util/ -run TestDetectEditor -v
```
Expected: 编译失败，`detectEditor` 未定义

- [x] **Step 7: 实现 detectEditor**

追加到 `pkg/util/editor.go`:

```go
import (
	"os"
	"os/exec"
)

// editorPriority 是 PATH 中查找编辑器的顺序（优先靠前）
var editorPriority = []string{"vi", "nano", "vim", "emacs", "code --wait"}

// detectEditor 返回当前系统可用的编辑器命令。
// 优先级：$VISUAL > $EDITOR > PATH 查找 > "vi"
func detectEditor() string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if e := os.Getenv(env); e != "" {
			return e
		}
	}
	for _, name := range editorPriority {
		// code --wait 取第一部分（"code"）来查 PATH
		cmd := strings.Fields(name)[0]
		if _, err := exec.LookPath(cmd); err == nil {
			return name
		}
	}
	return "vi"
}
```

- [x] **Step 8: 运行测试，验证通过**

```bash
go test ./pkg/util/ -run "TestDetectEditor|TestSplitEditorArgs" -v
```
Expected: ALL PASS

- [x] **Step 9: 提交**

```bash
git add pkg/util/editor.go pkg/util/editor_test.go
git commit -m "feat(util): add detectEditor and splitEditorArgs helpers"
```

---

### Task 2: EditFile 主函数

**Files:**
- Modify: `pkg/util/editor.go` — 添加 EditFile 和 ErrEmptyPath
- Modify: `pkg/util/editor_test.go` — 添加 EditFile 测试

**Interfaces:**
- Consumes: `detectEditor()`, `splitEditorArgs()` (Task 1)
- Produces: `func EditFile(path string, content []byte) ([]byte, error)`
- Errors: `var ErrEmptyPath`

- [x] **Step 1: 写 EditFile 测试**

追加到 `pkg/util/editor_test.go`:

```go
import (
	"os"
	"path/filepath"
)

func TestEditFile_EmptyPath(t *testing.T) {
	_, err := EditFile("", nil)
	assert.ErrorIs(t, err, ErrEmptyPath)
}

func TestEditFile_EditorNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello"), 0644)

	// 强制使用不存在的编辑器
	t.Setenv("VISUAL", "/no/such/editor/binary")
	t.Setenv("EDITOR", "")

	_, err := EditFile(file, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEditFile_WithMockEditor(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(targetFile, []byte("original"), 0644)

	// 创建模拟编辑器脚本 — 把固定内容写入目标文件
	mockEditor := filepath.Join(tmpDir, "mock_editor.sh")
	os.WriteFile(mockEditor, []byte("#!/bin/sh\nprintf 'edited content' > \"$1\"\n"), 0755)

	t.Setenv("VISUAL", mockEditor)
	t.Setenv("EDITOR", "")

	content, err := EditFile(targetFile, nil)
	assert.NoError(t, err)
	assert.Equal(t, "edited content", string(content))
}

func TestEditFile_WithInitialContent(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")

	mockEditor := filepath.Join(tmpDir, "mock_editor.sh")
	os.WriteFile(mockEditor, []byte("#!/bin/sh\nprintf 'new content' > \"$1\"\n"), 0755)

	t.Setenv("VISUAL", mockEditor)
	t.Setenv("EDITOR", "")

	content, err := EditFile(targetFile, []byte("initial"))
	assert.NoError(t, err)
	assert.Equal(t, "new content", string(content))

	// 验证磁盘上的文件也被更新
	readBack, _ := os.ReadFile(targetFile)
	assert.Equal(t, "new content", string(readBack))
}
```

- [x] **Step 2: 运行测试，验证失败**

```bash
go test ./pkg/util/ -run TestEditFile -v
```
Expected: 编译失败，`EditFile` 和 `ErrEmptyPath` 未定义

- [x] **Step 3: 写最小实现**

更新 `pkg/util/editor.go`，添加完整实现：

```go
package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ErrEmptyPath 表示传入的文件路径为空
var ErrEmptyPath = errors.New("file path is empty")

// editorPriority is the list of editors to search in PATH
var editorPriority = []string{"vi", "nano", "vim", "emacs", "code --wait"}

// EditFile opens a file in the user's preferred editor and blocks until the
// editor exits, then returns the file's contents.
//
// path: 文件路径（必填），文件不存在时编辑器会创建新文件
// content: 可选，不为 nil 时先写入 path 再打开编辑器
//
// 编辑器优先级：$VISUAL > $EDITOR > PATH 查找 > "vi"
func EditFile(path string, content []byte) ([]byte, error) {
	if path == "" {
		return nil, ErrEmptyPath
	}

	// 如果有初始内容，先写入文件
	if content != nil {
		if err := os.WriteFile(path, content, 0644); err != nil {
			return nil, fmt.Errorf("write initial content: %w", err)
		}
	}

	editor := detectEditor()
	args := splitEditorArgs(editor, path)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// 找不到可执行文件 → 明确错误
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("editor %q not found", editor)
		}
		// 其他错误（如非零退出），继续读取文件
	}

	out, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read edited file: %w", err)
	}
	return out, nil
}

func detectEditor() string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if e := os.Getenv(env); e != "" {
			return e
		}
	}
	for _, name := range editorPriority {
		cmd := strings.Fields(name)[0]
		if _, err := exec.LookPath(cmd); err == nil {
			return name
		}
	}
	return "vi"
}

func splitEditorArgs(editor, filePath string) []string {
	if editor == "" {
		return []string{filePath}
	}
	parts := strings.Fields(editor)
	return append(parts, filePath)
}
```

- [x] **Step 4: 运行测试，验证通过**

```bash
go test ./pkg/util/ -v
```
Expected: ALL PASS

- [x] **Step 5: 运行全量测试，确保不破坏已有功能**

```bash
go test ./...
```
Expected: ALL PASS

- [x] **Step 6: 提交**

```bash
git add pkg/util/editor.go pkg/util/editor_test.go
git commit -m "feat(util): implement EditFile for local editor editing"
```
