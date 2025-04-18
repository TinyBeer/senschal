package tool

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"seneschal/config"
	"strings"
)

type Remote struct {
	SSH *config.SSHConfig
}

type PathType uint8

const (
	PathType_Unknown PathType = iota
	PathType_File
	PathType_Dir
)

type Path struct {
	Type   PathType
	Remote *Remote
	Path   string
}

func NewPath(path string) (*Path, error) {
	splited := strings.Split(path, ":")
	if len(splited) > 2 {
		return nil, errors.New("invalid path: " + path)
	}
	var p *Path
	if len(splited) == 2 {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			return nil, err
		}
		alias := splited[0]
		if sc, ok := m[alias]; ok {
			p = &Path{
				Remote: &Remote{
					SSH: sc,
				},
				Path: splited[1],
			}
		} else {
			return nil, fmt.Errorf("未找到%s的配置信息", alias)
		}
		se, err := NewSSHExecutor(p.Remote.SSH)
		if err != nil {
			return nil, err
		}
		session, err := se.NewSession()
		if err != nil {
			return nil, err
		}
		defer session.Close()

		// 执行 stat 命令获取文件信息
		output, err := session.CombinedOutput(fmt.Sprintf("stat -c %%F %s", p.Path))
		if err != nil {
			return nil, err
		}

		result := strings.TrimSpace(string(output))
		if strings.Contains(result, "directory") {
			p.Type = PathType_Dir
		} else if strings.Contains(result, "regular file") {
			p.Type = PathType_File
		}
	} else {
		p = &Path{
			Remote: nil,
			Path:   path,
		}
		// 获取文件信息
		fileInfo, err := os.Stat(p.Path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}

		if fileInfo.IsDir() {
			p.Type = PathType_Dir
		} else {
			p.Type = PathType_File
		}
	}
	return p, nil
}

func (p *Path) NewFile(fileName string) (*File, error) {
	return &File{
		Remote: p.Remote,
		Path:   filepath.Join(p.Path, fileName),
	}, nil
}

type File struct {
	Remote *Remote
	Path   string
}

func NewFile(filePath string) (*File, error) {
	splited := strings.Split(filePath, ":")
	if len(splited) > 2 {
		return nil, errors.New("invalid file path: " + filePath)
	}
	if len(splited) == 2 {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			return nil, err
		}
		alias := splited[0]
		if sc, ok := m[alias]; ok {
			return &File{
				Remote: &Remote{
					SSH: sc,
				},
				Path: splited[1],
			}, nil
		} else {
			return nil, fmt.Errorf("未找到%s的配置信息", alias)
		}
	}
	return &File{
		Remote: nil,
		Path:   filePath,
	}, nil

}

// GetReader implements IFile.
func (f *File) GetReader() (io.Reader, error) {
	if f.Remote == nil {
		return os.Open(f.Path)
	}
	return nil, errors.New("remote file reader is not supported yet")
}

// GetWriter implements IFile.
func (f *File) GetWriter() (io.Writer, error) {
	if f.Remote == nil {
		return nil, errors.New("loacl writer is not supported yet")
	}

	se, err := NewSSHExecutor(f.Remote.SSH)
	if err != nil {
		return nil, err
	}

	session, err := se.newSession()
	if err != nil {
		return nil, err
	}
	out, err := se.ExecuteCommand("mkdir -p " + filepath.Dir(f.Path))
	if err != nil {
		log.Println(string(out))
		return nil, err
	}

	// 创建远程文件的写入器
	writer, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	// 启动远程命令来接收文件
	go func() {
		cmd := fmt.Sprintf("cat > %s", f.Path)
		if err := session.Run(cmd); err != nil {
			log.Fatalf("Failed to run remote command: %v", err)
		}
	}()
	return writer, nil
}

type IFile interface {
	GetReader() (io.Reader, error)
	GetWriter() (io.Writer, error)
}

var _ IFile = new(File)
