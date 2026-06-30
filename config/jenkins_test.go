package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_readJenkinsConfigFromToml(t *testing.T) {
	dir := t.TempDir()
	content := `
alias = "myjenkins"
host = "http://jenkins:8080"
user_name = "admin"
password = "token123"
`
	err := os.WriteFile(filepath.Join(dir, "jenkins1.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readJenkinsConfigFromToml(dir, "jenkins1")
	require.NoError(t, err)
	require.Equal(t, "myjenkins", cfg.Alias)
	require.Equal(t, "http://jenkins:8080", cfg.Host)
	require.Equal(t, "admin", cfg.UserName)
	require.Equal(t, "token123", cfg.Password)
}

func Test_readJenkinsConfigFromToml_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := readJenkinsConfigFromToml(dir, "nonexistent")
	require.Error(t, err)
}

func Test_getJenkinsConfigMap(t *testing.T) {
	dir := t.TempDir()

	content := `
alias = "ci"
host = "http://ci.example.com"
user_name = "bot"
password = "api-token"
`
	err := os.WriteFile(filepath.Join(dir, "main.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := getJenkinsConfigMap(dir)
	require.NoError(t, err)
	require.Len(t, m, 1)

	_, hasFileKey := m["main"]
	require.False(t, hasFileKey, "should not use filename when Alias is set")

	cfg, ok := m["ci"]
	require.True(t, ok)
	require.Equal(t, "http://ci.example.com", cfg.Host)
}

func Test_getJenkinsConfigMap_DuplicateAlias(t *testing.T) {
	dir := t.TempDir()

	content1 := `alias = "dup"`
	err := os.WriteFile(filepath.Join(dir, "a.toml"), []byte(content1), 0o644)
	require.NoError(t, err)

	content2 := `alias = "dup"`
	err = os.WriteFile(filepath.Join(dir, "b.toml"), []byte(content2), 0o644)
	require.NoError(t, err)

	_, err = getJenkinsConfigMap(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicated jenkins alias")
}

func Test_getJenkinsConfigMap_NoFiles(t *testing.T) {
	dir := t.TempDir()
	m, err := getJenkinsConfigMap(dir)
	require.NoError(t, err)
	require.Empty(t, m)
}

func Test_getJenkinsConfigMap_FilenameAsAlias(t *testing.T) {
	dir := t.TempDir()

	content := `host = "http://localhost:8080"`
	err := os.WriteFile(filepath.Join(dir, "local.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := getJenkinsConfigMap(dir)
	require.NoError(t, err)

	cfg, ok := m["local"]
	require.True(t, ok, "should use filename as key when Alias is empty")
	require.Equal(t, "http://localhost:8080", cfg.Host)
}

func Test_getJenkinsConfigMap_ReadError(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad.toml"), []byte("[[[invalid toml"), 0o644)
	require.NoError(t, err)

	_, err = getJenkinsConfigMap(dir)
	require.Error(t, err)
}

func TestGetJenkinsConfigMap(t *testing.T) {
	dir := t.TempDir()
	orig := JenkinsConfigDir
	JenkinsConfigDir = dir
	defer func() { JenkinsConfigDir = orig }()

	content := `
alias = "ci"
host = "http://jenkins:8080"
user_name = "admin"
password = "token"
`
	err := os.WriteFile(filepath.Join(dir, "main.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := GetJenkinsConfigMap()
	require.NoError(t, err)
	require.Len(t, m, 1)
	cfg, ok := m["ci"]
	require.True(t, ok)
	require.Equal(t, "http://jenkins:8080", cfg.Host)
}

func TestWriteJenkinsConfig_MissingUser(t *testing.T) {
	err := WriteJenkinsConfig(&Jenkins{Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing user name or password")
}

func TestWriteJenkinsConfig_MissingPassword(t *testing.T) {
	err := WriteJenkinsConfig(&Jenkins{UserName: "admin"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing user name or password")
}

func TestWriteJenkinsConfig_MissingBoth(t *testing.T) {
	err := WriteJenkinsConfig(&Jenkins{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing user name or password")
}

func TestWriteJenkinsConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	orig := JenkinsConfigDir
	JenkinsConfigDir = dir
	defer func() { JenkinsConfigDir = orig }()

	cfg := &Jenkins{
		Alias:    "my-ci",
		Host:     "http://jenkins:8080",
		UserName: "bot",
		Password: "api-secret",
	}

	err := WriteJenkinsConfig(cfg)
	require.NoError(t, err)

	m, err := GetJenkinsConfigMap()
	require.NoError(t, err)
	readCfg, ok := m["my-ci"]
	require.True(t, ok)
	require.Equal(t, "bot", readCfg.UserName)
	require.Equal(t, "api-secret", readCfg.Password)
}
