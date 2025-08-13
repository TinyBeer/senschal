package file

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FileContain(filePath string, str string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	contain := bytes.Contains(content, []byte(str))
	return contain, nil
}

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

func ListFileWithExt(dir string, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(info.Name()) == "."+ext {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ListFileNameWithExt(dir string, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == "."+ext {
			files = append(files, strings.TrimSuffix(info.Name(), "."+ext))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ListFileWithoutExt(dir string, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(info.Name()) != "."+ext {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ListDir(dir string) ([]string, error) {
	infoList, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dirList []string
	for _, info := range infoList {
		if info.IsDir() {
			dirList = append(dirList, info.Name())
		}
	}
	return dirList, nil
}
