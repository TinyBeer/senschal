package util

import (
	"os"
	"path/filepath"
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
