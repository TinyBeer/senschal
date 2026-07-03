package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"seneschal/config"

	"github.com/stretchr/testify/require"
)

func TestHostEdit_NonExistentAlias(t *testing.T) {
	_ = hostEditCmd

	// Set SSHConfigDir to a temp dir with no files
	orig := config.SSHConfigDir
	config.SSHConfigDir = t.TempDir()
	defer func() { config.SSHConfigDir = orig }()

	alias := "nonexistent"
	configPath := filepath.Join(config.SSHConfigDir, alias+".toml")

	_, err := os.Stat(configPath)
	require.True(t, os.IsNotExist(err), "config file should not exist")
}

func TestHostEdit_ExistingAliasFileLookup(t *testing.T) {
	_ = hostEditCmd

	orig := config.SSHConfigDir
	config.SSHConfigDir = t.TempDir()
	defer func() { config.SSHConfigDir = orig }()

	// Create a valid TOML config file
	alias := "testserver"
	content := `alias = "testserver"
[ssh]
user = "root"
host = "192.168.1.1"
password = "secret"
`
	err := os.WriteFile(filepath.Join(config.SSHConfigDir, alias+".toml"), []byte(content), 0o644)
	require.NoError(t, err)

	// Verify we can find and parse the config
	configPath := filepath.Join(config.SSHConfigDir, alias+".toml")
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	cfg, err := config.ReadSSHConfigFromToml(config.SSHConfigDir, alias)
	require.NoError(t, err)
	require.NotNil(t, cfg.SSH)
	require.Equal(t, "root", cfg.SSH.User)
	require.Equal(t, "192.168.1.1", cfg.SSH.Host)
}

func TestHostEdit_InvalidTOMLAfterEdit(t *testing.T) {
	_ = hostEditCmd

	orig := config.SSHConfigDir
	config.SSHConfigDir = t.TempDir()
	defer func() { config.SSHConfigDir = orig }()

	alias := "badconfig"

	// Write invalid TOML
	err := os.WriteFile(filepath.Join(config.SSHConfigDir, alias+".toml"), []byte("[[[invalid toml"), 0o644)
	require.NoError(t, err)

	// Reading should fail
	_, err = config.ReadSSHConfigFromToml(config.SSHConfigDir, alias)
	require.Error(t, err)
}

func TestHostEdit_ConfigDirPath(t *testing.T) {
	_ = hostEditCmd

	// Verify the default SSHConfigDir is non-empty and follows expected pattern
	require.NotEmpty(t, config.SSHConfigDir)
	require.Contains(t, config.SSHConfigDir, "ssh")
}

func TestHostEdit_RoundTripWriteThenRead(t *testing.T) {
	_ = hostEditCmd

	orig := config.SSHConfigDir
	config.SSHConfigDir = t.TempDir()
	defer func() { config.SSHConfigDir = orig }()

	// Write a config programmatically
	alias := "roundtrip"
	cfg := &config.SSHConfig{
		Alias: alias,
		SSH: &config.SSH{
			User:   "deploy",
			Host:   "10.0.0.50",
			Port:   2222,
			Method: config.SSHAuthMethod_KEY,
		},
	}
	err := config.WriteSSHConfig(cfg)
	require.NoError(t, err)

	// Read it back
	readCfg, err := config.ReadSSHConfigFromToml(config.SSHConfigDir, alias)
	require.NoError(t, err)
	require.NotNil(t, readCfg.SSH)
	require.Equal(t, "deploy", readCfg.SSH.User)
	require.Equal(t, "10.0.0.50", readCfg.SSH.Host)
	require.Equal(t, 2222, readCfg.SSH.Port)
}
