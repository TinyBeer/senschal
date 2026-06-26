package cmd

import (
	"fmt"
	"seneschal/internal/runner"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cpCmd)
}

// copy command
var cpCmd = &cobra.Command{
	Use:     "cp",
	Short:   "agent copy file",
	Long:    "copy file between agent(not support fold yet)",
	Example: "seneschal agent cp <src> <dst>",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("cp %s to %s\n", args[0], args[1])
		err := runner.Copy(args[0], args[1])
		if err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}
		return nil
	},
}
