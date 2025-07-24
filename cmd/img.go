package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"seneschal/tool"

	"github.com/spf13/cobra"
)

func init() {
	imgCmd.AddCommand(edgeCmd)
	rootCmd.AddCommand(imgCmd)
}

var imgCmd = &cobra.Command{
	Use:   "img",
	Short: "image tool",
	Long:  "some useful image handler",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var edgeCmd = &cobra.Command{
	Use:   "edge",
	Short: "edge effect",
	Long:  "用法: img edge <input.gif>\n处理后的文件将保存为 <input_edges.gif>",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		inputPath := args[0]
		ext := filepath.Ext(inputPath)
		outputPath := inputPath[:len(inputPath)-len(ext)] + "_edges" + ext

		err := tool.ProcessGIF(inputPath, outputPath)
		if err != nil {
			println("处理失败:", err.Error())
			os.Exit(1)
		}

		println("处理完成，输出文件:", outputPath)
	},
}
