package cmd

import (
	"fmt"
	"seneschal/config"
	"seneschal/tool"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		img := config.Image("mysql:5.7")
		fmt.Println(img.Name())
		fmt.Println(img.LocalFileExist())
		fmt.Println(img.LocalFilePath())
		if !img.LocalFileExist() {
			bs, err := tool.ExecuteCommand(fmt.Sprintf("docker pull %s", string(img)))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(bs))
			bs, err = tool.ExecuteCommand(fmt.Sprintf("docker save -o %s %s", img.LocalFilePath(), string(img)))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(bs))
		}

	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
