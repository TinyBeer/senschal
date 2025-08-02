package config

import (
	"fmt"
	"os"
	"path/filepath"
	"seneschal/tool/file"
	"strings"

	"github.com/spf13/viper"
)

type Image string

func (img Image) Name() string {
	return strings.ReplaceAll(strings.ReplaceAll(string(img), ":", "_"), "/", "-") + ".tar"
}

func (img Image) LocalFilePath() string {
	return filepath.Join(DOCKER_IMAGE_DIR, img.Name())
}

func (img Image) LocalFileExist() bool {
	_, err := os.Stat(img.LocalFilePath())
	return !os.IsNotExist(err)
}

type EnvConfig struct {
	Alias string `mapstructure:"alias"`
	// Default bool    `mapstructure:"default"`
	Docker *Docker `mapstructure:"docker"`
}

type Docker struct {
	Enable         bool    `mapstructure:"enable"`
	Version        string  `mapstructure:"version"`
	CheckUserGroup bool    `mapstructure:"check_user_group"`
	ImageList      []Image `mapstructure:"image_list"`
}

// func GetEnvConfig() (*EnvConfig, error) {
// 	return getEnvConfig(ENV_CFG_DIR)
// }
// func getEnvConfig(dir string) (*EnvConfig, error) {
// 	v := viper.New()
// 	v.SetConfigName("env")
// 	v.SetConfigType(Ext_TOML)
// 	v.AddConfigPath(dir)

// 	// 读取配置文件
// 	err := v.ReadInConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	config := new(EnvConfig)
// 	err = v.Unmarshal(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return config, nil
// }

func readEnvConfigFromToml(dir, name string) (*EnvConfig, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(dir)

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := new(EnvConfig)
	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetEnvConfigMap() (map[string]*EnvConfig, error) {
	return getEnvConfigMap(ENV_CFG_DIR)
}

func getEnvConfigMap(dir string) (map[string]*EnvConfig, error) {
	m := make(map[string]*EnvConfig)
	fileNames, err := file.ListFileNameWithExt(dir, file.Ext_TOML)
	if err != nil {
		return nil, err
	}
	for _, name := range fileNames {
		c, err := readEnvConfigFromToml(dir, name)
		if err != nil {
			return nil, fmt.Errorf("failed to read config from [%s, %s], err: %v", dir, name, err)
		}
		alias := name
		if c.Alias != "" {
			alias = c.Alias
		}
		if _, has := m[alias]; has {
			return nil, fmt.Errorf("duplicated env alias[%s]", alias)
		}
		m[alias] = c
	}
	return m, nil
}
