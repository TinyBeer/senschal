package config

import (
	"os"
	"path/filepath"
	"strings"
)

func ListFilesWithExt(dir string, ext string) ([]string, error) {
	var tomlFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == "."+ext {
			tomlFiles = append(tomlFiles, strings.TrimSuffix(info.Name(), "."+ext))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tomlFiles, nil
}
