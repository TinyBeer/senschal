package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seneschal/config"
	"seneschal/tool"
	"seneschal/tool/file"
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
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Please enter a workout name!")
			return
		}
		workoutName := args[0]
		// 打开 CSV 文件
		file, err := os.OpenFile(filepath.Join(config.Workout_Dir, workoutName+"."+file.Ext_CSV), os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// 解析 CSV 到结构体切片
		var itemList []*config.WorkoutItem
		if err := gocsv.MarshalFile(&itemList, file); err != nil {
			panic(err)
		}
	},
}

var workoutCmd = &cobra.Command{
	Use:   "workout <workout_name>",
	Short: "start a workout",
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
