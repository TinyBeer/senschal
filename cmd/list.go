package cmd

import (
	"fmt"
	"seneschal/config"
	"seneschal/tool"

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
		var data [][]string
		data = append(data, []string{"alias", "host"})
		for k, v := range m {
			data = append(data, []string{k, v.SSH.Host})
		}
		tool.ShowTable(data)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
