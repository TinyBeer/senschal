package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"seneschal/config"
	"seneschal/pkg/util"
	"seneschal/internal/command/file"
	"seneschal/ui/terminal"
	"sort"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
)

func init() {
	workoutCmd.AddCommand(newWorkoutCmd)
	workoutCmd.Flags().BoolP("list", "l", false, "lsit available workout")
	rootCmd.AddCommand(workoutCmd)
}

var newWorkoutCmd = &cobra.Command{
	Use:   "new <workout_name>",
	Short: "creat a new workout",
	Long:  `generate a new workout config file at workout config dir`,
	Example: "seneschal workout new <name>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workoutName := args[0]
		f, err := os.OpenFile(filepath.Join(config.Workout_Dir, workoutName+"."+file.Ext_CSV), os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create workout file: %w", err)
		}
		defer f.Close()

		var itemList []*config.WorkoutItem
		if err := gocsv.MarshalFile(&itemList, f); err != nil {
			return fmt.Errorf("failed to write workout csv: %w", err)
		}
		return nil
	},
}

var workoutCmd = &cobra.Command{
	Use:   "workout <workout_name>",
	Short: "start a workout",
	Long: `workout tool:
	workout workout_name: run workout
	-l: list workout
`,
	Example: "seneschal workout <workout_name> [-l]",
	RunE: func(cmd *cobra.Command, args []string) error {
		wcm, err := config.GetWorkoutConfigMap(config.Workout_Dir)
		if err != nil {
			return fmt.Errorf("failed to get workout config: %w", err)
		}

		wcList := make([]*config.WorkoutConfig, 0, len(wcm))
		for _, wc := range wcm {
			wcList = append(wcList, wc)
		}
		sort.Slice(wcList, func(i, j int) bool {
			return wcList[i].Name < wcList[j].Name
		})

		list, err := cmd.Flags().GetBool("list")
		if err != nil {
			return fmt.Errorf("failed to parse --list flag: %w", err)
		}
		if list {
			var data [][]string
			data = append(data, []string{"name"})
			for _, wc := range wcList {
				data = append(data, []string{wc.Name})
			}
			util.ShowTable(data)
			return nil
		}
		var wc *config.WorkoutConfig
		var has bool
		if len(args) == 1 {
			workoutName := args[0]
			if wc, has = wcm[workoutName]; !has {
				return fmt.Errorf("workout [%s] not found", workoutName)
			}
		}

		terminal.Workout(wcList, wc)
		return nil
	}}
