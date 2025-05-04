package tool

import (
	"errors"
	"os/exec"
	"strings"
)

func ExecuteCommand(command string) ([]byte, error) {
	splits := strings.Split(command, " ")
	if len(splits) == 0 {
		return nil, errors.New("can not run empty command")
	}
	c := exec.Command(splits[0], splits[1:]...)
	return c.CombinedOutput()
}
