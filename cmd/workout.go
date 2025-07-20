package cmd

import (
	"fmt"
	"log"
	"seneschal/config"
	"seneschal/tool"
	"seneschal/ui/terminal"

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
		list, err := cmd.Flags().GetBool("list")
		if err != nil {
			log.Fatal(err)
		}
		if list {
			var data [][]string
			data = append(data, []string{"name"})
			for name, wc := range wcm {
				data = append(data, []string{name})
				_ = wc
			}
			tool.ShowTable(data)
			return
		} else {
			if len(args) != 1 {
				log.Fatal("Please enter a workout name!")
			}
			fmt.Println(args)
		}

		workoutName := args[0]
		if wc, has := wcm[workoutName]; !has {
			fmt.Println("Please enter a correct workout name!")
		} else {
			terminal.RunWithWorkoutConfig(wc)
		}
	}}

func init() {
	workoutCmd.Flags().BoolP("list", "l", false, "lsit available workout")
	rootCmd.AddCommand(workoutCmd)
}
