package config

import "github.com/spf13/viper"

type EnvConfig struct {
	Docker *Docker `mapstructure:"docker"`
}

type Docker struct {
	Enable    bool     `mapstructure:"enable"`
	ImageList []string `mapstructure:"image_list"`
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
