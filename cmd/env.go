package cmd

import (
	"fmt"
	"seneschal/config"
	"seneschal/pkg/util"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:     "env",
	Short:   "environment manage tool",
	Long:    "list environment in config file",
	Example: "seneschal env",
	RunE: func(cmd *cobra.Command, args []string) error {
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get env config: %w", err)
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
		util.ShowTable(data)
		return nil
	},
}
