package ilog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenSimpleTraceID_Length(t *testing.T) {
	id := GenSimpleTraceID()
	assert.Len(t, id, traceLen)
}

func TestGenSimpleTraceID_Charset(t *testing.T) {
	id := GenSimpleTraceID()
	for _, c := range id {
		assert.Contains(t, letterBytes, string(c), "char %q not in charset", c)
	}
}

func TestGenSimpleTraceID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		id := GenSimpleTraceID()
		assert.False(t, seen[id], "duplicate trace ID: %s", id)
		seen[id] = true
	}
}

func TestNewTraceCtx(t *testing.T) {
	ctx := NewTraceCtx(context.Background())
	tid := GetTraceID(ctx)
	assert.Len(t, tid, traceLen)
}

func TestGetTraceID_WithTraceID(t *testing.T) {
	ctx := NewTraceCtx(context.Background())
	tid := GetTraceID(ctx)
	assert.Len(t, tid, traceLen)

	assert.Equal(t, tid, GetTraceID(ctx))
}

func TestGetTraceID_WithoutTraceID(t *testing.T) {
	tid := GetTraceID(context.Background())
	assert.Empty(t, tid)
}

func TestGetTraceID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), traceIDKey, 12345)
	assert.Empty(t, GetTraceID(ctx))
}
