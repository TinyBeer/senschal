package ilog

import (
	"context"
	"log/slog"
)

func NewTraceHandler(next slog.Handler) slog.Handler {
	return &TraceHandler{
		next: next,
	}
}

type TraceHandler struct {
	next slog.Handler
}

// Enabled implements [slog.Handler].
func (t *TraceHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return t.next.Enabled(ctx, l)
}

// Handle implements [slog.Handler].
func (t *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	tid := GetTraceID(ctx)
	if tid != "" {
		r.AddAttrs(slog.String(traceIDTag, tid))
	}
	return t.next.Handle(ctx, r)
}

// WithAttrs implements [slog.Handler].
func (t *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return t.next.WithAttrs(attrs)
}

// WithGroup implements [slog.Handler].
func (t *TraceHandler) WithGroup(name string) slog.Handler {
	return t.next.WithGroup(name)
}

var _ slog.Handler = (*TraceHandler)(nil)
