package config

import (
	"path/filepath"
	"seneschal/tool/file"
	"sort"

	"github.com/spf13/viper"
)

type ProjectConfig struct {
	Alias                 string `mapstructure:"alias"`
	ProjectDir            string `mapstructure:"project_dir"` // absolute path
	ProtoDir              string `mapstructure:"proto_dir"`   // relative path
	ServiceDir            string `mapstructure:"service_dir"` // relative path
	LobbyRegisterWithTool bool   `mapstructure:"lobby_register_with_tool"`
	LobbyRegisterFile     string `mapstructure:"lobby_register_file"` // relative path
}

func (c *ProjectConfig) GetProtoDir() string {
	return filepath.Join(c.ProjectDir, c.ProtoDir)
}

func (c *ProjectConfig) GetLobbyRegisterFile() string {
	return filepath.Join(c.ProjectDir, c.LobbyRegisterFile)
}

func (c *ProjectConfig) GetServiceDir() string {
	return filepath.Join(c.ProjectDir, c.ServiceDir)
}

func NewProjectConfig() *ProjectConfig {
	return new(ProjectConfig)
}

func GetProjectConfigList(dir string) ([]*ProjectConfig, error) {
	fileNameList, err := file.ListFileNameWithExt(Project_Dir, file.Ext_TOML)
	if err != nil {
		panic(err)
	}
	cl := make([]*ProjectConfig, 0, len(fileNameList))
	for _, name := range fileNameList {
		c, err := readProjectConfigFromToml(dir, name)
		if err != nil {
			return nil, err
		}
		cl = append(cl, c)
	}
	sort.Slice(cl, func(i, j int) bool {
		return cl[i].Alias < cl[j].Alias
	})
	return cl, nil
}

func readProjectConfigFromToml(dir, name string) (*ProjectConfig, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(dir)

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := NewProjectConfig()
	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
