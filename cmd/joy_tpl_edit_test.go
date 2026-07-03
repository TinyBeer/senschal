package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"seneschal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoyTplEdit_NoArgs(t *testing.T) {
	// cobra.ExactArgs(1) should reject 0 args
	err := joyTplEditCmd.Args(joyTplEditCmd, []string{})
	assert.Error(t, err)
}

func TestJoyTplEdit_TemplateNotFound(t *testing.T) {
	orig := config.TplDir
	config.TplDir = t.TempDir()
	defer func() { config.TplDir = orig }()

	err := joyTplEditCmd.RunE(joyTplEditCmd, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestJoyTplEdit_SettingFileNotFound(t *testing.T) {
	orig := config.TplDir
	config.TplDir = t.TempDir()
	defer func() { config.TplDir = orig }()

	// Create template directory but no setting.toml
	tplDir := filepath.Join(config.TplDir, "mytpl")
	require.NoError(t, os.MkdirAll(tplDir, 0o755))
	// The template system needs the template/ subdir to look valid
	require.NoError(t, os.MkdirAll(filepath.Join(tplDir, "template"), 0o755))

	err := joyTplEditCmd.RunE(joyTplEditCmd, []string{"mytpl"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "setting file")
}

func TestJoyTplEdit_SuccessWithMockEditor(t *testing.T) {
	orig := config.TplDir
	config.TplDir = t.TempDir()
	defer func() { config.TplDir = orig }()

	// Create template directory with setting.toml
	tplDir := filepath.Join(config.TplDir, "mytpl")
	require.NoError(t, os.MkdirAll(tplDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tplDir, "template"), 0o755))

	settingContent := `__alias__ = "mytpl"
__description__ = "test template"
image = "ubuntu"
`
	settingPath := filepath.Join(tplDir, config.TplSettingName+".toml")
	require.NoError(t, os.WriteFile(settingPath, []byte(settingContent), 0o644))

	// Create mock editor that appends a line
	mockEditor := filepath.Join(t.TempDir(), "mock_editor.sh")
	mockScript := `#!/bin/sh
printf '%s\nnew_param = "added"' "$(cat "$1")" > "$1"
`
	require.NoError(t, os.WriteFile(mockEditor, []byte(mockScript), 0o755))

	t.Setenv("VISUAL", mockEditor)
	t.Setenv("EDITOR", "")

	err := joyTplEditCmd.RunE(joyTplEditCmd, []string{"mytpl"})
	require.NoError(t, err)

	// Verify the setting file was edited
	content, err := os.ReadFile(settingPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `new_param = "added"`)
}

func TestJoyTplEdit_ExistingSettingHasAlias(t *testing.T) {
	// Verify that templates with __alias__ can be found by alias too
	orig := config.TplDir
	config.TplDir = t.TempDir()
	defer func() { config.TplDir = orig }()

	tplDir := filepath.Join(config.TplDir, "my-template-dir")
	require.NoError(t, os.MkdirAll(tplDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tplDir, "template"), 0o755))

	settingContent := `__alias__ = "mytpl"
__description__ = "test template"
`
	settingPath := filepath.Join(tplDir, config.TplSettingName+".toml")
	require.NoError(t, os.WriteFile(settingPath, []byte(settingContent), 0o644))

	mockEditor := filepath.Join(t.TempDir(), "mock_editor.sh")
	mockScript := "#!/bin/sh\necho 'edited' > \"$1\"\n"
	require.NoError(t, os.WriteFile(mockEditor, []byte(mockScript), 0o755))

	t.Setenv("VISUAL", mockEditor)
	t.Setenv("EDITOR", "")

	// Find by __alias__ "mytpl" not directory name "my-template-dir"
	err := joyTplEditCmd.RunE(joyTplEditCmd, []string{"mytpl"})
	require.NoError(t, err)
}
