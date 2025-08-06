package cmd

import (
	"fmt"
	"seneschal/ui/component"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		r1 := component.NewRectangle(component.NewInlineTextWithStyle(6, "asdfghjkl", component.StyleFinish), true, component.StyleTodo)
		r2 := component.NewRectangle(component.NewInlineTextWithStyle(6, "asdfghjkl", component.StyleWorking), true, component.StyleBreaking)
		v := component.NewBox(component.Direction_H)
		v.AddSub(r1)
		v.AddSub(r2)
		for f := range 100 {
			fmt.Print("\033[H\033[2J")
			fmt.Println(strings.Join(component.JoinStyleStringMatrix(v.GetCurrentContent(f)), "\n"))
			time.Sleep(time.Second * 2)
		}
	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
