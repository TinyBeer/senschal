package config

import (
	"fmt"
	"os"
	"path/filepath"

	"seneschal/internal/command/file"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// writeConfigToToml 通过 viper 将任意配置结构体写入 TOML 文件
func writeConfigToToml(cfg any, dir, alias, typeName string) error {
	var m map[string]any
	if err := mapstructure.Decode(cfg, &m); err != nil {
		return fmt.Errorf("decode %s config: %w", typeName, err)
	}

	v := viper.New()
	v.SetConfigName(alias)
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(dir)

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

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s config dir: %w", typeName, err)
	}

	cfgPath := filepath.Join(dir, alias+".toml")
	if err := v.WriteConfigAs(cfgPath); err != nil {
		return fmt.Errorf("failed to write %s config: %w", typeName, err)
	}
	return nil
}
