package cmd

import (
	"fmt"
	"log"
	"seneschal/tool/file"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println(args)
			return
		}

		filePath := args[0]
		err := file.InsertCodeIntoProto(filePath, file.ReplaceProbe_Message, "message Test2 {\n}\n\n")
		if err != nil {
			log.Fatal(err)
		}

	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
