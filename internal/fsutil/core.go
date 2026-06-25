package fsutil

import (
	"errors"
	"io"
)

// ===================== 对外导出配置模型 =====================
// SSHConfig SSH连接配置，供业务层初始化使用
type SSHConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	PrivateKey     []byte
	DialTimeoutSec int // TCP+SSH握手超时
	SessTimeoutSec int // 单次session读写/命令超时
}

// ===================== 路径解析结构与方法 =====================
// PathRef 统一路径封装，格式：/local/file 、alias:/remote/file
type PathRef struct {
	IsRemote bool
	Alias    string
	RawPath  string
}

// Parse 解析路径字符串生成PathRef
func Parse(input string) (*PathRef, error) { return nil, nil }

func (p *PathRef) IsLocal() bool { return !p.IsRemote }

func (p *PathRef) Validate() error {
	if p.IsRemote && p.Alias == "" {
		return errors.New("remote path must contain ssh alias")
	}
	return nil
}

// ===================== 统一抹平本地/远端文件元数据 =====================
type FileStat struct {
	Exist bool
	IsDir bool
	Size  int64
}

type DirEntry struct {
	Name string
	Stat FileStat
}

// ===================== 顶层统一文件系统抽象接口 =====================
type FileSystem interface {
	Stat(ref *PathRef) (FileStat, error)
	OpenRead(ref *PathRef) (io.ReadCloser, error)
	OpenWrite(ref *PathRef) (io.WriteCloser, error)
	ListDir(ref *PathRef) ([]DirEntry, error)
}

// ===================== 内部适配层（实现FileSystem，桥接localfs/sshfs逻辑） =====================
type localFS struct {
	cli *localClient
}

var _ FileSystem = (*localFS)(nil)

func (l *localFS) Stat(ref *PathRef) (FileStat, error)            { return FileStat{}, nil }
func (l *localFS) OpenRead(ref *PathRef) (io.ReadCloser, error)   { return nil, nil }
func (l *localFS) OpenWrite(ref *PathRef) (io.WriteCloser, error) { return nil, nil }
func (l *localFS) ListDir(ref *PathRef) ([]DirEntry, error)       { return nil, nil }

type remoteFS struct {
	cli *sshClient
}

var _ FileSystem = (*remoteFS)(nil)

func (r *remoteFS) Stat(ref *PathRef) (FileStat, error)            { return FileStat{}, nil }
func (r *remoteFS) OpenRead(ref *PathRef) (io.ReadCloser, error)   { return nil, nil }
func (r *remoteFS) OpenWrite(ref *PathRef) (io.WriteCloser, error) { return nil, nil }
func (r *remoteFS) ListDir(ref *PathRef) ([]DirEntry, error)       { return nil, nil }

// ===================== 工厂方法：自动匹配本地/远端FileSystem =====================
func GetFS(ref *PathRef, sshConfMap map[string]*SSHConfig) (FileSystem, error) {
	if ref.IsLocal() {
		return &localFS{cli: newLocalClient()}, nil
	}
	conf, ok := sshConfMap[ref.Alias]
	if !ok {
		return nil, errors.New("ssh alias not found in config map")
	}
	innerConf := &sshClientConfig{
		Host:           conf.Host,
		Port:           conf.Port,
		User:           conf.User,
		Password:       conf.Password,
		PrivateKey:     conf.PrivateKey,
		DialTimeoutSec: conf.DialTimeoutSec,
		SessTimeoutSec: conf.SessTimeoutSec,
	}
	cli, err := newSSHClient(innerConf)
	if err != nil {
		return nil, err
	}
	return &remoteFS{cli: cli}, nil
}

// ===================== 传输管理器：业务上传/下载/目录拷贝入口 =====================
type Transfer struct {
	sshConfigMap map[string]*SSHConfig
}

func NewTransfer(sshConfMap map[string]*SSHConfig) *Transfer {
	return &Transfer{sshConfigMap: sshConfMap}
}

// CopyFile 通用单文件互拷：本地↔远端、远端↔远端
func (t *Transfer) CopyFile(srcRef, dstRef *PathRef) error {
	srcFS, err := GetFS(srcRef, t.sshConfigMap)
	if err != nil {
		return err
	}
	dstFS, err := GetFS(dstRef, t.sshConfigMap)
	if err != nil {
		return err
	}

	r, err := srcFS.OpenRead(srcRef)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := dstFS.OpenWrite(dstRef)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

// CopyDir 递归完整拷贝目录
func (t *Transfer) CopyDir(srcRef, dstRef *PathRef) error { return nil }

// Upload 快捷API：本地文件上传SSH远端
func (t *Transfer) Upload(localPath, remotePath string) error {
	src, err := Parse(localPath)
	if err != nil {
		return err
	}
	dst, err := Parse(remotePath)
	if err != nil {
		return err
	}
	return t.CopyFile(src, dst)
}

// Download 快捷API：SSH远端文件下载本地
func (t *Transfer) Download(remotePath, localPath string) error {
	src, err := Parse(remotePath)
	if err != nil {
		return err
	}
	dst, err := Parse(localPath)
	if err != nil {
		return err
	}
	return t.CopyFile(src, dst)
}
