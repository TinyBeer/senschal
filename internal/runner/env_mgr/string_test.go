package envmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment_String(t *testing.T) {
	assert.Equal(t, "docker", Environment_Docker.String())
}

func TestEnvironment_String_Invalid(t *testing.T) {
	var e Environment = 99
	assert.Contains(t, e.String(), "Environment(99)")
}
