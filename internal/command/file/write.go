package file

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func InsertCodeIntoFile(filePath string, probe ReplaceProbe, codes ...string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	old := []byte("//" + probe.String())
	code := strings.Join(codes, "\n")
	code = "\n" + code
	if probe == ReplaceProbe_Message {
		code = "\n" + code
	}
	new := append(old, []byte(code)...)

	contain := bytes.Contains(content, old)
	if !contain {
		return fmt.Errorf("file[%s] not contain insert probe[%s], please manually add first", filePath, probe.String())
	}
	newContent := bytes.Replace(content, old, new, 1)

	return os.WriteFile(filePath, newContent, 0644)
}
