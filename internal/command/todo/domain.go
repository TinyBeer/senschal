package todo

import (
	"context"
	"path/filepath"
	"sync"

	"seneschal/config"
)

var (
	once sync.Once
	repo TodoRepository
)

func GetRepo() TodoRepository {
	once.Do(func() {
		repo = NewTodoFileRepo(filepath.Join(config.TodoDir, "todo.json"))
	})
	return repo
}

type TodoRepository interface {
	List(ctx context.Context) ([]*Todo, error)
	GetByID(ctx context.Context, id string) (*Todo, error)
	AddTodo(ctx context.Context, todo *Todo) (string, error)
	CompleteByID(ctx context.Context, id string) error
	// CloseTodoByID(ctx context.Context, id string) error
	DeleteTodoByID(ctx context.Context, id string) error
}
