package file

import (
	"bytes"
	"os"
)

func FileContain(filePath string, str string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	contain := bytes.Contains(content, []byte(str))
	return contain, nil
}
