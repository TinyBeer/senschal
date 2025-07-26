package cmd

import (
	"fmt"
	"seneschal/ui/component"
	"time"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		c := component.NewInlineText(10, "abcdefghijklmnopqrstuvwxyz")
		for i := range 50 {
			time.Sleep(time.Millisecond * 500)
			fmt.Print("\033[H\033[2J")
			fmt.Println(c.GetCurrentContent(i))
		}
	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
