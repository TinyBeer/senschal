package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTemplateTest creates a template directory structure for testing.
//
//	tmpDir/
//	├── setting.toml
//	├── template/
//	│   ├── greeting.txt.tpl        (Go template file)
//	│   └── subdir/
//	│       └── info.txt.tpl
func setupTemplateTest(t *testing.T) (tplPath, genDir, settingPath string) {
	t.Helper()
	tmpDir := t.TempDir()

	// setting.toml
	settingPath = filepath.Join(tmpDir, "setting.toml")
	settingContent := `name = "World"
count = 3
`
	require.NoError(t, os.WriteFile(settingPath, []byte(settingContent), 0o644))

	// template dir
	tplPath = tmpDir
	tplSrcDir := filepath.Join(tplPath, tplTemplateDirName)
	require.NoError(t, os.Mkdir(tplSrcDir, 0o755))

	// greeting.txt.tpl
	greetingContent := `Hello, {{ .name }}! Count: {{ .count }}
`
	require.NoError(t, os.WriteFile(filepath.Join(tplSrcDir, "greeting.txt.tpl"), []byte(greetingContent), 0o644))

	// subdir/info.txt.tpl
	subDir := filepath.Join(tplSrcDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0o755))
	infoContent := `Info: {{ .name }}
`
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "info.txt.tpl"), []byte(infoContent), 0o644))

	// gen dir
	genDir = filepath.Join(tmpDir, "_gen")
	return
}

func TestExecuteTemplate_Success(t *testing.T) {
	tplPath, genDir, settingPath := setupTemplateTest(t)

	err := ExecuteTemplate(tplPath, genDir, settingPath)
	require.NoError(t, err)

	// Verify generated files
	greetingGen := filepath.Join(genDir, "greeting.txt.tpl")
	content, err := os.ReadFile(greetingGen)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World! Count: 3\n", string(content))

	infoGen := filepath.Join(genDir, "subdir", "info.txt.tpl")
	content, err = os.ReadFile(infoGen)
	require.NoError(t, err)
	assert.Equal(t, "Info: World\n", string(content))
}

func TestExecuteTemplate_TplPathNotDir(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not_a_dir.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0o644))

	settingPath := filepath.Join(tmpDir, "setting.toml")
	require.NoError(t, os.WriteFile(settingPath, []byte("key = 1"), 0o644))

	err := ExecuteTemplate(filePath, t.TempDir(), settingPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not a dir")
}

func TestExecuteTemplate_SettingNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	tplSrcDir := filepath.Join(tmpDir, tplTemplateDirName)
	require.NoError(t, os.Mkdir(tplSrcDir, 0o755))

	err := ExecuteTemplate(tmpDir, t.TempDir(), "/no/such/setting.toml")
	assert.Error(t, err)
}

func TestExecuteTemplate_NoTemplateSrcDir(t *testing.T) {
	tmpDir := t.TempDir()

	settingPath := filepath.Join(tmpDir, "setting.toml")
	require.NoError(t, os.WriteFile(settingPath, []byte("key = 1"), 0o644))

	// tplPath exists but has no "template" subdir
	err := ExecuteTemplate(tmpDir, t.TempDir(), settingPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template source dir")
}

func TestExecuteTemplate_EmptyTemplateDir(t *testing.T) {
	tmpDir := t.TempDir()

	settingPath := filepath.Join(tmpDir, "setting.toml")
	require.NoError(t, os.WriteFile(settingPath, []byte("key = 1"), 0o644))

	tplSrcDir := filepath.Join(tmpDir, tplTemplateDirName)
	require.NoError(t, os.Mkdir(tplSrcDir, 0o755))

	err := ExecuteTemplate(tmpDir, t.TempDir(), settingPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nothing found")
}
