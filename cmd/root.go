package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "seneschal",
	Short: "环境部署工具",
	Long:  "一个用于快速部署测试环境的工具",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("欢迎使用 seneschal")
	},
}

func Execut() error {
	return rootCmd.Execute()
}
