package main

import (
	"log/slog"
	"os"
	"seneschal/cmd"
	"seneschal/internal/ilog"
)

func main() {

	ilog.UseTraceLogger(os.Stdout)

	err := cmd.Execut()
	if err != nil {
		slog.Error(
			"failed to execute command",
			ilog.ErrorAttr(
				err,
			))
	}
}
