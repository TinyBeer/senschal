package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"seneschal/config"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		dir := config.Workout_Dir
		ext := config.Ext_CSV
		fileNameList, err := config.ListFilesWithExt(config.Workout_Dir, config.Ext_CSV)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, name := range fileNameList {
			fileName := filepath.Join(dir, name+"."+ext)
			fmt.Println("workout config file:", fileName)
			// 打开 CSV 文件
			file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			// 解析 CSV 到结构体切片
			var itemList []*config.WorkoutItem
			if err := gocsv.UnmarshalFile(file, &itemList); err != nil {
				panic(err)
			}

			// 打印结果
			for _, item := range itemList {
				fmt.Printf("%+v\n", item)
			}
		}

	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
