package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ProtoMessageWithField
// ---------------------------------------------------------------------------

func TestProtoMessageWithField(t *testing.T) {
	opt := ProtoMessageWithField("string", "name")
	assert.Equal(t, "string name", opt())
}

func TestGenerateMessage_WithFields(t *testing.T) {
	result := GenerateMessage("Foo",
		ProtoMessageWithField("string", "name"),
		ProtoMessageWithField("int32", "age"),
	)
	expected := "message Foo{\n\tstring name = 1;\n\tint32 age = 2;\n}\n"
	assert.Equal(t, expected, result)
}

func TestGenerateMessage_NoFields(t *testing.T) {
	result := GenerateMessage("EmptyMsg")
	assert.Equal(t, "message EmptyMsg{}\n", result)
}

func TestGenerateMessage_SingleField(t *testing.T) {
	result := GenerateMessage("Bar", ProtoMessageWithField("bool", "flag"))
	assert.Equal(t, "message Bar{\n\tbool flag = 1;\n}\n", result)
}

// ---------------------------------------------------------------------------
// GenerateRPC
// ---------------------------------------------------------------------------

func TestGenerateRPC(t *testing.T) {
	result := GenerateRPC("Hello", []string{"Req"}, []string{"Resp"})
	assert.Equal(t, "\trpc Hello(Req) returns(Resp){};", result)
}

func TestGenerateRPC_Empty(t *testing.T) {
	result := GenerateRPC("Noop", []string{}, []string{})
	assert.Equal(t, "\trpc Noop() returns(){};", result)
}

func TestGenerateRPC_MultipleParams(t *testing.T) {
	result := GenerateRPC("Stream", []string{"ReqA", "ReqB"}, []string{"Resp"})
	assert.Equal(t, "\trpc Stream(ReqA,ReqB) returns(Resp){};", result)
}

// ---------------------------------------------------------------------------
// ParseProtoFile
// ---------------------------------------------------------------------------

func TestParseProtoFile(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "hello.proto")
	content := `syntax = "proto3";
package hello;

service Greeter {
	rpc SayHello (HelloReq) returns (HelloResp);
}

message HelloReq {
	string name = 1;
}

message HelloResp {
	string greeting = 1;
}
`
	require.NoError(t, os.WriteFile(fPath, []byte(content), 0o644))

	info, err := ParseProtoFile(fPath)
	require.NoError(t, err)
	assert.Equal(t, fPath, info.Path)
	assert.ElementsMatch(t, []string{"HelloReq", "HelloResp"}, info.MessageList)
	assert.ElementsMatch(t, []string{"SayHello"}, info.RPCList)
}

func TestParseProtoFile_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "invalid.proto")
	require.NoError(t, os.WriteFile(fPath, []byte("garbage content {{{"), 0o644))

	_, err := ParseProtoFile(fPath)
	assert.Error(t, err)
}
