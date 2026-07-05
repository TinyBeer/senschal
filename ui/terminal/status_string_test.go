package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkoutStatus_String(t *testing.T) {
	assert.Equal(t, "INACTIVE", WorkoutStatus_Inactive.String())
	assert.Equal(t, "ACTIVE", WorkoutStatus_Active.String())
	assert.Equal(t, "FIN", WorkoutStatus_Fin.String())
}

func TestWorkoutStatus_String_Invalid(t *testing.T) {
	var s WorkoutStatus = 99
	assert.Contains(t, s.String(), "WorkoutStatus(99)")
}
