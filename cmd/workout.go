package cmd

import (
	"fmt"
	"log"
	"seneschal/config"
	"seneschal/tool"
	"seneschal/ui/terminal"
	"sort"

	"github.com/spf13/cobra"
)

var workoutCmd = &cobra.Command{
	Use:   "workout",
	Short: "workout",
	Long: `workout tool:
workout workout_name: run workout
-l: list workout
`,
	Run: func(cmd *cobra.Command, args []string) {
		wcm, err := config.GetWorkoutConfigMap(config.Workout_Dir)
		if err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}
		if list {
			var data [][]string
			data = append(data, []string{"name"})
			for _, wc := range wcList {
				data = append(data, []string{wc.Name})
			}
			tool.ShowTable(data)
			return
		}
		var wc *config.WorkoutConfig
		var has bool
		if len(args) == 1 {
			workoutName := args[0]
			if wc, has = wcm[workoutName]; !has {
				fmt.Println("Please enter a correct workout name!")
				return
			}
		}

		terminal.Workout(wcList, wc)
	}}

func init() {
	workoutCmd.Flags().BoolP("list", "l", false, "lsit available workout")
	rootCmd.AddCommand(workoutCmd)
}
