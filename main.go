package main

import (
	"os"
	"seneschal/cmd"
	"seneschal/internal/ilog"
)

func main() {

	ilog.UseTraceLogger(os.Stdout)

	cmd.Execut()
}
