package file

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emicklei/proto"
)

type ProtoInfo struct {
	Path        string
	MessageList []string
	RPCList     []string
}

type ProtoMessageOpt func() string

func ProtoMessageWithField(typ, name string) ProtoMessageOpt {
	return func() string {
		return fmt.Sprintf("%s %s", typ, name)
	}
}

func GenerateMessage(name string, fields ...ProtoMessageOpt) string {
	var filedList []string
	for i, f := range fields {
		filedList = append(filedList, fmt.Sprintf("\t%s = %d;", f(), i+1))
	}
	if len(fields) == 0 {
		return fmt.Sprintf("message %s{}\n", name)
	}
	return fmt.Sprintf("message %s{\n%s\n}\n", name, strings.Join(filedList, "\n"))
}

func GenerateRPC(name string, reqs []string, rets []string) string {
	return fmt.Sprintf("\trpc %s(%s) returns(%s){};", name, strings.Join(reqs, ","), strings.Join(rets, ","))
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
