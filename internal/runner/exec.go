package runner

import (
	"errors"
	"os/exec"
	"strings"
)

// ExecuteCommand runs a shell command and returns its output.
// Mockable var for testing — override in tests to avoid running real commands.
var ExecuteCommand = executeCommand

func executeCommand(command string) ([]byte, error) {
	splits := strings.Split(command, " ")
	if len(splits) == 0 {
		return nil, errors.New("can not run empty command")
	}
	c := exec.Command(splits[0], splits[1:]...)
	return c.CombinedOutput()
}
