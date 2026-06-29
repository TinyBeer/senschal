package config

import (
	"fmt"
	"os"
	"path/filepath"
	"seneschal/internal/command/file"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Jenkins struct {
	Alias    string `mapstructure:"alias"`
	Host     string `mapstructure:"host"`
	UserName string `mapstructure:"user_name"`
	Password string `mapstructure:"password"` // password or token
}

func GetJenkinsConfigMap() (map[string]*Jenkins, error) {
	return getJenkinsConfigMap(Jenkins_CFG_DIR)
}

func getJenkinsConfigMap(dir string) (map[string]*Jenkins, error) {
	m := make(map[string]*Jenkins)
	fileNames, err := file.ListFileNameWithExt(dir, file.Ext_TOML)
	if err != nil {
		return nil, err
	}
	for _, name := range fileNames {
		c, err := readJenkinsConfigFromToml(dir, name)
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

func readJenkinsConfigFromToml(dir, name string) (*Jenkins, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(dir)

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := new(Jenkins)
	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// WriteSSHConfig 通过 viper 将 Jenkins 配置写入 TOML 文件
func WriteJenkinsConfig(cfg *Jenkins) error {
	if cfg.UserName == "" || cfg.Password == "" {
		return fmt.Errorf("jenkins config missing user name or password")
	}

	var m map[string]any
	if err := mapstructure.Decode(cfg, &m); err != nil {
		return fmt.Errorf("decode jenkins config: %w", err)
	}

	v := viper.New()
	v.SetConfigName(cfg.Alias)
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(Jenkins_CFG_DIR)

	for k, val := range m {
		if sub, ok := val.(map[string]any); ok {
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

	if err := os.MkdirAll(Jenkins_CFG_DIR, 0o755); err != nil {
		return fmt.Errorf("failed to create jenkins config dir: %w", err)
	}

	cfgPath := filepath.Join(Jenkins_CFG_DIR, cfg.Alias+".toml")
	if err := v.WriteConfigAs(cfgPath); err != nil {
		return fmt.Errorf("failed to write jenkins config: %w", err)
	}
	return nil
}
