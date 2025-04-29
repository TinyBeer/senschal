package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type SSHAuthMethod string

const (
	SSHAuthMethod_PW  SSHAuthMethod = "password"
	SSHAuthMethod_KEY SSHAuthMethod = "key"
)

type SSHConfig struct {
	Alias string `mapstructure:"alias"`
	SSH   *SSH   `mapstructure:"ssh"`
}
type SSH struct {
	User       string        `mapstructure:"user"`
	Password   string        `mapstructure:"password"`
	Host       string        `mapstructure:"host"`
	Port       int           `mapstructure:"port"`
	Method     SSHAuthMethod `mapstructure:"method"`
	PrivateKey string        `mapstructure:"private_key"`
}

func NewSSHConfig() *SSHConfig {
	return new(SSHConfig)

}

func ListFilesWithExt(dir string, ext string) ([]string, error) {
	var tomlFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == "."+ext {
			tomlFiles = append(tomlFiles, strings.TrimSuffix(info.Name(), "."+ext))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tomlFiles, nil
}

func readSSHConfigFromToml(dir, name string) (*SSHConfig, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(Ext_TOML)
	v.AddConfigPath(dir)

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := NewSSHConfig()
	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	if config.SSH != nil {
		if config.SSH.Method == "" {
			config.SSH.Method = SSHAuthMethod_PW
		}
		if config.SSH.Port == 0 {
			config.SSH.Port = 22
		}
	}

	return config, nil
}

func GetSSHConfigMap() (map[string]*SSHConfig, error) {
	return getSSHConfigMap(SSH_CFG_DIR)
}

func getSSHConfigMap(dir string) (map[string]*SSHConfig, error) {
	m := make(map[string]*SSHConfig)
	fileNames, err := ListFilesWithExt(dir, Ext_TOML)
	if err != nil {
		return nil, err
	}
	for _, name := range fileNames {
		c, err := readSSHConfigFromToml(dir, name)
		if err != nil {
			return nil, err
		}
		if c.Alias != "" {
			m[c.Alias] = c
		} else {
			m[name] = c
		}
	}
	return m, nil
}
