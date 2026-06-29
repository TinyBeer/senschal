package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultGenerateDir   = "build"
	ConfigDirName        = "conf"
	EnvConfigDirName     = "env"
	DockerImageDirName   = "docker_image"
	DockerDebDirName     = "docker_deb"
	SSHConfigDirName     = "ssh"
	SSHKeyDirName        = "ssh_key"
	WorkoutDirName       = "workout"
	ProjectDirName       = "project"
	TodoDirName          = "todo"
	TplDirName           = "tpl"
	TplGenDirName        = "_gen"
	TplTemplateDirName   = "template"
	TplSettingName       = "setting"
	JenkinsConfigDirName = "jenkins"
)

var DefaultDataDir = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("cannot get user home directory: %v", err))
	}
	return filepath.Join(home, ".seneschal")
}()

var (
	ConfigDir        = filepath.Join(DefaultDataDir, ConfigDirName)
	EnvConfigDir     = filepath.Join(ConfigDir, EnvConfigDirName)
	DockerDebDir     = filepath.Join(DefaultDataDir, DockerDebDirName)
	DockerImageDir   = filepath.Join(DefaultDataDir, DockerImageDirName)
	SSHConfigDir     = filepath.Join(ConfigDir, SSHConfigDirName)
	SSHKeyDir        = filepath.Join(SSHConfigDir, SSHKeyDirName)
	WorkoutDir       = filepath.Join(ConfigDir, WorkoutDirName)
	ProjectDir       = filepath.Join(ConfigDir, ProjectDirName)
	TodoDir          = filepath.Join(DefaultDataDir, TodoDirName)
	TplDir           = filepath.Join(DefaultDataDir, TplDirName)
	JenkinsConfigDir = filepath.Join(ConfigDir, JenkinsConfigDirName)
)
