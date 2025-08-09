package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type TodoFileRepo struct {
	File string
}

func NewTodoFileRepo(file string) *TodoFileRepo {
	return &TodoFileRepo{
		File: file,
	}
}

type TodoStore struct {
	CurID int     `json:"cur_id,omitempty"`
	List  []*Todo `json:"list,omitempty"`
}

// AddTodo implements TodoRepository.
func (r *TodoFileRepo) AddTodo(ctx context.Context, todo *Todo) (string, error) {
	curUnix := time.Now().Unix()
	store, err := r.getStore()
	if err != nil {
		return "", err
	}

	store.CurID++
	id := strconv.Itoa(store.CurID)
	todo.ID = id
	todo.CreatedAt = curUnix
	todo.UpdatedAt = curUnix
	store.List = append(store.List, todo)
	bs, _ := json.Marshal(store)
	err = os.WriteFile(r.File, bs, 0655)
	return id, err
}

// CloseTodoByID implements TodoRepository.
func (r *TodoFileRepo) CloseTodoByID(ctx context.Context, id string) error {
	curUnix := time.Now().Unix()
	store, err := r.getStore()
	if err != nil {
		return err
	}
	var todo *Todo
	for _, t := range store.List {
		if t.ID == id {
			todo = t
			break
		}
	}
	if todo == nil {
		return fmt.Errorf("todo with id[%s] not exist", id)
	}
	todo.UpdatedAt = curUnix
	todo.Status = TodoStatus_Closed

	bs, _ := json.Marshal(store)
	return os.WriteFile(r.File, bs, 0644)
}

// CompleteByID implements TodoRepository.
func (r *TodoFileRepo) CompleteByID(ctx context.Context, id string) error {
	curUnix := time.Now().Unix()
	store, err := r.getStore()
	if err != nil {
		return err
	}
	var todo *Todo
	for _, t := range store.List {
		if t.ID == id {
			todo = t
			break
		}
	}
	if todo == nil {
		return fmt.Errorf("todo with id[%s] not exist", id)
	}
	todo.CompletedAt = curUnix
	todo.UpdatedAt = curUnix
	todo.Status = TodoStatus_Completed

	bs, _ := json.Marshal(store)
	return os.WriteFile(r.File, bs, 0644)
}

// DeleteTodoByID implements TodoRepository.
func (r *TodoFileRepo) DeleteTodoByID(ctx context.Context, id string) error {
	curUnix := time.Now().Unix()
	store, err := r.getStore()
	if err != nil {
		return err
	}
	var todo *Todo
	for _, t := range store.List {
		if t.ID == id {
			todo = t
			break
		}
	}
	if todo == nil {
		return fmt.Errorf("todo with id[%s] not exist", id)
	}
	todo.DeletedAt = curUnix

	bs, _ := json.Marshal(store)
	return os.WriteFile(r.File, bs, 0644)
}

// GetByID implements TodoRepository.
func (r *TodoFileRepo) GetByID(ctx context.Context, id string) (*Todo, error) {
	store, err := r.getStore()
	if err != nil {
		return nil, err
	}

	for _, todo := range store.List {
		if todo.ID == id {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("todo with id[%s] not exist", id)
}

// List implements TodoRepository.
func (r *TodoFileRepo) List(ctx context.Context) ([]*Todo, error) {
	store, err := r.getStore()
	if err != nil {
		return nil, err
	}

	notDeletedList := make([]*Todo, 0, len(store.List))
	for _, todo := range store.List {
		if todo.DeletedAt == 0 {
			notDeletedList = append(notDeletedList, todo)
		}
	}
	return notDeletedList, nil
}

func (r *TodoFileRepo) getStore() (*TodoStore, error) {
	if err := os.MkdirAll(filepath.Dir(r.File), 0755); err != nil {
		return nil, err
	}
	store := &TodoStore{
		CurID: 0,
		List:  nil,
	}

	stat, err := os.Stat(r.File)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}

	if stat.Size() == 0 {
		return store, nil
	}

	bs, err := os.ReadFile(r.File)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bs, store)
	return store, err
}

var _ TodoRepository = new(TodoFileRepo)
