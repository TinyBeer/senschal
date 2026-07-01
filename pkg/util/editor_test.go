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
