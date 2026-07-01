package util

import (
	"os"
	"os/exec"
	"strings"
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

func splitEditorArgs(editor, filePath string) []string {
	if editor == "" {
		return []string{filePath}
	}
	parts := strings.Fields(editor)
	return append(parts, filePath)
}
