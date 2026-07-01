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

// editorPriority 是 PATH 中查找编辑器的顺序（优先靠前）
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
		if errors.Is(err, exec.ErrNotFound) || os.IsNotExist(err) {
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

func splitEditorArgs(editor, filePath string) []string {
	if editor == "" {
		return []string{filePath}
	}
	parts := strings.Fields(editor)
	return append(parts, filePath)
}
