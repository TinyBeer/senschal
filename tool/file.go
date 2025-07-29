package tool

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"seneschal/config"
	"strconv"
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

	Name         string
	BasePath     string
	RelativePath string
	Children     []*Path
	Size         int64
}

func (p *Path) Show() {
	fmt.Printf("%s %d %t\n", filepath.Join(p.BasePath, p.RelativePath, p.Name), p.Size, p.Type == PathType_Dir)
	for _, child := range p.Children {
		child.Show()
	}
}

func (p *Path) syncStat() error {
	_, err := p.IsDir()
	if err != nil {
		return err
	}
	size, err := p.getSize()
	if err != nil {
		return err
	}
	p.Size = size
	if p.Type == PathType_Dir {
		dirName := filepath.Join(p.BasePath, p.RelativePath, p.Name)
		if p.Remote == nil {
			entries, err := os.ReadDir(dirName)
			if err != nil {
				return err
			}
			relativePath := filepath.Join(p.RelativePath, p.Name)
			for _, e := range entries {
				subP := &Path{
					Type:         PathType_File,
					Remote:       p.Remote,
					Name:         e.Name(),
					BasePath:     p.BasePath,
					RelativePath: relativePath,
					Children:     nil,
				}
				if e.IsDir() {
					subP.Type = PathType_Dir
				}
				p.Children = append(p.Children, subP)
			}
		} else {
			se, err := NewSSHExecutor(p.Remote.SSH)
			if err != nil {
				return err
			}
			session, err := se.NewSession()
			if err != nil {
				return err
			}
			defer session.Close()

			// 执行 stat 命令获取文件信息
			output, err := session.CombinedOutput(fmt.Sprintf("ls -p " + dirName))
			if err != nil {
				return err
			}

			result := strings.TrimSpace(string(output))
			fmt.Println(result)
		}
		for _, child := range p.Children {
			child.syncStat()
		}
	}
	return nil
}

func (p *Path) IsExist() (bool, error) {
	filePath := filepath.Join(p.BasePath, p.RelativePath, p.Name)
	if p.Remote == nil {
		_, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	} else {
		se, err := NewSSHExecutor(p.Remote.SSH)
		if err != nil {
			return false, err
		}

		session, err := se.newSession()
		if err != nil {
			return false, err
		}
		defer session.Close()

		cmd := fmt.Sprintf("if stat %s &> /dev/null; then echo exists; fi", filePath)
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return false, err
		}
		return string(output) == "exists\n", nil
	}
}

func (p *Path) IsDir() (bool, error) {
	if p.Type != PathType_Unknown {
		return p.Type == PathType_Dir, nil
	}
	name := filepath.Join(p.BasePath, p.RelativePath, p.Name)
	if p.Remote == nil {
		fileInfo, err := os.Stat(name)
		if err != nil {
			return false, err
		}
		if fileInfo.IsDir() {
			p.Type = PathType_Dir
		} else {
			p.Type = PathType_File
		}
	} else {
		se, err := NewSSHExecutor(p.Remote.SSH)
		if err != nil {
			return false, err
		}

		session, err := se.newSession()
		if err != nil {
			return false, err
		}
		defer session.Close()

		// 执行命令获取文件大小
		output, err := session.CombinedOutput("stat -c %F " + name)
		if err != nil {
			return false, err
		}

		// 处理输出结果
		result := strings.TrimSpace(string(output))
		if strings.Contains(result, "directory") {
			p.Type = PathType_Dir
		} else if strings.Contains(result, "regular file") {
			p.Type = PathType_File
		} else {
			return false, errors.New("unknown path type")
		}
	}
	return p.Type == PathType_Dir, nil
}

func (p *Path) getSize() (int64, error) {
	name := filepath.Join(p.BasePath, p.RelativePath, p.Name)
	var size int64
	if p.Remote == nil {
		fileInfo, err := os.Stat(name)
		if err != nil {
			return 0, err
		}
		size = fileInfo.Size()
	} else {
		se, err := NewSSHExecutor(p.Remote.SSH)
		if err != nil {
			return 0, err
		}

		session, err := se.newSession()
		if err != nil {
			return 0, err
		}
		defer session.Close()

		// 执行命令获取文件大小
		output, err := session.CombinedOutput("stat -c %s " + name)
		if err != nil {
			return 0, err
		}

		// 处理输出结果
		sizeStr := strings.TrimSpace(string(output))
		size, err = strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return size, nil
}

func newPath(path string) (*Path, error) {
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
				BasePath: splited[1],
			}
		} else {
			return nil, fmt.Errorf("未找到%s的配置信息", alias)
		}
	} else {
		p = &Path{
			Remote:   nil,
			BasePath: path,
			Name:     "",
		}
	}
	return p, nil
}

func NewPath(path string) (*Path, error) {
	return newPath(path)
}

func (p *Path) NewFile(fileName string) (*File, error) {
	return &File{
		Remote: p.Remote,
		Path:   filepath.Join(p.BasePath, p.RelativePath, p.Name, fileName),
	}, nil
}

type File struct {
	Remote *Remote
	Path   string
}

// GetSize implements IFile.
func (f *File) GetSize() (int64, error) {
	var size int64
	if f.Remote == nil {
		fileInfo, err := os.Stat(f.Path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("文件不存在")
			} else {
				fmt.Printf("读取文件信息时出错: %v\n", err)
			}
			return 0, err
		}
		size = fileInfo.Size()
	} else {
		se, err := NewSSHExecutor(f.Remote.SSH)
		if err != nil {
			return 0, err
		}

		session, err := se.newSession()
		if err != nil {
			return 0, err
		}
		defer session.Close()

		// 执行命令获取文件大小
		output, err := session.CombinedOutput("stat -c %s " + f.Path)
		if err != nil {
			return 0, err
		}

		// 处理输出结果
		sizeStr := strings.TrimSpace(string(output))
		size, err = strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return size, nil
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
func (f *File) GetWriter() (io.WriteCloser, error) {
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
	GetWriter() (io.WriteCloser, error)
	GetSize() (int64, error)
}

var _ IFile = new(File)
