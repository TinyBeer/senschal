package todo

//go:generate  stringer -type=TodoStatus -linecomment
type TodoStatus int

const (
	TodoStatus_Default   TodoStatus = iota //default
	TodoStatus_Completed                   // completed
	TodoStatus_Closed                      // closed
)

type Todo struct {
	ID          string
	Status      TodoStatus
	Content     string
	CompletedAt int64
	CreatedAt   int64
	UpdatedAt   int64
	DeletedAt   int64
}

func NewTodo(content string) *Todo {
	return &Todo{
		ID:          "",
		Status:      TodoStatus_Default,
		Content:     content,
		CompletedAt: 0,
		CreatedAt:   0,
		UpdatedAt:   0,
		DeletedAt:   0,
	}
}
