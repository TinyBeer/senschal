package file

import (
	"os"
	"path/filepath"
	"strings"
)

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

func ListDirName(dir string) ([]string, error) {
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
