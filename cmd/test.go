package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"seneschal/config"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试",
	Long:  "开发阶段用于测试命令",
	Run: func(cmd *cobra.Command, args []string) {
		wcm, err := config.GetWorkoutConfigMap(config.Workout_Dir)
		if err != nil {
			log.Fatal(err)
		}
		for name, conf := range wcm {
			bs, _ := json.Marshal(conf)
			fmt.Printf("name: %s conf: %v\n", name, string(bs))
		}
	}}

func init() {
	rootCmd.AddCommand(testCmd)
}
