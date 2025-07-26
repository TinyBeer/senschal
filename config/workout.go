package config

import (
	"github.com/spf13/viper"
)

type WorkoutType string

const (
	DefaultBreak         = 10
	WorkoutType_Duration = "duration"
	WorkoutType_Count    = "count"
)

type WorkoutConfig struct {
	Name     string         `mapstructure:"name"`
	ItemList []*WorkoutItem `mapstructure:"item_list"`
}

type WorkoutItem struct {
	Name   string      `mapstructure:"name"`
	Type   WorkoutType `mapstructure:"type"`
	Repeat int         `mapstructure:"repeat"`
	Target int         `mapstructure:"target"`
	Break  int         `mapstructure:"break"`
}

func NewWorkoutConfig() *WorkoutConfig {
	return new(WorkoutConfig)

}

func GetWorkoutConfigMap(dir string) (map[string]*WorkoutConfig, error) {
	fileNameList, err := ListFilesWithExt(dir, Ext_TOML)
	if err != nil {
		return nil, err
	}
	cm := make(map[string]*WorkoutConfig)
	for _, fileName := range fileNameList {
		wc, err := readWorkoutConfigFromToml(dir, fileName)
		if err != nil {
			return nil, err
		}
		cm[wc.Name] = wc
	}
	return cm, nil
}

func readWorkoutConfigFromToml(dir, name string) (*WorkoutConfig, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(Ext_TOML)
	v.AddConfigPath(dir)

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := NewWorkoutConfig()
	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	for _, v := range config.ItemList {
		if v.Repeat == 0 {
			v.Repeat = 1
		}
	}

	return config, nil
}
