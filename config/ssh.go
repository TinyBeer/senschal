package config

import (
	"fmt"
	"os"
	"path/filepath"

	"seneschal/internal/command/file"

	"github.com/go-viper/mapstructure/v2"
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

func readSSHConfigFromToml(dir, name string) (*SSHConfig, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(file.Ext_TOML)
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
	return getSSHConfigMap(SSHConfigDir)
}

func getSSHConfigMap(dir string) (map[string]*SSHConfig, error) {
	m := make(map[string]*SSHConfig)
	fileNames, err := file.ListFileNameWithExt(dir, file.Ext_TOML)
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

// WriteSSHConfig 通过 viper 将 SSH 配置写入 TOML 文件
func WriteSSHConfig(cfg *SSHConfig) error {
	if cfg == nil || cfg.SSH == nil {
		return fmt.Errorf("ssh config is nil")
	}

	var m map[string]interface{}
	if err := mapstructure.Decode(cfg, &m); err != nil {
		return fmt.Errorf("decode ssh config: %w", err)
	}

	v := viper.New()
	v.SetConfigName(cfg.Alias)
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(SSHConfigDir)

	for k, val := range m {
		if sub, ok := val.(map[string]interface{}); ok {
			for sk, sv := range sub {
				if s, ok := sv.(string); ok && s == "" {
					continue
				}
				if sv != nil {
					v.Set(k+"."+sk, sv)
				}
			}
		} else {
			if val != nil {
				v.Set(k, val)
			}
		}
	}

	if err := os.MkdirAll(SSHConfigDir, 0o755); err != nil {
		return fmt.Errorf("failed to create ssh config dir: %w", err)
	}

	cfgPath := filepath.Join(SSHConfigDir, cfg.Alias+".toml")
	if err := v.WriteConfigAs(cfgPath); err != nil {
		return fmt.Errorf("failed to write ssh config: %w", err)
	}
	return nil
}
