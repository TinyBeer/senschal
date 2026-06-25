package fsutil

import (
	"io"

	"golang.org/x/crypto/ssh"
)

// sshClientConfig SSH内部私有配置
type sshClientConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	PrivateKey     []byte
	DialTimeoutSec int
	SessTimeoutSec int
}

// sshRemoteStat 远端文件元数据
type sshRemoteStat struct {
	Exist bool
	IsDir bool
	Size  int64
}

// sshClient SSH连接与文件操作封装
type sshClient struct {
	conf   *sshClientConfig
	client *ssh.Client
}

func newSSHClient(conf *sshClientConfig) (*sshClient, error)              { return nil, nil }
func (c *sshClient) close() error                                         { return nil }
func (c *sshClient) stat(remotePath string) (*sshRemoteStat, error)       { return nil, nil }
func (c *sshClient) openReader(remotePath string) (io.ReadCloser, error)  { return nil, nil }
func (c *sshClient) openWriter(remotePath string) (io.WriteCloser, error) { return nil, nil }
func (c *sshClient) listDir(remoteDir string) ([]string, error)           { return nil, nil }
