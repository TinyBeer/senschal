package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"seneschal/tool"
	"time"

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
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		bs, err := io.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		data := new(tool.ImageTextData)
		err = json.Unmarshal(bs, data)
		if err != nil {
			log.Fatal(err)
		}
		for i := range 100 {
			time.Sleep(time.Millisecond * 500)
			fmt.Print("\033[H\033[2J")
			fmt.Println(data.Data[i%len(data.Data)])
		}
	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
