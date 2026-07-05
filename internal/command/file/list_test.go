package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupListTest creates a temp directory structure for testing list functions:
//
//	tmpDir/
//	├── a.txt
//	├── b.go
//	├── c.toml
//	├── subdir/
//	│   ├── sub_a.txt
//	│   └── sub_b.go
//	└── empty/
func setupListTest(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "b.go"), []byte("b"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "c.toml"), []byte("c"), 0o644))
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub_a.txt"), []byte("d"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub_b.go"), []byte("e"), 0o644))
	emptyDir := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.Mkdir(emptyDir, 0o755))
	return tmpDir
}

// ---------------------------------------------------------------------------
// ListFileWithExt
// ---------------------------------------------------------------------------

func TestListFileWithExt_Found(t *testing.T) {
	tmpDir := setupListTest(t)
	files, err := ListFileWithExt(tmpDir, "txt")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join(tmpDir, "a.txt"),
		filepath.Join(tmpDir, "subdir", "sub_a.txt"),
	}, files)
}

func TestListFileWithExt_NoMatch(t *testing.T) {
	tmpDir := setupListTest(t)
	files, err := ListFileWithExt(tmpDir, "md")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListFileWithExt_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	files, err := ListFileWithExt(tmpDir, "txt")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListFileWithExt_NonExistentDir(t *testing.T) {
	_, err := ListFileWithExt("/no/such/dir", "txt")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListFileNameWithExt
// ---------------------------------------------------------------------------

func TestListFileNameWithExt_Found(t *testing.T) {
	tmpDir := setupListTest(t)
	names, err := ListFileNameWithExt(tmpDir, "txt")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "sub_a"}, names)
}

func TestListFileNameWithExt_NoMatch(t *testing.T) {
	tmpDir := setupListTest(t)
	names, err := ListFileNameWithExt(tmpDir, "md")
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestListFileNameWithExt_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	names, err := ListFileNameWithExt(tmpDir, "txt")
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestListFileNameWithExt_NonExistentDir(t *testing.T) {
	_, err := ListFileNameWithExt("/no/such/dir", "txt")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListDirName
// ---------------------------------------------------------------------------

func TestListDirName_Found(t *testing.T) {
	tmpDir := setupListTest(t)
	dirs, err := ListDirName(tmpDir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"subdir", "empty"}, dirs)
}

func TestListDirName_NoDirs(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "only_file.txt"), []byte("x"), 0o644))
	dirs, err := ListDirName(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, dirs)
}

func TestListDirName_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	dirs, err := ListDirName(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, dirs)
}

func TestListDirName_NonExistentDir(t *testing.T) {
	_, err := ListDirName("/no/such/dir")
	assert.Error(t, err)
}
