package todo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *TodoFileRepo {
	t.Helper()
	tmpDir := t.TempDir()
	return NewTodoFileRepo(filepath.Join(tmpDir, "todo.json"))
}

func TestTodoFileRepo_AddTodo(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	todo := NewTodo("buy milk")
	id, err := repo.AddTodo(ctx, todo)
	require.NoError(t, err)
	assert.Equal(t, "1", id)
	assert.Equal(t, "1", todo.ID)
	assert.Equal(t, TodoStatus_Default, todo.Status)
	assert.Equal(t, "buy milk", todo.Content)
	assert.NotZero(t, todo.CreatedAt)
	assert.NotZero(t, todo.UpdatedAt)
}

func TestTodoFileRepo_AddTodoMultiple(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	id1, err := repo.AddTodo(ctx, NewTodo("first"))
	require.NoError(t, err)
	assert.Equal(t, "1", id1)

	id2, err := repo.AddTodo(ctx, NewTodo("second"))
	require.NoError(t, err)
	assert.Equal(t, "2", id2)
}

func TestTodoFileRepo_GetByID_Found(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.AddTodo(ctx, NewTodo("my task"))
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "my task", got.Content)
}

func TestTodoFileRepo_GetByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exist")
}

func TestTodoFileRepo_CompleteByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.AddTodo(ctx, NewTodo("do something"))
	require.NoError(t, err)

	err = repo.CompleteByID(ctx, id)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, TodoStatus_Completed, got.Status)
	assert.NotZero(t, got.CompletedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func TestTodoFileRepo_CompleteByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.CompleteByID(ctx, "999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exist")
}

func TestTodoFileRepo_DeleteByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.AddTodo(ctx, NewTodo("delete me"))
	require.NoError(t, err)

	err = repo.DeleteTodoByID(ctx, id)
	require.NoError(t, err)

	// Deleted item still exists in store but has DeletedAt set
	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.NotZero(t, got.DeletedAt)
}

func TestTodoFileRepo_DeleteByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.DeleteTodoByID(ctx, "999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exist")
}

func TestTodoFileRepo_List(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.AddTodo(ctx, NewTodo("active"))
	require.NoError(t, err)
	_, err = repo.AddTodo(ctx, NewTodo("active too"))
	require.NoError(t, err)

	items, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestTodoFileRepo_ListExcludesDeleted(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.AddTodo(ctx, NewTodo("will delete"))
	require.NoError(t, err)
	_, err = repo.AddTodo(ctx, NewTodo("keep me"))
	require.NoError(t, err)

	err = repo.DeleteTodoByID(ctx, id)
	require.NoError(t, err)

	items, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "keep me", items[0].Content)
}

func TestTodoFileRepo_LoadExistingStore(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "todo.json")

	// Create an initial store with one todo
	initialStore := `{"cur_id":5,"list":[{"id":"5","status":0,"content":"existing"}]}`
	require.NoError(t, os.WriteFile(fPath, []byte(initialStore), 0o644))

	repo := NewTodoFileRepo(fPath)
	ctx := context.Background()

	items, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "5", items[0].ID)
	assert.Equal(t, "existing", items[0].Content)

	// New todo should continue from cur_id=5
	id, err := repo.AddTodo(ctx, NewTodo("new one"))
	require.NoError(t, err)
	assert.Equal(t, "6", id)
}

func TestTodoFileRepo_LoadEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "todo.json")
	require.NoError(t, os.WriteFile(fPath, []byte(""), 0o644))

	repo := NewTodoFileRepo(fPath)
	ctx := context.Background()

	items, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestTodoFileRepo_LoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "todo.json")
	require.NoError(t, os.WriteFile(fPath, []byte("garbage{invalid"), 0o644))

	repo := NewTodoFileRepo(fPath)
	_, err := repo.List(context.Background())
	assert.Error(t, err)
}
