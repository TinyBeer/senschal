package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestReadAllFromFile_Success 正常读取配置文件
func TestReadAllFromFile_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// 构造 yaml 文件
	fileName := "app.yaml"
	fullPath := filepath.Join(tmpDir, fileName)
	content := `
server:
  addr: 127.0.0.1
  port: 8080
mysql:
  max_conn: 20
`
	err := os.WriteFile(fullPath, []byte(content), 0o644)
	assert.NoError(t, err)

	// 调用目标函数
	settings, err := readAllFromFile(fullPath, "yaml")
	assert.NoError(t, err)

	// 校验返回 map 数据
	server, ok := settings["server"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "127.0.0.1", server["addr"])
	assert.Equal(t, 8080, server["port"])

	mysql, ok := settings["mysql"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, 20, mysql["max_conn"])
}

// TestReadAllFromFile_FileNotFound 文件不存在场景
func TestReadAllFromFile_FileNotFound(t *testing.T) {
	// 随便传一个不存在路径
	_, err := readAllFromFile("./not_exist_123456.yaml", "yaml")
	assert.Error(t, err)
	// viper 文件不存在标准错误判断
	var configErr viper.ConfigFileNotFoundError
	assert.ErrorAs(t, err, &configErr)
}

// TestReadAllFromFile_WrongConfigType 文件类型不匹配，解析失败
func TestReadAllFromFile_WrongConfigType(t *testing.T) {
	tmpDir := t.TempDir()

	fullPath := filepath.Join(tmpDir, "test.yaml")
	content := `name: test`
	_ = os.WriteFile(fullPath, []byte(content), 0o644)

	// yaml 文件指定 type=json，必然解析报错
	_, err := readAllFromFile(fullPath, "json")
	assert.Error(t, err)
}

// TestReadAllFromFile_JsonFile 兼容 json 格式测试
func TestReadAllFromFile_JsonFile(t *testing.T) {
	tmpDir := t.TempDir()

	fullPath := filepath.Join(tmpDir, "conf.json")
	content := `{"log":{"level":"info"}}`
	_ = os.WriteFile(fullPath, []byte(content), 0o644)

	settings, err := readAllFromFile(fullPath, "json")
	assert.NoError(t, err)

	log, ok := settings["log"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "info", log["level"])
}
