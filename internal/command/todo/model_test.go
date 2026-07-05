package todo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodoStatus_String(t *testing.T) {
	assert.Equal(t, "default", TodoStatus_Default.String())
	assert.Equal(t, "completed", TodoStatus_Completed.String())
	assert.Equal(t, "closed", TodoStatus_Closed.String())
}

func TestTodoStatus_String_Invalid(t *testing.T) {
	var s TodoStatus = 99
	assert.Contains(t, s.String(), "TodoStatus(99)")
}

func TestNewTodo(t *testing.T) {
	todo := NewTodo("test content")
	assert.Equal(t, "test content", todo.Content)
	assert.Equal(t, TodoStatus_Default, todo.Status)
	assert.Empty(t, todo.ID)
	assert.Zero(t, todo.CreatedAt)
	assert.Zero(t, todo.CompletedAt)
	assert.Zero(t, todo.UpdatedAt)
	assert.Zero(t, todo.DeletedAt)
}
