package cmd

import (
	"context"
	"fmt"
	"log"
	"seneschal/pkg/util"
	"seneschal/internal/command/todo"

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
	Example: "seneschal todo [add|done|del] [args]",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := todo.GetRepo()
		ctx := context.Background()
		list, err := repo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list todos: %w", err)
		}
		if len(list) == 0 {
			fmt.Println("no todo exist, add one")
			return nil
		}

		util.ShowTableWithSlice(list)
		return nil
	},
}

var todoAddCmd = &cobra.Command{
	Use:   "add content",
	Short: "add new todo",
	Example: "seneschal todo add <content>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content := args[0]
		repo := todo.GetRepo()
		ctx := context.Background()
		id, err := repo.AddTodo(ctx, todo.NewTodo(content))
		if err != nil {
			return fmt.Errorf("failed to add todo: %w", err)
		}
		log.Printf("add todo[%s] success, get id[%s]", content, id)
		return nil
	},
}

var todoDoneCmd = &cobra.Command{
	Use:   "done id",
	Short: "complete a todo",
	Example: "seneschal todo done <id>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		repo := todo.GetRepo()
		ctx := context.Background()
		err := repo.CompleteByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to complete todo: %w", err)
		}
		log.Printf("complete todo[%s] success", id)
		return nil
	},
}

var todoDelCmd = &cobra.Command{
	Use:   "del id",
	Short: "delete a todo",
	Example: "seneschal todo del <id>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		repo := todo.GetRepo()
		ctx := context.Background()
		err := repo.DeleteTodoByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete todo: %w", err)
		}
		log.Printf("delete todo[%s] success", id)
		return nil
	},
}
