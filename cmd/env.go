package cmd

import (
	"fmt"
	"seneschal/config"
	"seneschal/tool"

	"github.com/spf13/cobra"
)

type IEnvMgr interface {
	GetName() string
	Check(c *config.SSHConfig) (any, error)
	Deploy(c *config.SSHConfig) error
}

func init() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "environment manage tool",
	Long:  "list environment in config file",
	Run: func(cmd *cobra.Command, args []string) {
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			fmt.Println(err)
			return
		}

		var data [][]string
		data = append(data, []string{"alias", "abstract"})
		for alias, ec := range ecm {
			abstract := "null"
			if ec.Docker != nil && ec.Docker.Enable {
				abstract = "docker enable"
				if len(ec.Docker.ImageList) != 0 {
					abstract = fmt.Sprintf("docker with images: %v", ec.Docker.ImageList)
				}
			}
			data = append(data, []string{alias, abstract})
		}
		tool.ShowTable(data)
	},
}
