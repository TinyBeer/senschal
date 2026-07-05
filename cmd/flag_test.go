package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func newTestFlagCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("flag-bool", false, "test bool flag")
	cmd.Flags().Int("flag-int", 0, "test int flag")
	cmd.Flags().String("flag-str", "", "test string flag")
	return cmd
}

func TestGetBoolFlag(t *testing.T) {
	cmd := newTestFlagCmd()

	t.Run("默认值", func(t *testing.T) {
		v, err := getBoolFlag(cmd, "flag-bool")
		require.NoError(t, err)
		require.False(t, v)
	})

	t.Run("解析已设置的值", func(t *testing.T) {
		err := cmd.Flags().Set("flag-bool", "true")
		require.NoError(t, err)
		v, err := getBoolFlag(cmd, "flag-bool")
		require.NoError(t, err)
		require.True(t, v)
	})

	t.Run("不存在的flag返回错误", func(t *testing.T) {
		_, err := getBoolFlag(cmd, "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent")
	})
}

func TestGetIntFlag(t *testing.T) {
	cmd := newTestFlagCmd()

	t.Run("默认值", func(t *testing.T) {
		v, err := getIntFlag(cmd, "flag-int")
		require.NoError(t, err)
		require.Equal(t, 0, v)
	})

	t.Run("解析已设置的值", func(t *testing.T) {
		err := cmd.Flags().Set("flag-int", "42")
		require.NoError(t, err)
		v, err := getIntFlag(cmd, "flag-int")
		require.NoError(t, err)
		require.Equal(t, 42, v)
	})

	t.Run("不存在的flag返回错误", func(t *testing.T) {
		_, err := getIntFlag(cmd, "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent")
	})
}

func TestGetStringFlag(t *testing.T) {
	cmd := newTestFlagCmd()

	t.Run("默认值", func(t *testing.T) {
		v, err := getStringFlag(cmd, "flag-str")
		require.NoError(t, err)
		require.Equal(t, "", v)
	})

	t.Run("解析已设置的值", func(t *testing.T) {
		err := cmd.Flags().Set("flag-str", "hello")
		require.NoError(t, err)
		v, err := getStringFlag(cmd, "flag-str")
		require.NoError(t, err)
		require.Equal(t, "hello", v)
	})

	t.Run("不存在的flag返回错误", func(t *testing.T) {
		_, err := getStringFlag(cmd, "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent")
	})
}
