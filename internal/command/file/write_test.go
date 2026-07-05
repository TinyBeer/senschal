package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertCodeIntoFile_RPCProbe(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "test.proto")
	content := "syntax = \"proto3\";\n//rpc place\n"
	require.NoError(t, os.WriteFile(fPath, []byte(content), 0o644))

	err := InsertCodeIntoFile(fPath, ReplaceProbe_RPC, "\trpc Hello(Req) returns (Resp){};")
	require.NoError(t, err)

	result, err := os.ReadFile(fPath)
	require.NoError(t, err)
	assert.Contains(t, string(result), "//rpc place")
	assert.Contains(t, string(result), "rpc Hello(Req) returns (Resp){};")
}

func TestInsertCodeIntoFile_MessageProbe(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "test.proto")
	content := "syntax = \"proto3\";\n//message place\n"
	require.NoError(t, os.WriteFile(fPath, []byte(content), 0o644))

	err := InsertCodeIntoFile(fPath, ReplaceProbe_Message, "\tmessage Foo{}")
	require.NoError(t, err)

	result, err := os.ReadFile(fPath)
	require.NoError(t, err)
	assert.Contains(t, string(result), "//message place")
	assert.Contains(t, string(result), "message Foo{}")
}

func TestInsertCodeIntoFile_ProbeNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "test.proto")
	content := "syntax = \"proto3\";\n"
	require.NoError(t, os.WriteFile(fPath, []byte(content), 0o644))

	err := InsertCodeIntoFile(fPath, ReplaceProbe_RPC, "\trpc Foo()returns(){};")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not contain insert probe")
}

func TestInsertCodeIntoFile_FileNotExist(t *testing.T) {
	err := InsertCodeIntoFile("/no/such/file.proto", ReplaceProbe_RPC, "code")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ReplaceProbe String
// ---------------------------------------------------------------------------

func TestReplaceProbe_String(t *testing.T) {
	assert.Equal(t, "rpc place", ReplaceProbe_RPC.String())
	assert.Equal(t, "message place", ReplaceProbe_Message.String())
	assert.Equal(t, "func place", ReplaceProbe_Func.String())
}

func TestReplaceProbe_String_Invalid(t *testing.T) {
	var rp ReplaceProbe = 99
	assert.Contains(t, rp.String(), "ReplaceProbe(99)")
}
