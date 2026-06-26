package fsutil

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"seneschal/config"

	"golang.org/x/crypto/ssh"
)

// ===================== 远端文件元数据 =====================

var _ RemoteClient = (*sshClient)(nil)

// RemoteStat 远端文件元数据
type RemoteStat struct {
	Exist bool
	IsDir bool
	Size  int64
}

// RemoteDirEntry 远端目录条目
type RemoteDirEntry struct {
	Name string
	Stat FileStat
}

// sshClient SSH连接与文件操作封装
type sshClient struct {
	conf   *config.SSHConfig
	client *ssh.Client
}

// newSSHClient 建立SSH连接，支持 password / key 两种认证方式
func newSSHClient(conf *config.SSHConfig) (*sshClient, error) {
	sshConf := conf.SSH
	if sshConf == nil {
		return nil, fmt.Errorf("alias[%s]: missing ssh config", conf.Alias)
	}

	cc := &ssh.ClientConfig{
		User:            sshConf.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	switch sshConf.Method {
	case config.SSHAuthMethod_PW:
		cc.Auth = append(cc.Auth, ssh.Password(sshConf.Password))
	case config.SSHAuthMethod_KEY:
		keyPath := filepath.Join(config.SSH_KEY_DIR, sshConf.PrivateKey)
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key %s: %w", keyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		cc.Auth = append(cc.Auth, ssh.PublicKeys(signer))
	default:
		return nil, fmt.Errorf("unsupported auth method: %s", sshConf.Method)
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshConf.Host, sshConf.Port), cc)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s:%d: %w", sshConf.Host, sshConf.Port, err)
	}

	return &sshClient{conf: conf, client: client}, nil
}

// Close 关闭SSH连接
func (c *sshClient) Close() error {
	return c.client.Close()
}

// runCommand 在远端执行单条命令，返回 stdout+stderr
func (c *sshClient) runCommand(cmd string) (string, error) {
	sess, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("run %q: %w\n%s", cmd, err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// Stat 查询远端文件/目录元数据
func (c *sshClient) Stat(remotePath string) (*RemoteStat, error) {
	// 判断是否存在
	exist, err := c.runCommand(fmt.Sprintf("test -e %s && echo 1 || echo 0", remotePath))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(exist) != "1" {
		return &RemoteStat{Exist: false, IsDir: false, Size: 0}, nil
	}

	// 判断是否目录
	isDirOut, err := c.runCommand(fmt.Sprintf("test -d %s && echo 1 || echo 0", remotePath))
	if err != nil {
		return nil, err
	}
	isDir := strings.TrimSpace(isDirOut) == "1"

	// 获取大小（目录大小统一返回 0）
	var size int64
	if !isDir {
		sizeOut, err := c.runCommand(fmt.Sprintf("stat -c %%s %s", remotePath))
		if err != nil {
			return nil, err
		}
		size, _ = strconv.ParseInt(strings.TrimSpace(sizeOut), 10, 64)
	}

	return &RemoteStat{
		Exist: true,
		IsDir: isDir,
		Size:  size,
	}, nil
}

// OpenReader 打开远端文件读流（cat <file> 通过 stdout 返回）
func (c *sshClient) OpenReader(remotePath string) (io.ReadCloser, error) {
	sess, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	cmd := fmt.Sprintf("cat %s", remotePath)
	if err := sess.Start(cmd); err != nil {
		sess.Close()
		return nil, fmt.Errorf("start %q: %w", cmd, err)
	}

	return &remoteReadCloser{session: sess, reader: stdout}, nil
}

// OpenWriter 打开远端文件写流，参考 runner/file.go GetWriter 模式：
//   - 获取 StdinPipe，在 goroutine 中运行 session.Run("cat > file")
//   - 返回的 WriteCloser 即为 StdinPipe，关闭时发送 EOF 给远端 cat
//   - goroutine 内的 Run 在 cat 退出后结束并清理 session
func (c *sshClient) OpenWriter(remotePath string) (io.WriteCloser, error) {
	sess, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}

	writer, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	cmd := fmt.Sprintf("cat > %s", remotePath)
	go func() {
		// Run 阻塞直到 cat 退出（writer 关闭时收到 EOF）
		_ = sess.Run(cmd)
	}()

	// 直接返回 StdinPipe 作为 WriteCloser，由 goroutine 管理 session 生命周期
	return writer, nil
}

// ListDir 列出远端目录内容，通过 find + printf 获取名称/大小/类型
func (c *sshClient) ListDir(remoteDir string) ([]RemoteDirEntry, error) {
	// 使用 find 遍历一级目录，输出格式：name\t size\t type
	// type：f=file d=directory l=symlink（统一视作文件）
	cmd := fmt.Sprintf(`find %s -mindepth 1 -maxdepth 1 -printf '%%f\t%%s\t%%y\n'`, remoteDir)
	out, err := c.runCommand(cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// 空目录时 find 输出为空
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []RemoteDirEntry{}, nil
	}

	result := make([]RemoteDirEntry, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		name := parts[0]
		// 避免 find 返回完整路径时截取 basename（-printf %f 已经只返回名称）
		if strings.Contains(name, "/") {
			name = path.Base(name)
		}

		size, _ := strconv.ParseInt(parts[1], 10, 64)
		fileType := parts[2]

		result = append(result, RemoteDirEntry{
			Name: name,
			Stat: FileStat{
				Exist: true,
				IsDir: fileType == "d",
				Size:  size,
			},
		})
	}

	return result, nil
}

// ===================== 流封装 =====================

// remoteReadCloser 封装 SSH session 的 stdout pipe
type remoteReadCloser struct {
	session *ssh.Session
	reader  io.Reader
}

func (r *remoteReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *remoteReadCloser) Close() error {
	return r.session.Close()
}

