package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkoutType_String(t *testing.T) {
	tests := []struct {
		name string
		wt   WorkoutType
		want string
	}{
		{name: "duration", wt: WorkoutType_Duration, want: "duration"},
		{name: "count", wt: WorkoutType_Count, want: "count"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.wt.String()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewWorkoutConfig(t *testing.T) {
	cfg := NewWorkoutConfig()
	require.NotNil(t, cfg)
	require.Nil(t, cfg.ItemList)
}

func TestDefaultBreak(t *testing.T) {
	require.Equal(t, 10, DefaultBreak)
}

func Test_readWorkoutConfigFromCsv(t *testing.T) {
	dir := t.TempDir()
	content := "name,type,repeat,target,break\nPush Up,1,3,30,10\nSquat,2,2,15,5\n"
	err := os.WriteFile(filepath.Join(dir, "full_body.csv"), []byte(content), 0o644)
	require.NoError(t, err)

	wc, err := readWorkoutConfigFromCsv(dir, "full_body")
	require.NoError(t, err)
	require.Equal(t, "full_body", wc.Name)
	require.Len(t, wc.ItemList, 2)

	item1 := wc.ItemList[0]
	require.Equal(t, "Push Up", item1.Name)
	require.Equal(t, WorkoutType_Duration, item1.Type)
	require.Equal(t, 3, item1.Repeat)
	require.Equal(t, 30, item1.Target)
	require.Equal(t, 10, item1.Break)

	item2 := wc.ItemList[1]
	require.Equal(t, "Squat", item2.Name)
	require.Equal(t, WorkoutType_Count, item2.Type)
	require.Equal(t, 2, item2.Repeat)
	require.Equal(t, 15, item2.Target)
	require.Equal(t, 5, item2.Break)
}

func Test_readWorkoutConfigFromCsv_RepeatDefaultsToOne(t *testing.T) {
	dir := t.TempDir()
	content := "name,type,repeat,target,break\nJumping Jack,1,0,20,5\n"
	err := os.WriteFile(filepath.Join(dir, "test.csv"), []byte(content), 0o644)
	require.NoError(t, err)

	wc, err := readWorkoutConfigFromCsv(dir, "test")
	require.NoError(t, err)
	require.Len(t, wc.ItemList, 1)
	require.Equal(t, 1, wc.ItemList[0].Repeat, "Repeat=0 should default to 1")
}

func Test_readWorkoutConfigFromCsv_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := readWorkoutConfigFromCsv(dir, "nonexistent")
	require.Error(t, err)
}

func Test_GetWorkoutConfigMap(t *testing.T) {
	dir := t.TempDir()

	content1 := "name,type,repeat,target,break\nPush Up,1,3,30,10\n"
	err := os.WriteFile(filepath.Join(dir, "upper.csv"), []byte(content1), 0o644)
	require.NoError(t, err)

	content2 := "name,type,repeat,target,break\nSquat,2,2,15,5\n"
	err = os.WriteFile(filepath.Join(dir, "lower.csv"), []byte(content2), 0o644)
	require.NoError(t, err)

	m, err := GetWorkoutConfigMap(dir)
	require.NoError(t, err)
	require.Len(t, m, 2)

	upper, ok := m["upper"]
	require.True(t, ok)
	require.Len(t, upper.ItemList, 1)
	require.Equal(t, "Push Up", upper.ItemList[0].Name)

	lower, ok := m["lower"]
	require.True(t, ok)
	require.Len(t, lower.ItemList, 1)
	require.Equal(t, "Squat", lower.ItemList[0].Name)
}

func Test_GetWorkoutConfigMap_NoFiles(t *testing.T) {
	dir := t.TempDir()
	m, err := GetWorkoutConfigMap(dir)
	require.NoError(t, err)
	require.Empty(t, m)
}

func Test_GetWorkoutConfigMap_NonExistentDir(t *testing.T) {
	_, err := GetWorkoutConfigMap("/path/that/does/not/exist")
	require.Error(t, err)
}
