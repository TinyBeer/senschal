package util

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// 保存字符串到文件
func SaveStringToFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// matchExt 检查文件扩展名是否匹配给定的扩展名列表。
// exts 中的扩展名可以带或不带前置点号。
func matchExt(ext string, exts []string) bool {
	for _, e := range exts {
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		if ext == e {
			return true
		}
	}
	return false
}

// ListDirOneLevelFile 列出文件夹下一层所有文件，返回相对路径（文件名）列表。
// 可通过 exts 筛选扩展名，不传 exts 时返回所有文件。
func ListDirOneLevelFile(dir string, exts ...string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if len(exts) > 0 && !matchExt(filepath.Ext(entry.Name()), exts) {
			continue
		}
		files = append(files, entry.Name())
	}
	return files, nil
}

// ListDirOneLevelDir 列出文件夹下一层所有文件夹，返回相对路径（目录名）列表。
func ListDirOneLevelDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

// ListDilrRecursiveFile 递归列出文件夹下所有文件，返回相对于 dir 的路径列表。
// 可通过 exts 筛选扩展名，不传 exts 时返回所有文件。
func ListDilrRecursiveFile(dir string, exts ...string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: dir, Err: errors.New("not a directory")}
	}

	var files []string

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if len(exts) > 0 && !matchExt(filepath.Ext(d.Name()), exts) {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ListDirRecursiveDir 递归列出文件夹下所有文件夹，返回相对于 dir 的路径列表。
func ListDirRecursiveDir(dir string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: dir, Err: errors.New("not a directory")}
	}

	var dirs []string

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == dir {
			return nil // 跳过根目录自身
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		dirs = append(dirs, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}
