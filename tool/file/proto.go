package file

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/emicklei/proto"
)

type ProtoInfo struct {
	Path        string
	MessageList []string
	RPCList     []string
}

func InsertCodeIntoProto(filePath string, probe ReplaceProbe, code string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	old := []byte("//" + probe.String())
	new := append([]byte(code), old...)

	contain := bytes.Contains(content, old)
	if !contain {
		return fmt.Errorf("file[%s] not contain insert probe[%s], please manually add first", filePath, probe.String())
	}
	newContent := bytes.Replace(content, old, new, 1)

	return os.WriteFile(filePath, newContent, 0644)
}

func ParseProtoFile(filePath string) (*ProtoInfo, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	info := &ProtoInfo{Path: filePath}
	proto.Walk(definition,
		proto.WithMessage(func(m *proto.Message) {
			info.MessageList = append(info.MessageList, m.Name)
		}),
		proto.WithRPC(func(r *proto.RPC) {
			info.RPCList = append(info.RPCList, r.Name)
		}))

	return info, nil
}
