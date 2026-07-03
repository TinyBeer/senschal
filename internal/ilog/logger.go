package ilog

import (
	"io"
	"log/slog"
)

var logLevel = new(slog.LevelVar)

func SetLogLevel(l slog.Level) {
	logLevel.Set(l)
}

func NewTraceLogger(writer io.Writer) *slog.Logger {
	return slog.New(NewTraceHandler(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	})))
}

func UseTraceLogger(w io.Writer) {
	logger := NewTraceLogger(w)
	slog.SetDefault(logger)
}
