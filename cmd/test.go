package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(
			`
╭─────╮
│     │
│     │
╰─────╯
`,
		)

	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
