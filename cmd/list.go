package cmd

import (
	"fmt"
	"seneschal/config"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出服务器",
	Long:  "列出所有可以用于部署环境的机器",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(m) == 0 {
			fmt.Println("没有找到可用的配置")
			return
		}
		fmt.Printf("别名\t主机\n")
		for k, v := range m {
			fmt.Printf("%s\t%s\n", k, v.SSH.Host)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
