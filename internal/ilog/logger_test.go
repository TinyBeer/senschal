package ilog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTraceLogger_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTraceLogger(&buf)

	ctx := NewTraceCtx(context.Background())
	expectedTID := GetTraceID(ctx)

	logger.LogAttrs(ctx, slog.LevelInfo, "hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.Equal(t, expectedTID, m["trace_id"])
	assert.Equal(t, "hello", m["msg"])
}

func TestNewTraceLogger_WithoutTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTraceLogger(&buf)

	logger.LogAttrs(context.Background(), slog.LevelInfo, "hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.NotContains(t, m, "trace_id")
	assert.Equal(t, "hello", m["msg"])
}

func TestNewTraceLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTraceLogger(&buf)

	logger.LogAttrs(context.Background(), slog.LevelDebug, "should-not-appear")
	assert.Empty(t, buf.Bytes(), "Debug log should be filtered when level is Info")
}

func TestSetLogLevel(t *testing.T) {
	prev := logLevel.Level()
	t.Cleanup(func() { logLevel.Set(prev) })

	var buf bytes.Buffer
	logger := NewTraceLogger(&buf)

	logger.LogAttrs(context.Background(), slog.LevelDebug, "silent")
	assert.Empty(t, buf.Bytes())

	SetLogLevel(slog.LevelDebug)
	logger.LogAttrs(context.Background(), slog.LevelDebug, "visible")

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.Equal(t, "visible", m["msg"])
}

func TestUseTraceLogger(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var buf bytes.Buffer
	UseTraceLogger(&buf)

	ctx := NewTraceCtx(context.Background())
	expectedTID := GetTraceID(ctx)

	slog.LogAttrs(ctx, slog.LevelInfo, "global")

	var m map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	assert.Equal(t, expectedTID, m["trace_id"])
	assert.Equal(t, "global", m["msg"])
}
