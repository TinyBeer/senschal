package cmd

import (
	"fmt"
	"log"
	"seneschal/config"
	"seneschal/tool"

	"github.com/spf13/cobra"
)

type File struct {
	SSHConfig *config.SSHConfig
	Path      string
}

var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "文件拷贝",
	Long:  "在机器之间拷贝文件 暂不支持文件夹操作",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Println("请输入需要拷贝的文件以及要拷贝到的路径")
			return
		}
		fmt.Printf("cp %s to %s\n", args[0], args[1])
		err := tool.Copy(args[0], args[1])
		if err != nil {
			log.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
