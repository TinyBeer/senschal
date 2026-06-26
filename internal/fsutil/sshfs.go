package fsutil

import (
	"io"

	"seneschal/config"

	"golang.org/x/crypto/ssh"
)

// sshRemoteStat 远端文件元数据
type RemoteStat struct {
	Exist bool
	IsDir bool
	Size  int64
}

type RemoteDirEntry struct {
	Name string
	Stat FileStat
}

// sshClient SSH连接与文件操作封装
type sshClient struct {
	conf   *config.SSHConfig
	client *ssh.Client
}

func newSSHClient(conf *config.SSHConfig) (*sshClient, error)             { return nil, nil }
func (c *sshClient) Close() error                                         { return nil }
func (c *sshClient) Stat(remotePath string) (*RemoteStat, error)          { return nil, nil }
func (c *sshClient) OpenReader(remotePath string) (io.ReadCloser, error)  { return nil, nil }
func (c *sshClient) OpenWriter(remotePath string) (io.WriteCloser, error) { return nil, nil }
func (c *sshClient) ListDir(remoteDir string) ([]RemoteDirEntry, error)   { return nil, nil }
