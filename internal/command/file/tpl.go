package file

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
)

func ExecuteTemplate(tplPath string, genDir string, settingPath string) error {
	name := filepath.Base(tplPath)
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

	tplFileList, err := ListFileWithoutExt(tplPath, Ext_TOML)
	if err != nil {
		return err
	}

	if len(tplFileList) == 0 {
		return fmt.Errorf("execute template[%s] failed, err: not tpl found in [%s] ", tplPath, tplPath)
	}

	tpl, err := template.New(name).ParseFiles(tplFileList...)
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigName(filepath.Base(settingPath))
	v.SetConfigType(Ext_TOML)
	v.AddConfigPath(filepath.Dir(settingPath))
	err = v.ReadInConfig()
	if err != nil {
		return err
	}
	setting := v.AllSettings()

	err = os.MkdirAll(genDir, 0755)
	if err != nil {
		return err
	}

	for _, path := range tplFileList {
		tplName := filepath.Base(path)
		genPath := filepath.Join(genDir, tplName)
		f, err := os.OpenFile(genPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		err = tpl.ExecuteTemplate(f, filepath.Base(path), setting)
		if err != nil {
			return err
		}
	}
	return nil

}
