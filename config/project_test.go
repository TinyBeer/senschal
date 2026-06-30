package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectConfig_GetProtoDir(t *testing.T) {
	cfg := &ProjectConfig{ProjectDir: "/home/user/project", ProtoDir: "proto"}
	got := cfg.GetProtoDir()
	require.Equal(t, "/home/user/project/proto", got)
}

func TestProjectConfig_GetLobbyRegisterFile(t *testing.T) {
	cfg := &ProjectConfig{ProjectDir: "/home/user/project", LobbyRegisterFile: "register/lobby.go"}
	got := cfg.GetLobbyRegisterFile()
	require.Equal(t, "/home/user/project/register/lobby.go", got)
}

func TestProjectConfig_GetServiceDir(t *testing.T) {
	cfg := &ProjectConfig{ProjectDir: "/home/user/project", ServiceDir: "service"}
	got := cfg.GetServiceDir()
	require.Equal(t, "/home/user/project/service", got)
}

func TestNewProjectConfig(t *testing.T) {
	cfg := NewProjectConfig()
	require.NotNil(t, cfg)
	require.Empty(t, cfg.Alias)
}

func Test_readProjectConfigFromToml(t *testing.T) {
	dir := t.TempDir()
	content := `
alias = "myapp"
project_dir = "/home/user/projects/myapp"
proto_dir = "api/proto"
service_dir = "internal/service"
lobby_register_with_tool = true
lobby_register_file = "register/lobby.go"
service_message_template = "service.tmpl"
`
	err := os.WriteFile(filepath.Join(dir, "myapp.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readProjectConfigFromToml(dir, "myapp")
	require.NoError(t, err)
	require.Equal(t, "myapp", cfg.Alias)
	require.Equal(t, "/home/user/projects/myapp", cfg.ProjectDir)
	require.Equal(t, "api/proto", cfg.ProtoDir)
	require.Equal(t, "internal/service", cfg.ServiceDir)
	require.True(t, cfg.LobbyRegisterWithTool)
	require.Equal(t, "register/lobby.go", cfg.LobbyRegisterFile)
	require.Equal(t, "service.tmpl", cfg.ServiceMessageTemplate)
}

func Test_readProjectConfigFromToml_Minimal(t *testing.T) {
	dir := t.TempDir()
	content := `alias = "minimal"`
	err := os.WriteFile(filepath.Join(dir, "minimal.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readProjectConfigFromToml(dir, "minimal")
	require.NoError(t, err)
	require.Equal(t, "minimal", cfg.Alias)
	require.Empty(t, cfg.ProjectDir)
	require.False(t, cfg.LobbyRegisterWithTool)
}

func Test_readProjectConfigFromToml_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := readProjectConfigFromToml(dir, "nonexistent")
	require.Error(t, err)
}
