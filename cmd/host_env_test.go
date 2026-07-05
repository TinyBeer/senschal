package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"seneschal/config"

	"github.com/stretchr/testify/require"
)

func TestPrepareHostEnv(t *testing.T) {
	// 保存原始路径，测试后恢复
	origSSHDir := config.SSHConfigDir
	origEnvDir := config.EnvConfigDir
	t.Cleanup(func() {
		config.SSHConfigDir = origSSHDir
		config.EnvConfigDir = origEnvDir
	})

	sshDir := t.TempDir()
	envDir := t.TempDir()
	config.SSHConfigDir = sshDir
	config.EnvConfigDir = envDir

	// 创建 SSH 配置
	sshContent := `alias = "server1"
[ssh]
user = "root"
host = "10.0.0.1"
port = 22
method = "password"
password = "test"
`
	err := os.WriteFile(filepath.Join(sshDir, "server1.toml"), []byte(sshContent), 0o644)
	require.NoError(t, err)

	// 创建环境配置（Docker 启用但无详细配置，以便 prepareHostEnv 正常执行）
	envContent := `alias = "test-env"
`
	err = os.WriteFile(filepath.Join(envDir, "test-env.toml"), []byte(envContent), 0o644)
	require.NoError(t, err)

	scm, aliases, envMgrs, err := prepareHostEnv(context.Background(), []string{"server1", "test-env"})
	require.NoError(t, err)
	require.NotNil(t, scm)
	require.Contains(t, scm, "server1")
	require.Equal(t, []string{"server1"}, aliases)
	require.Len(t, envMgrs, 1)
	// Docker config 不存在时 envMgr 为 nil
	require.Nil(t, envMgrs[0])
}

func TestPrepareHostEnv_MissingSSHAlias(t *testing.T) {
	origSSHDir := config.SSHConfigDir
	origEnvDir := config.EnvConfigDir
	t.Cleanup(func() {
		config.SSHConfigDir = origSSHDir
		config.EnvConfigDir = origEnvDir
	})
	sshDir := t.TempDir()
	envDir := t.TempDir()
	config.SSHConfigDir = sshDir
	config.EnvConfigDir = envDir

	// 创建 SSH 配置 - 只有一个 server1
	sshContent := `alias = "server1"
[ssh]
user = "root"
host = "10.0.0.1"
`
	err := os.WriteFile(filepath.Join(sshDir, "server1.toml"), []byte(sshContent), 0o644)
	require.NoError(t, err)

	envContent := `alias = "test-env"
`
	err = os.WriteFile(filepath.Join(envDir, "test-env.toml"), []byte(envContent), 0o644)
	require.NoError(t, err)

	// args中包含 server2（不存在于 SSH 配置中）
	_, _, _, err = prepareHostEnv(context.Background(), []string{"server1,server2", "test-env"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "未找到的机器配置")
}

func TestPrepareHostEnv_MissingEnvAlias(t *testing.T) {
	origSSHDir := config.SSHConfigDir
	origEnvDir := config.EnvConfigDir
	t.Cleanup(func() {
		config.SSHConfigDir = origSSHDir
		config.EnvConfigDir = origEnvDir
	})
	config.SSHConfigDir = t.TempDir()
	config.EnvConfigDir = t.TempDir()

	// 不创建任何文件，环境配置不存在
	_, _, _, err := prepareHostEnv(context.Background(), []string{"server1", "nonexistent-env"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "未找到环境")
}
