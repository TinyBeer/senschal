package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand_Echo(t *testing.T) {
	out, err := ExecuteCommand("echo hello world")
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", string(out))
}

func TestExecuteCommand_EmptyCommand(t *testing.T) {
	_, err := ExecuteCommand("")
	assert.Error(t, err)
}

func TestExecuteCommand_NonExistentCommand(t *testing.T) {
	_, err := ExecuteCommand("no_such_command_xyz")
	assert.Error(t, err)
}

func TestExecuteCommand_MultipleArgs(t *testing.T) {
	out, err := ExecuteCommand("echo a b c")
	require.NoError(t, err)
	assert.Contains(t, string(out), "a b c")
}

func TestExecuteCommand_CommandNotFound(t *testing.T) {
	_, err := ExecuteCommand("/no/such/binary --flag")
	assert.Error(t, err)
}

func TestExecuteCommand_WithOutput(t *testing.T) {
	out, err := ExecuteCommand("echo hello")
	require.NoError(t, err)
	assert.Contains(t, string(out), "hello")
}
