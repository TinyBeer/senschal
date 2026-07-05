package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"seneschal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerPullImageList_Empty(t *testing.T) {
	orig := ExecuteCommand
	ExecuteCommand = func(cmd string) ([]byte, error) {
		if cmd == `docker images --format "{{.Repository}}:{{.Tag}}"` {
			return []byte(""), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", cmd)
	}
	t.Cleanup(func() { ExecuteCommand = orig })

	err := DockerPullImageList(context.Background(), nil)
	require.NoError(t, err)

	err = DockerPullImageList(context.Background(), []config.Image{})
	require.NoError(t, err)
}

func TestDockerPullImageList_PullsMissingImages(t *testing.T) {
	orig := ExecuteCommand
	var commands []string
	ExecuteCommand = func(cmd string) ([]byte, error) {
		commands = append(commands, cmd)
		switch cmd {
		case `docker images --format "{{.Repository}}:{{.Tag}}"`:
			// 模拟本地已有 nginx，缺少 alpine
			return []byte("nginx:latest\n"), nil
		case `docker pull alpine:latest -q`:
			return []byte(""), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", cmd)
	}
	t.Cleanup(func() { ExecuteCommand = orig })

	err := DockerPullImageList(context.Background(), []config.Image{"nginx:latest", "alpine:latest"})
	require.NoError(t, err)

	// 应只 pull 缺少的 alpine
	assert.Contains(t, commands[len(commands)-1], "pull alpine")
}

func TestDockerPullImageList_CommandError(t *testing.T) {
	orig := ExecuteCommand
	ExecuteCommand = func(cmd string) ([]byte, error) {
		return nil, errors.New("docker not available")
	}
	t.Cleanup(func() { ExecuteCommand = orig })

	err := DockerPullImageList(context.Background(), []config.Image{"nginx:latest"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list docker images")
}

func TestDockerLoadImageList_Empty(t *testing.T) {
	orig := ExecuteCommand
	ExecuteCommand = func(cmd string) ([]byte, error) {
		if cmd == `docker images --format "{{.Repository}}:{{.Tag}}"` {
			return []byte(""), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", cmd)
	}
	t.Cleanup(func() { ExecuteCommand = orig })

	err := DockerLoadImageList(context.Background(), nil)
	require.NoError(t, err)
}

func TestDockerLoadImageList_LocalFileExists(t *testing.T) {
	orig := ExecuteCommand
	ExecuteCommand = func(cmd string) ([]byte, error) {
		if cmd == `docker images --format "{{.Repository}}:{{.Tag}}"` {
			return []byte("nginx:latest\n"), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", cmd)
	}
	t.Cleanup(func() { ExecuteCommand = orig })

	origDir := config.DockerImageDir
	config.DockerImageDir = t.TempDir()
	t.Cleanup(func() { config.DockerImageDir = origDir })

	image := config.Image("nginx:latest")
	err := os.WriteFile(filepath.Join(config.DockerImageDir, image.Name()), []byte("fake tar"), 0o644)
	require.NoError(t, err)
	assert.True(t, image.LocalFileExist())

	// 本地文件已存在，跳过 save
	err = DockerLoadImageList(context.Background(), []config.Image{image})
	require.NoError(t, err)
}

func TestDockerLoadImageList_NeedsSave(t *testing.T) {
	orig := ExecuteCommand
	cmds := make(map[string]bool)
	ExecuteCommand = func(cmd string) ([]byte, error) {
		cmds[cmd] = true
		switch cmd {
		case `docker images --format "{{.Repository}}:{{.Tag}}"`:
			return []byte(""), nil
		case `docker pull alpine:latest -q`:
			return []byte(""), nil
		case `docker save -o ` + config.Image("alpine:latest").LocalFilePath() + ` alpine:latest`:
			return []byte(""), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", cmd)
	}
	t.Cleanup(func() { ExecuteCommand = orig })

	origDir := config.DockerImageDir
	config.DockerImageDir = t.TempDir()
	t.Cleanup(func() { config.DockerImageDir = origDir })

	err := DockerLoadImageList(context.Background(), []config.Image{"alpine:latest"})
	require.NoError(t, err)
	assert.True(t, cmds["docker pull alpine:latest -q"], "should pull missing image")
}
