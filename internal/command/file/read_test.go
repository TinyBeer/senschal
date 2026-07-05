package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileContain_Contains(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(fPath, []byte("hello world\nfoo bar"), 0o644))

	ok, err := FileContain(fPath, "foo")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestFileContain_NotContains(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(fPath, []byte("hello world"), 0o644))

	ok, err := FileContain(fPath, "zzz")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestFileContain_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "empty.txt")
	require.NoError(t, os.WriteFile(fPath, []byte(""), 0o644))

	ok, err := FileContain(fPath, "")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestFileContain_FileNotExist(t *testing.T) {
	_, err := FileContain("/no/such/file.txt", "x")
	assert.Error(t, err)
}
