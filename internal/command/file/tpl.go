package file

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
)

const tplTemplateDirName = "template"

func ExecuteTemplate(tplPath string, genDir string, settingPath string) error {
	tplInfo, err := os.Stat(tplPath)
	if err != nil {
		return err
	}
	if !tplInfo.IsDir() {
		return fmt.Errorf("execute template[%s] failed, err: [%s] is not a dir", tplPath, tplPath)
	}

	_, err = os.Stat(settingPath)
	if err != nil {
		return err
	}

	// 模板源目录：tplPath/template/
	tplSrcDir := filepath.Join(tplPath, tplTemplateDirName)
	if _, err := os.Stat(tplSrcDir); err != nil {
		return fmt.Errorf("template source dir [%s] not found: %w", tplSrcDir, err)
	}

	// 读取配置
	v := viper.New()
	v.SetConfigName(filepath.Base(settingPath))
	v.SetConfigType(Ext_TOML)
	v.AddConfigPath(filepath.Dir(settingPath))
	err = v.ReadInConfig()
	if err != nil {
		return err
	}
	setting := v.AllSettings()

	// 遍历 template 目录，收集文件和空目录
	var tplFiles []string
	var emptyDirs []string
	err = filepath.Walk(tplSrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			entries, readErr := os.ReadDir(path)
			if readErr != nil {
				return readErr
			}
			if len(entries) == 0 {
				rel, _ := filepath.Rel(tplSrcDir, path)
				emptyDirs = append(emptyDirs, rel)
			}
			return nil
		}
		tplFiles = append(tplFiles, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk template dir: %w", err)
	}

	if len(tplFiles) == 0 && len(emptyDirs) == 0 {
		return fmt.Errorf("execute template[%s] failed: nothing found in [%s]", tplPath, tplSrcDir)
	}

	// 解析模板
	tplName := filepath.Base(tplPath)
	tpl, err := template.New(tplName).ParseFiles(tplFiles...)
	if err != nil {
		return err
	}

	// 创建输出目录
	err = os.MkdirAll(genDir, 0755)
	if err != nil {
		return err
	}

	// 渲染模板文件并保留相对路径
	for _, path := range tplFiles {
		relPath, err := filepath.Rel(tplSrcDir, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for [%s]: %w", path, err)
		}

		genPath := filepath.Join(genDir, relPath)
		if err := os.MkdirAll(filepath.Dir(genPath), 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(genPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		err = tpl.ExecuteTemplate(f, filepath.Base(path), setting)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("failed to execute template [%s]: %w", relPath, err)
		}
	}

	// 创建空目录
	for _, dir := range emptyDirs {
		if err := os.MkdirAll(filepath.Join(genDir, dir), 0755); err != nil {
			return err
		}
	}

	return nil
}
