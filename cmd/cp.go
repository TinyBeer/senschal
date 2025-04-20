package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"seneschal/config"
	"seneschal/tool"

	"github.com/spf13/cobra"
)

type File struct {
	SSHConfig *config.SSHConfig
	Path      string
}

var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "文件拷贝",
	Long:  "在机器之间拷贝文件 暂不支持文件夹操作",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Println("请输入需要拷贝的文件以及要拷贝到的路径")
			return
		}
		err := cp(args[0], args[1])
		if err != nil {
			log.Println(err)
			return
		}
	},
}

func cp(srcPath, dstPath string) error {
	srcFile, err := tool.NewFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to new source file with arg %v, err:%v", srcPath, err)
	}
	srcSize, err := srcFile.GetSize()
	if err != nil {
		return fmt.Errorf("failed to get source file size with arg %v, err:%v", dstPath, err)
	}

	dstFile, err := tool.NewFile(dstPath)
	if err != nil {
		return fmt.Errorf("failed to new destination file with arg %v, err:%v", dstPath, err)
	}
	dst, err := dstFile.GetWriter()
	if err != nil {
		return fmt.Errorf("failed to get writer from dst file, err:%v", err)
	}

	src, err := srcFile.GetReader()
	if err != nil {
		return fmt.Errorf("failed to get reader from src file, err:%v", err)
	}
	byteCnt, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file from %v to %v, err:%v", srcPath, dstPath, err)
	}
	log.Printf("copy %v byte from %v to %v", byteCnt, srcPath, dstPath)

	dstSize, err := dstFile.GetSize()
	if err != nil {
		return fmt.Errorf("failed to get destination file size with arg %v, err:%v", dstPath, err)
	}

	if srcSize != byteCnt || srcSize != dstSize {
		return errors.New("failed to copy file")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
