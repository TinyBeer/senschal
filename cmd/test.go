package cmd

import (
	"fmt"
	"log"
	"seneschal/config"
	"seneschal/tool"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			log.Println(err)
			return
		}
		if len(args) == 0 {
			log.Println("请输入需要检查的机器别名")
			return
		}
		alias := args[0]
		sc := m[alias]
		se, err := tool.NewSSHExecutor(sc)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = se.ExecuteCommand("getent group docker > /dev/null 2>&1")
		if err != nil {
			if err.Error() != "Process exited with status 2" {
				fmt.Println(err)
				return
			}
			fmt.Println("missing docker group")
			_, err = se.ExecuteCommand("sudo groupadd docker")
			if err != nil {
				fmt.Println(err)
				return
			}

		}

		_, err = se.ExecuteCommand("id -Gn | grep docker")
		if err != nil {
			if err.Error() != "Process exited with status 1" {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("user not in docker group")
			_, err = se.ExecuteCommand("sudo usermod -aG docker $USER & newgrp docker")
			if err != nil {
				fmt.Println(err)
				return
			}
		}

	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
