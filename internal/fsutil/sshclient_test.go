package fsutil

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"seneschal/config"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// ===================== 测试 SSH 服务器 =====================

const testPassword = "seneschal-test-pw"

// testSSHServer 管理一个内存中启动的真实 SSH 服务器
type testSSHServer struct {
	t        *testing.T
	listener net.Listener
	hostPort int
	hostKey  ssh.Signer
	mu       sync.Mutex
	closed   bool
}

// startPasswordSSHServer 启动一个密码认证的 SSH 测试服务器
func startPasswordSSHServer(t *testing.T, workDir string) *testSSHServer {
	t.Helper()

	hostKey := generateHostKey(t)
	listener, port := listenTCP(t)

	srv := &testSSHServer{
		t:        t,
		listener: listener,
		hostPort: port,
		hostKey:  hostKey,
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if string(pass) == testPassword {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %s", conn.User())
		},
	}
	config.AddHostKey(hostKey)

	go srv.acceptLoop(config, workDir)
	return srv
}

// startKeySSHServer 启动一个公钥认证的 SSH 测试服务器
func startKeySSHServer(t *testing.T, workDir string, allowedPubKey ssh.PublicKey) *testSSHServer {
	t.Helper()

	hostKey := generateHostKey(t)
	listener, port := listenTCP(t)

	srv := &testSSHServer{
		t:        t,
		listener: listener,
		hostPort: port,
		hostKey:  hostKey,
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(key.Marshal(), allowedPubKey.Marshal()) {
				return nil, nil
			}
			return nil, fmt.Errorf("unknown public key for %s", conn.User())
		},
	}
	config.AddHostKey(hostKey)

	go srv.acceptLoop(config, workDir)
	return srv
}

func generateHostKey(t *testing.T) ssh.Signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := ssh.NewSignerFromKey(key)
	require.NoError(t, err)
	return signer
}

func listenTCP(t *testing.T) (net.Listener, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return listener, listener.Addr().(*net.TCPAddr).Port
}

func (s *testSSHServer) acceptLoop(config *ssh.ServerConfig, workDir string) {
	for {
		nConn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return
			}
			s.t.Logf("accept error: %v", err)
			return
		}
		go s.handleConn(nConn, config, workDir)
	}
}

func (s *testSSHServer) handleConn(nConn net.Conn, config *ssh.ServerConfig, workDir string) {
	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		s.t.Logf("handshake error: %v", err)
		return
	}
	go ssh.DiscardRequests(reqs)
	s.handleChannels(chans, workDir)
}

func (s *testSSHServer) handleChannels(chans <-chan ssh.NewChannel, workDir string) {
	for newCh := range chans {
		if newCh.ChannelType() != "session" {
			_ = newCh.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		ch, requests, err := newCh.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(ch, requests, workDir)
	}
}

func (s *testSSHServer) handleSession(ch ssh.Channel, requests <-chan *ssh.Request, workDir string) {
	defer ch.Close()

	var command string
	for req := range requests {
		switch req.Type {
		case "exec":
			var execMsg struct{ Command string }
			if err := ssh.Unmarshal(req.Payload, &execMsg); err != nil {
				_ = req.Reply(false, nil)
				continue
			}
			command = execMsg.Command
			_ = req.Reply(true, nil)

		default:
			_ = req.Reply(false, nil)
		}
		if command != "" {
			break
		}
	}
	if command == "" {
		return
	}

	// 执行命令并将 stdin/stdout/stderr 与 SSH channel 对接
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workDir
	cmd.Stdout = ch
	cmd.Stderr = ch.Stderr()

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		return
	}

	// 将 SSH channel 收到的客户端 stdin 转发到命令
	go func() {
		_, _ = io.Copy(stdinPipe, ch)
		_ = stdinPipe.Close()
	}()

	runErr := cmd.Wait()

	var status uint32
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			status = uint32(exitErr.ExitCode())
		} else {
			status = 1
		}
	}
	_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
}

func (s *testSSHServer) addr() string {
	return fmt.Sprintf("127.0.0.1:%d", s.hostPort)
}

func (s *testSSHServer) stop() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	_ = s.listener.Close()
}

// ===================== 密钥生成 =====================

func generateKeyPair(t *testing.T) (privateKeyPEM []byte, publicKey ssh.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	publicKey, err = ssh.NewPublicKey(&key.PublicKey)
	require.NoError(t, err)
	return privateKeyPEM, publicKey
}

// ===================== 测试用例 =====================

// mkSSHConfig 构建测试用 SSH 配置
func mkSSHConfig(host string, port int, user, password string) *config.SSHConfig {
	return &config.SSHConfig{
		Alias: "test-server",
		SSH: &config.SSH{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			Method:   config.SSHAuthMethod_PW,
		},
	}
}

func TestNewSSHClient_PasswordAuth(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)

	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, cli)
	require.NotNil(t, cli.client)

	t.Cleanup(func() { _ = cli.Close() })
}

func TestNewSSHClient_WrongPassword(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", "wrong-password")

	_, err := newSSHClient(cfg)
	require.Error(t, err)
}

func TestNewSSHClient_KeyAuth(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("key auth test requires unix-like environment")
	}

	workDir := t.TempDir()
	privPEM, pubKey := generateKeyPair(t)
	srv := startKeySSHServer(t, workDir, pubKey)
	defer srv.stop()

	// 将私钥写入临时文件，模拟 SSH_KEY_DIR/private_key 结构
	keyDir := filepath.Join(t.TempDir(), "ssh_key")
	err := os.MkdirAll(keyDir, 0o700)
	require.NoError(t, err)

	keyFile := "test_key"
	err = os.WriteFile(filepath.Join(keyDir, keyFile), privPEM, 0o600)
	require.NoError(t, err)

	// 暂存原 SSH_KEY_DIR，恢复后清理
	origKeyDir := config.SSH_KEY_DIR
	config.SSH_KEY_DIR = keyDir
	t.Cleanup(func() { config.SSH_KEY_DIR = origKeyDir })

	cfg := &config.SSHConfig{
		Alias: "test-server",
		SSH: &config.SSH{
			Host:       "127.0.0.1",
			Port:       srv.hostPort,
			User:       "test-user",
			Method:     config.SSHAuthMethod_KEY,
			PrivateKey: keyFile,
		},
	}

	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, cli)

	t.Cleanup(func() { _ = cli.Close() })
}

func TestSSHClient_Close(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)
	cli, err := newSSHClient(cfg)
	require.NoError(t, err)

	err = cli.Close()
	require.NoError(t, err)

	// 关闭后再次操作应报错
	_, err = cli.Stat("/tmp")
	require.Error(t, err)
}

func TestSSHClient_runCommand(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)
	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })

	out, err := cli.runCommand("echo hello")
	require.NoError(t, err)
	require.Equal(t, "hello\n", out)
}

func TestSSHClient_Stat(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)
	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })

	// 准备测试文件
	filePath := filepath.Join(workDir, "stat_test.txt")
	err = os.WriteFile(filePath, []byte("hello"), 0o644)
	require.NoError(t, err)

	subDir := filepath.Join(workDir, "subdir")
	err = os.Mkdir(subDir, 0o755)
	require.NoError(t, err)

	t.Run("存在的文件", func(t *testing.T) {
		stat, err := cli.Stat(filePath)
		require.NoError(t, err)
		require.True(t, stat.Exist)
		require.False(t, stat.IsDir)
		require.EqualValues(t, 5, stat.Size)
	})

	t.Run("不存在的路径", func(t *testing.T) {
		stat, err := cli.Stat(filepath.Join(workDir, "not_exist"))
		require.NoError(t, err)
		require.False(t, stat.Exist)
		require.False(t, stat.IsDir)
		require.EqualValues(t, 0, stat.Size)
	})

	t.Run("目录", func(t *testing.T) {
		stat, err := cli.Stat(subDir)
		require.NoError(t, err)
		require.True(t, stat.Exist)
		require.True(t, stat.IsDir)
		require.EqualValues(t, 0, stat.Size)
	})
}

func TestSSHClient_OpenReader(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)
	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })

	content := []byte("hello world from ssh")
	filePath := filepath.Join(workDir, "read_test.txt")
	err = os.WriteFile(filePath, content, 0o644)
	require.NoError(t, err)

	t.Run("正常读取文件", func(t *testing.T) {
		rc, err := cli.OpenReader(filePath)
		require.NoError(t, err)
		defer rc.Close()

		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.Equal(t, content, data)
	})

	t.Run("不存在的文件读取到空数据", func(t *testing.T) {
		rc, err := cli.OpenReader(filepath.Join(workDir, "no_such_file.txt"))
		require.NoError(t, err)

		data, err := io.ReadAll(rc)
		require.Empty(t, data) // cat 失败时 stdout 为空
		_ = rc.Close()
	})
}

func TestSSHClient_WriteViaCommandThenRead(t *testing.T) {
	// 通过 runCommand（同步）+ OpenReader（流式）验证读写双向
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)
	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })

	content := []byte("written via command and read via stream")
	remotePath := filepath.Join(workDir, "roundtrip.txt")

	// 通过 base64 + runCommand 写入（避免 shell 转义问题）
	encoded := base64.StdEncoding.EncodeToString(content)
	_, err = cli.runCommand(fmt.Sprintf("echo %s | base64 -d > %s", encoded, remotePath))
	require.NoError(t, err)

	// 通过 OpenReader 流式读取验证
	rc, err := cli.OpenReader(remotePath)
	require.NoError(t, err)
	defer rc.Close()

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, content, got)
}

func TestSSHClient_ListDir(t *testing.T) {
	workDir := t.TempDir()
	srv := startPasswordSSHServer(t, workDir)
	defer srv.stop()

	cfg := mkSSHConfig("127.0.0.1", srv.hostPort, "test-user", testPassword)
	cli, err := newSSHClient(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })

	// 准备目录结构
	// workDir/
	// ├── file_a.txt   (content "aaa")
	// ├── file_b.bin   (content "bbbb")
	// └── sub_dir/
	_ = os.WriteFile(filepath.Join(workDir, "file_a.txt"), []byte("aaa"), 0o644)
	_ = os.WriteFile(filepath.Join(workDir, "file_b.bin"), []byte("bbbb"), 0o644)
	_ = os.Mkdir(filepath.Join(workDir, "sub_dir"), 0o755)

	t.Run("混合文件和目录", func(t *testing.T) {
		entries, err := cli.ListDir(workDir)
		require.NoError(t, err)
		require.Len(t, entries, 3)

		entryMap := make(map[string]RemoteDirEntry)
		for _, e := range entries {
			entryMap[e.Name] = e
		}

		fa := entryMap["file_a.txt"]
		require.True(t, fa.Stat.Exist)
		require.False(t, fa.Stat.IsDir)
		require.EqualValues(t, 3, fa.Stat.Size)

		fb := entryMap["file_b.bin"]
		require.True(t, fb.Stat.Exist)
		require.False(t, fb.Stat.IsDir)
		require.EqualValues(t, 4, fb.Stat.Size)

		sd := entryMap["sub_dir"]
		require.True(t, sd.Stat.Exist)
		require.True(t, sd.Stat.IsDir)
	})

	t.Run("空目录", func(t *testing.T) {
		emptyDir := filepath.Join(workDir, "empty_dir")
		err := os.Mkdir(emptyDir, 0o755)
		require.NoError(t, err)

		entries, err := cli.ListDir(emptyDir)
		require.NoError(t, err)
		require.Len(t, entries, 0)
	})

	t.Run("不存在的目录返回错误", func(t *testing.T) {
		_, err := cli.ListDir(filepath.Join(workDir, "not_exist"))
		require.Error(t, err)
	})
}
