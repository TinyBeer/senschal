package config

import (
	"os"
	"path/filepath"
	"seneschal/tool/file"

	"github.com/gocarina/gocsv"
)

const DefaultBreak = 10

//go:generate stringer --type=WorkoutType -linecomment
type WorkoutType int

const (
	WorkoutType_Duration WorkoutType = iota + 1 // duration
	WorkoutType_Count                           // count
)

type WorkoutConfig struct {
	Name     string         `mapstructure:"name"`
	ItemList []*WorkoutItem `mapstructure:"item_list"`
}

type WorkoutItem struct {
	Name   string      `mapstructure:"name" csv:"name"`
	Type   WorkoutType `mapstructure:"type" csv:"type"`
	Repeat int         `mapstructure:"repeat" csv:"repeat"`
	Target int         `mapstructure:"target" csv:"target"`
	Break  int         `mapstructure:"break" csv:"break"`
}

func NewWorkoutConfig() *WorkoutConfig {
	return new(WorkoutConfig)

}

func GetWorkoutConfigMap(dir string) (map[string]*WorkoutConfig, error) {
	fileNameList, err := file.ListFileNameWithExt(dir, file.Ext_CSV)
	if err != nil {
		return nil, err
	}
	cm := make(map[string]*WorkoutConfig)
	for _, fileName := range fileNameList {
		wc, err := readWorkoutConfigFromCsv(dir, fileName)
		if err != nil {
			return nil, err
		}
		cm[wc.Name] = wc
	}
	return cm, nil
}

// func readWorkoutConfigFromToml(dir, name string) (*WorkoutConfig, error) {
// 	v := viper.New()
// 	v.SetConfigName(name)
// 	v.SetConfigType(Ext_TOML)
// 	v.AddConfigPath(dir)

// 	// 读取配置文件
// 	err := v.ReadInConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	config := NewWorkoutConfig()
// 	err = v.Unmarshal(config)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, v := range config.ItemList {
// 		if v.Repeat == 0 {
// 			v.Repeat = 1
// 		}
// 	}

// 	return config, nil
// }

func readWorkoutConfigFromCsv(dir, name string) (*WorkoutConfig, error) {
	fileName := filepath.Join(dir, name+"."+file.Ext_CSV)
	// 打开 CSV 文件
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 解析 CSV 到结构体切片
	var itemList []*WorkoutItem
	if err := gocsv.UnmarshalFile(file, &itemList); err != nil {
		return nil, err
	}

	for _, v := range itemList {
		if v.Repeat == 0 {
			v.Repeat = 1
		}
	}

	return &WorkoutConfig{
		Name:     name,
		ItemList: itemList,
	}, nil
}
