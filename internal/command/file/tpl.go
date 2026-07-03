package file

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
)

const tplTemplateDirName = "template"

func ExecuteTemplate(tplPath, genDir, settingPath string) error {
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
	tpl := template.New("root") // 根容器，无实际渲染用途

	err = filepath.Walk(tplSrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		tplFiles = append(tplFiles, path)

		// 关键：模板名 = 文件相对路径，全局唯一
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// 解析模板
		_, err = tpl.New(path).Parse(string(content))
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to walk template dir: %w", err)
	}

	if len(tplFiles) == 0 {
		return fmt.Errorf("execute template[%s] failed: nothing found in [%s]", tplPath, tplSrcDir)
	}

	// 渲染模板文件并保留相对路径
	for _, path := range tplFiles {
		relPath, err := filepath.Rel(tplSrcDir, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for [%s]: %w", path, err)
		}

		genPath := filepath.Join(genDir, relPath)
		if err := os.MkdirAll(filepath.Dir(genPath), 0o755); err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(genPath), 0o755)
		if err != nil {
			return fmt.Errorf("failed to make gen path dir, err: %w", err)
		}
		f, err := os.OpenFile(genPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}

		err = tpl.ExecuteTemplate(f, path, setting)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("failed to execute template [%s]: %w", relPath, err)
		}
	}

	return nil
}
