package ilog

import (
	"io"
	"log/slog"
)

var logLevel = new(slog.LevelVar)

func SetLogLevel(l slog.Level) {
	logLevel.Set(l)
}

// NewTraceLogger 默认使用 Info 级别日志
func NewTraceLogger(writer io.Writer) *slog.Logger {
	logLevel.Set(slog.LevelInfo)
	return slog.New(NewTraceHandler(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	})))
}

func UseTraceLogger(w io.Writer) {
	logger := NewTraceLogger(w)
	slog.SetDefault(logger)
}
