package cmd

import (
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	cfgFile string
	output  string
)

var rootCmd = &cobra.Command{
	Use:     "seneschal",
	Short:   "环境部署工具",
	Long:    "一个用于快速部署测试环境的工具",
	Example: "seneschal -h",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !verbose {
			log.SetOutput(io.Discard)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("欢迎使用 seneschal")
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format (table, json, yaml)")
}

func Execut() error {
	return rootCmd.Execute()
}
