package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupListDirTest 创建供 ListDir 函数测试用的目录结构：
//
//	tmpDir/
//	├── file1.txt
//	├── file2.go
//	├── file3.md
//	├── subdir/
//	│   ├── subfile1.txt
//	│   └── subfile2.go
//	└── subdir2/
//	    └── nested/
//	        └── nested_file.txt
func setupListDirTest(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("b"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file3.md"), []byte("c"), 0o644))

	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "subfile1.txt"), []byte("d"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "subfile2.go"), []byte("e"), 0o644))

	nestedDir := filepath.Join(tmpDir, "subdir2")
	require.NoError(t, os.Mkdir(nestedDir, 0o755))
	nestedSub := filepath.Join(nestedDir, "nested")
	require.NoError(t, os.Mkdir(nestedSub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(nestedSub, "nested_file.txt"), []byte("f"), 0o644))

	return tmpDir
}

// ---------------------------------------------------------------------------
// ListDirOneLevelFile
// ---------------------------------------------------------------------------

func TestListDirOneLevelFile_AllFiles(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDirOneLevelFile(tmpDir)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"file1.txt", "file2.go", "file3.md"}, files)
}

func TestListDirOneLevelFile_FilterByExt(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDirOneLevelFile(tmpDir, ".txt")
	require.NoError(t, err)

	assert.Equal(t, []string{"file1.txt"}, files)
}

func TestListDirOneLevelFile_MultipleExts(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDirOneLevelFile(tmpDir, ".txt", ".go")
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"file1.txt", "file2.go"}, files)
}

func TestListDirOneLevelFile_ExtWithoutLeadingDot(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDirOneLevelFile(tmpDir, "txt")
	require.NoError(t, err)

	assert.Equal(t, []string{"file1.txt"}, files)
}

func TestListDirOneLevelFile_NoMatch(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDirOneLevelFile(tmpDir, ".xyz")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListDirOneLevelFile_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	files, err := ListDirOneLevelFile(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListDirOneLevelFile_NonExistentDir(t *testing.T) {
	_, err := ListDirOneLevelFile("/no/such/dir/12345")
	assert.Error(t, err)
}

func TestListDirOneLevelFile_PathIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a_regular_file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o644))

	_, err := ListDirOneLevelFile(filePath)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListDirOneLevelDir
// ---------------------------------------------------------------------------

func TestListDirOneLevelDir_AllDirs(t *testing.T) {
	tmpDir := setupListDirTest(t)

	dirs, err := ListDirOneLevelDir(tmpDir)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"subdir", "subdir2"}, dirs)
}

func TestListDirOneLevelDir_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	dirs, err := ListDirOneLevelDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, dirs)
}

func TestListDirOneLevelDir_NoDirsOnlyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "only_file.txt"), []byte("x"), 0o644))

	dirs, err := ListDirOneLevelDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, dirs)
}

func TestListDirOneLevelDir_NonExistentDir(t *testing.T) {
	_, err := ListDirOneLevelDir("/no/such/dir/12345")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListDilrRecursiveFile  (保留原始拼写）
// ---------------------------------------------------------------------------

func TestListDilrRecursiveFile_AllFiles(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDilrRecursiveFile(tmpDir)
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{
			"file1.txt",
			"file2.go",
			"file3.md",
			filepath.Join("subdir", "subfile1.txt"),
			filepath.Join("subdir", "subfile2.go"),
			filepath.Join("subdir2", "nested", "nested_file.txt"),
		},
		files,
	)
}

func TestListDilrRecursiveFile_FilterByExt(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDilrRecursiveFile(tmpDir, ".go")
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{
			"file2.go",
			filepath.Join("subdir", "subfile2.go"),
		},
		files,
	)
}

func TestListDilrRecursiveFile_MultipleExts(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDilrRecursiveFile(tmpDir, ".txt", ".go")
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{
			"file1.txt",
			"file2.go",
			filepath.Join("subdir", "subfile1.txt"),
			filepath.Join("subdir", "subfile2.go"),
			filepath.Join("subdir2", "nested", "nested_file.txt"),
		},
		files,
	)
}

func TestListDilrRecursiveFile_ExtWithoutLeadingDot(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDilrRecursiveFile(tmpDir, "txt")
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{
			"file1.txt",
			filepath.Join("subdir", "subfile1.txt"),
			filepath.Join("subdir2", "nested", "nested_file.txt"),
		},
		files,
	)
}

func TestListDilrRecursiveFile_NoMatch(t *testing.T) {
	tmpDir := setupListDirTest(t)

	files, err := ListDilrRecursiveFile(tmpDir, ".xyz")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListDilrRecursiveFile_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	files, err := ListDilrRecursiveFile(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListDilrRecursiveFile_NonExistentDir(t *testing.T) {
	_, err := ListDilrRecursiveFile("/no/such/dir/12345")
	assert.Error(t, err)
}

func TestListDilrRecursiveFile_PathIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a_regular_file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o644))

	_, err := ListDilrRecursiveFile(filePath)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListDirRecursiveDir
// ---------------------------------------------------------------------------

func TestListDirRecursiveDir_AllDirs(t *testing.T) {
	tmpDir := setupListDirTest(t)

	dirs, err := ListDirRecursiveDir(tmpDir)
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{
			"subdir",
			"subdir2",
			filepath.Join("subdir2", "nested"),
		},
		dirs,
	)
}

func TestListDirRecursiveDir_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	dirs, err := ListDirRecursiveDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, dirs)
}

func TestListDirRecursiveDir_NoDirsOnlyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "only_file.txt"), []byte("x"), 0o644))

	dirs, err := ListDirRecursiveDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, dirs)
}

func TestListDirRecursiveDir_NonExistentDir(t *testing.T) {
	_, err := ListDirRecursiveDir("/no/such/dir/12345")
	assert.Error(t, err)
}

func TestListDirRecursiveDir_PathIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a_regular_file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o644))

	_, err := ListDirRecursiveDir(filePath)
	assert.Error(t, err)
}
