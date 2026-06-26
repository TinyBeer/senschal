package fsutil

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"seneschal/config"
)

// ===================== 路径解析结构与方法 =====================
// PathRef 统一路径封装，格式：/local/file 、alias:/remote/file
type PathRef struct {
	IsRemote bool
	Alias    string
	RawPath  string
}

// Parse 解析路径字符串生成PathRef
func Parse(input string) (*PathRef, error) {
	if input == "" {
		return nil, errors.New("empty path")
	}
	splited := strings.Split(input, ":")
	switch len(splited) {
	case 1:
		return &PathRef{
			IsRemote: false,
			Alias:    "",
			RawPath:  input,
		}, nil
	case 2:
		return &PathRef{
			IsRemote: true,
			Alias:    splited[0],
			RawPath:  splited[1],
		}, nil
	default:
		return nil, fmt.Errorf("invalid path: %s", input)
	}
}

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
	MkdirAll(ref *PathRef) error
}

// ===================== 内部适配层（实现FileSystem，桥接localfs/sshfs逻辑） =====================
type localFS struct {
	cli *localClient
}

var _ FileSystem = (*localFS)(nil)

func (l *localFS) Stat(ref *PathRef) (FileStat, error) {
	fi, err := os.Stat(ref.RawPath)
	if err != nil {
		if os.IsNotExist(err) {
			return FileStat{
				Exist: false,
				IsDir: false,
				Size:  0,
			}, nil
		}
		return FileStat{}, err
	}
	return FileStat{
		Exist: true,
		IsDir: fi.IsDir(),
		Size:  fi.Size(),
	}, nil
}

func (l *localFS) OpenRead(ref *PathRef) (io.ReadCloser, error) {
	return os.Open(ref.RawPath)
}

func (l *localFS) OpenWrite(ref *PathRef) (io.WriteCloser, error) {
	return os.Create(ref.RawPath)
}

func (l *localFS) ListDir(ref *PathRef) ([]DirEntry, error) {
	entries, err := os.ReadDir(ref.RawPath)
	if err != nil {
		return nil, err
	}

	result := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		result = append(result, DirEntry{
			Name: entry.Name(),
			Stat: FileStat{
				Exist: true,
				IsDir: entry.IsDir(),
				Size:  info.Size(),
			},
		})
	}

	return result, nil
}

func (l *localFS) MkdirAll(ref *PathRef) error {
	return os.MkdirAll(ref.RawPath, 0o755)
}

type RemoteClient interface {
	Close() error
	Stat(remotePath string) (*RemoteStat, error)
	OpenReader(remotePath string) (io.ReadCloser, error)
	OpenWriter(remotePath string) (io.WriteCloser, error)
	ListDir(remoteDir string) ([]RemoteDirEntry, error)
	MkdirAll(remotePath string) error
}

type remoteFS struct {
	cli RemoteClient
}

var _ FileSystem = (*remoteFS)(nil)

func (r *remoteFS) Stat(ref *PathRef) (FileStat, error) {
	state, err := r.cli.Stat(ref.RawPath)
	if err != nil {
		return FileStat{}, err
	}
	return FileStat{
		Exist: state.Exist,
		IsDir: state.IsDir,
		Size:  state.Size,
	}, nil
}

func (r *remoteFS) OpenRead(ref *PathRef) (io.ReadCloser, error) {
	return r.cli.OpenReader(ref.RawPath)
}

func (r *remoteFS) OpenWrite(ref *PathRef) (io.WriteCloser, error) {
	return r.cli.OpenWriter(ref.RawPath)
}

func (r *remoteFS) ListDir(ref *PathRef) ([]DirEntry, error) {
	entries, err := r.cli.ListDir(ref.RawPath)
	if err != nil {
		return nil, err
	}

	result := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, DirEntry{
			Name: entry.Name,
			Stat: entry.Stat,
		})
	}
	return result, nil
}

func (r *remoteFS) MkdirAll(ref *PathRef) error {
	return r.cli.MkdirAll(ref.RawPath)
}

// ===================== 工厂方法：自动匹配本地/远端FileSystem =====================
func GetFS(ref *PathRef, sshConfMap map[string]*config.SSHConfig) (FileSystem, error) {
	if ref.IsLocal() {
		return &localFS{cli: newLocalClient()}, nil
	}
	conf, ok := sshConfMap[ref.Alias]
	if !ok {
		return nil, errors.New("ssh alias not found in config map")
	}

	if conf.SSH == nil {
		return nil, fmt.Errorf("alias[%s] missing ssh config", conf.Alias)
	}
	cli, err := newSSHClient(conf)
	if err != nil {
		return nil, err
	}

	return &remoteFS{cli: cli}, nil
}

// ===================== 传输管理器：业务上传/下载/目录拷贝入口 =====================
type Transfer struct {
	sshConfigMap map[string]*config.SSHConfig
}

func NewTransfer(sshConfMap map[string]*config.SSHConfig) *Transfer {
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
func (t *Transfer) CopyDir(srcRef, dstRef *PathRef) error {
	srcFS, err := GetFS(srcRef, t.sshConfigMap)
	if err != nil {
		return fmt.Errorf("get src fs: %w", err)
	}
	dstFS, err := GetFS(dstRef, t.sshConfigMap)
	if err != nil {
		return fmt.Errorf("get dst fs: %w", err)
	}

	return t.copyDirRecursive(srcFS, dstFS, srcRef, dstRef)
}

// copyDirRecursive 递归拷贝目录条目
func (t *Transfer) copyDirRecursive(srcFS, dstFS FileSystem, srcRef, dstRef *PathRef) error {
	// 先确保目标目录存在
	if err := dstFS.MkdirAll(dstRef); err != nil {
		return fmt.Errorf("mkdir destination %s: %w", dstRef.RawPath, err)
	}

	entries, err := srcFS.ListDir(srcRef)
	if err != nil {
		return fmt.Errorf("list source %s: %w", srcRef.RawPath, err)
	}

	for _, entry := range entries {
		srcChild := &PathRef{
			IsRemote: srcRef.IsRemote,
			Alias:    srcRef.Alias,
			RawPath:  joinPath(srcRef.IsRemote, srcRef.RawPath, entry.Name),
		}
		dstChild := &PathRef{
			IsRemote: dstRef.IsRemote,
			Alias:    dstRef.Alias,
			RawPath:  joinPath(dstRef.IsRemote, dstRef.RawPath, entry.Name),
		}

		if entry.Stat.IsDir {
			if err := t.copyDirRecursive(srcFS, dstFS, srcChild, dstChild); err != nil {
				return fmt.Errorf("subdir %s: %w", entry.Name, err)
			}
		} else {
			if err := t.CopyFile(srcChild, dstChild); err != nil {
				return fmt.Errorf("file %s: %w", entry.Name, err)
			}
		}
	}
	return nil
}

// joinPath 根据路径类型（本地/远端）选择合适的路径连接符
func joinPath(isRemote bool, elems ...string) string {
	if isRemote {
		return path.Join(elems...)
	}
	return filepath.Join(elems...)
}

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
