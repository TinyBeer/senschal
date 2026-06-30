package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_writeConfigToToml_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	cfg := &SSHConfig{
		Alias: "test-server",
		SSH: &SSH{
			User:       "admin",
			Host:       "192.168.1.1",
			Port:       2222,
			Method:     SSHAuthMethod_KEY,
			PrivateKey: "/home/admin/.ssh/id_rsa",
		},
	}

	err := writeConfigToToml(cfg, dir, cfg.Alias, "ssh")
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "test-server.toml"))
	require.NoError(t, err)

	readCfg, err := readSSHConfigFromToml(dir, "test-server")
	require.NoError(t, err)
	require.NotNil(t, readCfg.SSH)
	require.Equal(t, cfg.SSH.User, readCfg.SSH.User)
	require.Equal(t, cfg.SSH.Host, readCfg.SSH.Host)
	require.Equal(t, cfg.SSH.Port, readCfg.SSH.Port)
	require.Equal(t, cfg.SSH.Method, readCfg.SSH.Method)
}

func Test_writeConfigToToml_StripsEmptyStrings(t *testing.T) {
	dir := t.TempDir()

	cfg := &SSHConfig{
		Alias: "empty-fields",
		SSH: &SSH{
			User:       "test",
			Host:       "host",
			Password:   "",
			PrivateKey: "",
		},
	}

	err := writeConfigToToml(cfg, dir, cfg.Alias, "ssh")
	require.NoError(t, err)

	readCfg, err := readSSHConfigFromToml(dir, "empty-fields")
	require.NoError(t, err)
	require.Equal(t, "test", readCfg.SSH.User)
	require.Equal(t, "host", readCfg.SSH.Host)
}
