package todo

import (
	"context"
	"path/filepath"
	"sync"

	"seneschal/config"
)

var (
	once sync.Once
	repo TodoRepo
)

func GetRepo() TodoRepo {
	once.Do(func() {
		repo = NewTodoFileRepo(filepath.Join(config.TodoDir, "todo.json"))
	})
	return repo
}

type TodoRepo interface {
	TodoBasicRepo
	TodoBatchRepo
}

// TodoBasicRepo 基础操作
type TodoBasicRepo interface {
	GetByID(ctx context.Context, id string) (*Todo, error)
	AddTodo(ctx context.Context, todo *Todo) (string, error)
	CompleteByID(ctx context.Context, id string) error
	DeleteTodoByID(ctx context.Context, id string) error
}

// TodoBatchRepo 批量操作
type TodoBatchRepo interface {
	List(ctx context.Context) ([]*Todo, error)
}
