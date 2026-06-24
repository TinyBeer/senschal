package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"seneschal/pkg/util"
	"seneschal/internal/command/img"
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
	Use:     "img",
	Short:   "image tool",
	Long:    "some useful image handler",
	Example: "seneschal img [text|edge]",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var img2TextCmd = &cobra.Command{
	Use:     "text",
	Short:   "text effect",
	Long:    "用法: img text <input.ext>\n处理后的文件将保存为 <input_file_name.json>",
	Example: "seneschal img text <input.ext>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		width, err := getIntFlag(cmd, "width")
		if err != nil {
			return err
		}
		height, err := getIntFlag(cmd, "height")
		if err != nil {
			return fmt.Errorf("failed to parse --height flag: %w", err)
		}
		invert, err := getBoolFlag(cmd, "invert")
		if err != nil {
			return fmt.Errorf("failed to parse --invert flag: %w", err)
		}
		colors, err := getBoolFlag(cmd, "colors")
		if err != nil {
			return fmt.Errorf("failed to parse --colors flag: %w", err)
		}

		inputPath := args[0]

		data, err := img.ConvertImage2Text(inputPath, width, height, invert, colors)
		if err != nil {
			return fmt.Errorf("failed to convert image to text: %w", err)
		}
		ext := filepath.Ext(inputPath)
		outputPath := filepath.Join(strings.TrimSuffix(inputPath, ext) + ".json")
		bs, _ := json.Marshal(data)
		err = util.SaveStringToFile(outputPath, string(bs))
		if err != nil {
			return fmt.Errorf("failed to save output file: %w", err)
		}

		fmt.Println("处理完成，输出文件:", outputPath)
		return nil
	},
}

var imgEdgeEffectCmd = &cobra.Command{
	Use:     "edge",
	Short:   "edge effect",
	Long:    "用法: img edge <input.gif>\n处理后的文件将保存为 <input_edges.gif>",
	Example: "seneschal img edge <input.gif>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		ext := filepath.Ext(inputPath)
		outputPath := inputPath[:len(inputPath)-len(ext)] + "_edges" + ext

		err := img.ProcessGIF(inputPath, outputPath)
		if err != nil {
			return fmt.Errorf("处理失败: %w", err)
		}

		fmt.Println("处理完成，输出文件:", outputPath)
		return nil
	},
}
