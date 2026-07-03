package ilog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newJSONHandler 创建写入 bytes.Buffer 的 JSON handler 用于捕获日志输出。
func newJSONHandler(w *bytes.Buffer, minLevel slog.Level) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: minLevel})
}

func TestNewTraceHandler(t *testing.T) {
	var buf bytes.Buffer
	next := slog.NewJSONHandler(&buf, nil)
	h := NewTraceHandler(next)
	require.NotNil(t, h)
}

func TestTraceHandler_Enabled(t *testing.T) {
	var buf bytes.Buffer
	debugH := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	errorH := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})

	h := NewTraceHandler(debugH)
	assert.True(t, h.Enabled(context.Background(), slog.LevelDebug))
	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))

	h2 := NewTraceHandler(errorH)
	assert.False(t, h2.Enabled(context.Background(), slog.LevelDebug))
	assert.True(t, h2.Enabled(context.Background(), slog.LevelError))
}

func TestTraceHandler_Handle_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	h := NewTraceHandler(newJSONHandler(&buf, slog.LevelDebug))

	ctx := NewTraceCtx(context.Background())
	expectedTID := GetTraceID(ctx)

	require.NoError(t, h.Handle(ctx, slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)))

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.Equal(t, expectedTID, m["trace_id"])
	assert.Equal(t, "msg", m["msg"])
}

func TestTraceHandler_Handle_WithoutTraceID(t *testing.T) {
	var buf bytes.Buffer
	h := NewTraceHandler(newJSONHandler(&buf, slog.LevelDebug))

	require.NoError(t, h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)))

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.NotContains(t, m, "trace_id")
	assert.Equal(t, "msg", m["msg"])
}

func TestTraceHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	next := newJSONHandler(&buf, slog.LevelDebug)
	h := NewTraceHandler(next).WithAttrs([]slog.Attr{slog.String("k1", "v1")})

	require.NoError(t, h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)))

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.Equal(t, "v1", m["k1"])
}

func TestTraceHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	next := newJSONHandler(&buf, slog.LevelDebug)
	h := NewTraceHandler(next).WithGroup("g")

	rec := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
	rec.AddAttrs(slog.String("k", "v"))
	require.NoError(t, h.Handle(context.Background(), rec))

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.Equal(t, map[string]any{"k": "v"}, m["g"])
}
