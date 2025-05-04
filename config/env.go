package config

import (
	"os"
	"path/filepath"
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
	Docker *Docker `mapstructure:"docker"`
}

type Docker struct {
	Enable    bool    `mapstructure:"enable"`
	ImageList []Image `mapstructure:"image_list"`
}

func GetEnvConfig() (*EnvConfig, error) {
	return getEnvConfig(ENV_CFG_DIR)
}
func getEnvConfig(dir string) (*EnvConfig, error) {
	v := viper.New()
	v.SetConfigName("env")
	v.SetConfigType(Ext_TOML)
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
