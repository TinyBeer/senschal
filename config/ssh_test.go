package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSSHConfig(t *testing.T) {
	cfg := NewSSHConfig()
	require.NotNil(t, cfg)
	require.Nil(t, cfg.SSH)
}

func TestSSHAuthMethod(t *testing.T) {
	require.Equal(t, SSHAuthMethod("password"), SSHAuthMethod_PW)
	require.Equal(t, SSHAuthMethod("key"), SSHAuthMethod_KEY)
}

func Test_readSSHConfigFromToml_Defaults(t *testing.T) {
	dir := t.TempDir()
	content := `
[ssh]
user = "testuser"
host = "192.168.1.1"
password = "secret"
`
	err := os.WriteFile(filepath.Join(dir, "test.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readSSHConfigFromToml(dir, "test")
	require.NoError(t, err)
	require.NotNil(t, cfg.SSH)
	require.Equal(t, "testuser", cfg.SSH.User)
	require.Equal(t, "192.168.1.1", cfg.SSH.Host)
	require.Equal(t, "secret", cfg.SSH.Password)
	require.Equal(t, SSHAuthMethod_PW, cfg.SSH.Method)
	require.Equal(t, 22, cfg.SSH.Port)
}

func Test_readSSHConfigFromToml_ExplicitValues(t *testing.T) {
	dir := t.TempDir()
	content := `
[ssh]
user = "root"
host = "10.0.0.1"
port = 2222
method = "key"
private_key = "/path/to/key"
`
	err := os.WriteFile(filepath.Join(dir, "explicit.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readSSHConfigFromToml(dir, "explicit")
	require.NoError(t, err)
	require.Equal(t, "root", cfg.SSH.User)
	require.Equal(t, "10.0.0.1", cfg.SSH.Host)
	require.Equal(t, 2222, cfg.SSH.Port)
	require.Equal(t, SSHAuthMethod_KEY, cfg.SSH.Method)
	require.Equal(t, "/path/to/key", cfg.SSH.PrivateKey)
}

func Test_readSSHConfigFromToml_NoSSHSection(t *testing.T) {
	dir := t.TempDir()
	content := `alias = "test-alias"`
	err := os.WriteFile(filepath.Join(dir, "test.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readSSHConfigFromToml(dir, "test")
	require.NoError(t, err)
	require.Nil(t, cfg.SSH)
	require.Equal(t, "test-alias", cfg.Alias)
}

func Test_readSSHConfigFromToml_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := readSSHConfigFromToml(dir, "nonexistent")
	require.Error(t, err)
}

func Test_getSSHConfigMap(t *testing.T) {
	dir := t.TempDir()

	content1 := `[ssh]
user = "user1"
host = "host1"
`
	err := os.WriteFile(filepath.Join(dir, "node1.toml"), []byte(content1), 0o644)
	require.NoError(t, err)

	content2 := `
alias = "server2"
[ssh]
user = "user2"
host = "host2"
`
	err = os.WriteFile(filepath.Join(dir, "node2.toml"), []byte(content2), 0o644)
	require.NoError(t, err)

	m, err := getSSHConfigMap(dir)
	require.NoError(t, err)
	require.Len(t, m, 2)

	cfg1, ok := m["node1"]
	require.True(t, ok, "expected key 'node1'")
	require.Equal(t, "user1", cfg1.SSH.User)

	cfg2, ok := m["server2"]
	require.True(t, ok, "expected key 'server2'")
	require.Equal(t, "user2", cfg2.SSH.User)
}

func Test_getSSHConfigMap_NoFiles(t *testing.T) {
	dir := t.TempDir()
	m, err := getSSHConfigMap(dir)
	require.NoError(t, err)
	require.Empty(t, m)
}

func Test_getSSHConfigMap_ReadError(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad.toml"), []byte("[[[invalid toml"), 0o644)
	require.NoError(t, err)

	_, err = getSSHConfigMap(dir)
	require.Error(t, err)
}

func TestGetSSHConfigMap(t *testing.T) {
	dir := t.TempDir()
	orig := SSHConfigDir
	SSHConfigDir = dir
	defer func() { SSHConfigDir = orig }()

	content := `[ssh]
user = "testuser"
host = "10.0.0.1"
`
	err := os.WriteFile(filepath.Join(dir, "server.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := GetSSHConfigMap()
	require.NoError(t, err)
	require.Len(t, m, 1)
	cfg, ok := m["server"]
	require.True(t, ok)
	require.Equal(t, "testuser", cfg.SSH.User)
}

func TestWriteSSHConfig_Nil(t *testing.T) {
	err := WriteSSHConfig(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ssh config is nil")
}

func TestWriteSSHConfig_NilSSH(t *testing.T) {
	err := WriteSSHConfig(&SSHConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ssh config is nil")
}

func TestWriteSSHConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	orig := SSHConfigDir
	SSHConfigDir = dir
	defer func() { SSHConfigDir = orig }()

	cfg := &SSHConfig{
		Alias: "test-roundtrip",
		SSH: &SSH{
			User:     "deploy",
			Host:     "192.168.1.100",
			Port:     22,
			Method:   SSHAuthMethod_PW,
			Password: "s3cret",
		},
	}

	err := WriteSSHConfig(cfg)
	require.NoError(t, err)

	m, err := GetSSHConfigMap()
	require.NoError(t, err)
	readCfg, ok := m["test-roundtrip"]
	require.True(t, ok)
	require.Equal(t, "deploy", readCfg.SSH.User)
	require.Equal(t, "192.168.1.100", readCfg.SSH.Host)
}
