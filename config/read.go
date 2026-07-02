package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

func readAllFromFile(path, typ string) (map[string]any, error) {
	v := viper.New()
	v.SetConfigName(filepath.Base(path))
	v.SetConfigType(typ)
	v.AddConfigPath(filepath.Dir(path))
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	settings := v.AllSettings()
	return settings, nil
}
