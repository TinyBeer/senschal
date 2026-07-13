package main

import (
	"context"
	"log/slog"
	"os"

	"seneschal/cmd"
	"seneschal/internal/ilog"
)

func main() {
	ilog.UseTraceLogger(os.Stdout)

	ctx := ilog.NewTraceCtx(context.Background())
	err := cmd.Execut(ctx)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to execute command",
			ilog.ErrorAttr(
				err,
			),
		)
	}
}
