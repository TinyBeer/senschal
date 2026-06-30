package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImage_Name(t *testing.T) {
	tests := []struct {
		name  string
		image Image
		want  string
	}{
		{name: "simple image with tag", image: Image("nginx:latest"), want: "nginx_latest.tar"},
		{name: "image with registry and tag", image: Image("registry/my-image:1.0"), want: "registry-my-image_1.0.tar"},
		{name: "image without tag", image: Image("ubuntu"), want: "ubuntu.tar"},
		{name: "full registry path", image: Image("docker.io/library/alpine:3.18"), want: "docker.io-library-alpine_3.18.tar"},
		{name: "image with port in registry", image: Image("localhost:5000/myapp:latest"), want: "localhost_5000-myapp_latest.tar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.image.Name()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestImage_LocalFilePath(t *testing.T) {
	image := Image("nginx:latest")
	got := image.LocalFilePath()
	require.Contains(t, got, "nginx_latest.tar")
	require.True(t, filepath.IsAbs(got))
}

func TestImage_LocalFileExist(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		dir := t.TempDir()
		orig := DockerImageDir
		DockerImageDir = dir
		defer func() { DockerImageDir = orig }()

		err := os.WriteFile(filepath.Join(dir, "nginx_latest.tar"), []byte("data"), 0o644)
		require.NoError(t, err)

		image := Image("nginx:latest")
		require.True(t, image.LocalFileExist())
	})

	t.Run("file not exists", func(t *testing.T) {
		dir := t.TempDir()
		orig := DockerImageDir
		DockerImageDir = dir
		defer func() { DockerImageDir = orig }()

		image := Image("nginx:latest")
		require.False(t, image.LocalFileExist())
	})
}

func Test_readEnvConfigFromToml(t *testing.T) {
	dir := t.TempDir()
	content := `
[docker]
enable = true
version = "24.0"
check_user_group = true
image_list = ["nginx:latest", "redis:7"]
`
	err := os.WriteFile(filepath.Join(dir, "prod.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := readEnvConfigFromToml(dir, "prod")
	require.NoError(t, err)
	require.True(t, cfg.Docker.Enable)
	require.Equal(t, "24.0", cfg.Docker.Version)
	require.True(t, cfg.Docker.CheckUserGroup)
	require.Len(t, cfg.Docker.ImageList, 2)
	require.Equal(t, Image("nginx:latest"), cfg.Docker.ImageList[0])
	require.Equal(t, Image("redis:7"), cfg.Docker.ImageList[1])
}

func Test_readEnvConfigFromToml_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := readEnvConfigFromToml(dir, "nonexistent")
	require.Error(t, err)
}

func Test_getEnvConfigMap(t *testing.T) {
	dir := t.TempDir()

	content := `
alias = "prod-env"
[docker]
enable = true
`
	err := os.WriteFile(filepath.Join(dir, "production.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := getEnvConfigMap(dir)
	require.NoError(t, err)
	require.Len(t, m, 1)

	_, hasFileKey := m["production"]
	require.False(t, hasFileKey, "should not use filename as key when Alias is set")

	_, hasAliasKey := m["prod-env"]
	require.True(t, hasAliasKey, "should use Alias as key")
}

func Test_getEnvConfigMap_DuplicateAlias(t *testing.T) {
	dir := t.TempDir()

	content1 := `
alias = "shared"
[docker]
enable = true
`
	content2 := `
alias = "shared"
[docker]
enable = false
`
	err := os.WriteFile(filepath.Join(dir, "config1.toml"), []byte(content1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "config2.toml"), []byte(content2), 0o644)
	require.NoError(t, err)

	_, err = getEnvConfigMap(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicated env alias")
}

func Test_getEnvConfigMap_NoFiles(t *testing.T) {
	dir := t.TempDir()
	m, err := getEnvConfigMap(dir)
	require.NoError(t, err)
	require.Empty(t, m)
}

func Test_getEnvConfigMap_FilenameAsAlias(t *testing.T) {
	dir := t.TempDir()

	content := `[docker]
enable = true
`
	err := os.WriteFile(filepath.Join(dir, "default.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := getEnvConfigMap(dir)
	require.NoError(t, err)
	require.Len(t, m, 1)

	cfg, ok := m["default"]
	require.True(t, ok, "should use filename as key when Alias is empty")
	require.True(t, cfg.Docker.Enable)
}

func Test_getEnvConfigMap_ReadError(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad.toml"), []byte("[[[invalid toml"), 0o644)
	require.NoError(t, err)

	_, err = getEnvConfigMap(dir)
	require.Error(t, err)
}

func TestGetEnvConfigMap(t *testing.T) {
	dir := t.TempDir()
	orig := EnvConfigDir
	EnvConfigDir = dir
	defer func() { EnvConfigDir = orig }()

	content := `
[docker]
enable = true
version = "24.0"
`
	err := os.WriteFile(filepath.Join(dir, "prod.toml"), []byte(content), 0o644)
	require.NoError(t, err)

	m, err := GetEnvConfigMap()
	require.NoError(t, err)
	require.Len(t, m, 1)
	cfg, ok := m["prod"]
	require.True(t, ok)
	require.True(t, cfg.Docker.Enable)
}
