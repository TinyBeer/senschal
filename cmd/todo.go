package cmd

import (
	"context"
	"fmt"
	"log"
	"seneschal/tool"
	"seneschal/tool/todo"

	"github.com/spf13/cobra"
)

func init() {
	todoCmd.AddCommand(todoAddCmd)
	todoCmd.AddCommand(todoDoneCmd)
	todoCmd.AddCommand(todoDelCmd)
	rootCmd.AddCommand(todoCmd)
}

var todoCmd = &cobra.Command{
	Use:   "todo [command]",
	Short: "todo manage tool",
	Run: func(cmd *cobra.Command, args []string) {
		repo := todo.GetRepo()
		ctx := context.Background()
		list, err := repo.List(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if len(list) == 0 {
			fmt.Println("no todo exist, add one")
			return
		}

		tool.ShowTableWithSlice(list)
	},
}

var todoAddCmd = &cobra.Command{
	Use:   "add content",
	Short: "add new todo",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		content := args[0]
		repo := todo.GetRepo()
		ctx := context.Background()
		id, err := repo.AddTodo(ctx, todo.NewTodo(content))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("add todo[%s] success, get id[%s]", content, id)

	},
}

var todoDoneCmd = &cobra.Command{
	Use:   "done id",
	Short: "complete a todo",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		id := args[0]
		repo := todo.GetRepo()
		ctx := context.Background()
		err := repo.CompleteByID(ctx, id)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("complete todo[%s] success", id)

	},
}

var todoDelCmd = &cobra.Command{
	Use:   "del id",
	Short: "delete a todo",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		id := args[0]
		repo := todo.GetRepo()
		ctx := context.Background()
		err := repo.DeleteTodoByID(ctx, id)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("delete todo[%s] success", id)

	},
}
