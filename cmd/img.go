package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seneschal/tool"
	"seneschal/tool/img"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	img2TextCmd.Flags().IntP("width", "W", 80, "输出字符画宽度")
	img2TextCmd.Flags().IntP("height", "H", 0, "输出字符画高度(0表示根据原图比例自动计算)")
	img2TextCmd.Flags().BoolP("invert", "V", false, "是否反转亮度(黑白颠倒)")
	img2TextCmd.Flags().BoolP("colors", "C", false, "是否使用ANSI颜色输出")
	imgCmd.AddCommand(img2TextCmd)

	imgCmd.AddCommand(imgEdgeEffectCmd)
	rootCmd.AddCommand(imgCmd)
}

var imgCmd = &cobra.Command{
	Use:   "img",
	Short: "image tool",
	Long:  "some useful image handler",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var img2TextCmd = &cobra.Command{
	Use:   "text",
	Short: "text effect",
	Long:  "用法: img text <input.ext>\n处理后的文件将保存为 <input_file_name.json>",
	Run: func(cmd *cobra.Command, args []string) {
		width, err := cmd.Flags().GetInt("width")
		if err != nil {
			log.Fatal(err)
		}
		height, err := cmd.Flags().GetInt("height")
		if err != nil {
			log.Fatal(err)
		}
		invert, err := cmd.Flags().GetBool("invert")
		if err != nil {
			log.Fatal(err)
		}
		colors, err := cmd.Flags().GetBool("colors")
		if err != nil {
			log.Fatal(err)
		}
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		inputPath := args[0]

		data, err := img.ConvertImage2Text(inputPath, width, height, invert, colors)
		if err != nil {
			log.Fatal(err)
		}
		ext := filepath.Ext(inputPath)
		outputPath := filepath.Join(strings.TrimSuffix(inputPath, ext) + ".json")
		bs, _ := json.Marshal(data)
		err = tool.SaveStringToFile(outputPath, string(bs))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("处理完成，输出文件:", outputPath)
	},
}

var imgEdgeEffectCmd = &cobra.Command{
	Use:   "edge",
	Short: "edge effect",
	Long:  "用法: img edge <input.gif>\n处理后的文件将保存为 <input_edges.gif>",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		inputPath := args[0]
		ext := filepath.Ext(inputPath)
		outputPath := inputPath[:len(inputPath)-len(ext)] + "_edges" + ext

		err := img.ProcessGIF(inputPath, outputPath)
		if err != nil {
			fmt.Println("处理失败:", err.Error())
			os.Exit(1)
		}

		fmt.Println("处理完成，输出文件:", outputPath)
	},
}
