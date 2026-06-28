package cmd

import (
	"fmt"
	"path/filepath"

	"seneschal/config"
	"seneschal/internal/fsutil"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cpCmd)
}

// copy command
var cpCmd = &cobra.Command{
	Use:   "cp <src> <dst>",
	Short: "copy file or directory between local and remote",
	Long: `Copy file or directory between local and remote paths.

Supports alias:path format for remote paths.
Examples:
  seneschal cp file.txt myserver:/remote/path/
  seneschal cp myserver:/remote/file.txt /local/path/
  seneschal cp file.txt /local/backup/`,
	Example: "seneschal cp file.txt host1:/tmp/",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath := args[0]
		dstPath := args[1]

		srcRef, err := fsutil.Parse(srcPath)
		if err != nil {
			return fmt.Errorf("invalid source path: %w", err)
		}
		dstRef, err := fsutil.Parse(dstPath)
		if err != nil {
			return fmt.Errorf("invalid destination path: %w", err)
		}

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to load ssh config: %w", err)
		}

		srcFS, err := fsutil.GetFS(srcRef, scm)
		if err != nil {
			return fmt.Errorf("failed to access source: %w", err)
		}

		srcStat, err := srcFS.Stat(srcRef)
		if err != nil {
			return fmt.Errorf("stat source: %w", err)
		}
		if !srcStat.Exist {
			return fmt.Errorf("source not found: %s", srcPath)
		}

		// Check if destination exists and is a directory
		dstFS, err := fsutil.GetFS(dstRef, scm)
		if err != nil {
			return fmt.Errorf("failed to access destination: %w", err)
		}
		dstStat, err := dstFS.Stat(dstRef)
		if err != nil {
			return fmt.Errorf("stat destination: %w", err)
		}

		// Error: cannot copy directory into a file
		if srcStat.IsDir && dstStat.Exist && !dstStat.IsDir {
			return fmt.Errorf("cannot copy a directory to a file")
		}

		// If destination is an existing directory, nest source under it
		if dstStat.Exist && dstStat.IsDir {
			dstRef.RawPath = filepath.Join(dstRef.RawPath, filepath.Base(srcRef.RawPath))
		}

		transfer := fsutil.NewTransfer(scm)

		fmt.Printf("cp %s to %s\n", srcPath, dstPath)

		if srcStat.IsDir {
			if err := transfer.CopyDir(srcRef, dstRef); err != nil {
				return fmt.Errorf("copy directory failed: %w", err)
			}
		} else {
			if err := transfer.CopyFile(srcRef, dstRef); err != nil {
				return fmt.Errorf("copy file failed: %w", err)
			}
		}

		return nil
	},
}
