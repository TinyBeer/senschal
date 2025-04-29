package tool

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"seneschal/config"

	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	cfg     *config.SSHConfig
	client  *ssh.Client
	session *ssh.Session
}

func NewSSHExecutor(cfg *config.SSHConfig) (*SSHExecutor, error) {
	e := new(SSHExecutor)
	e.cfg = cfg
	return e, nil
}

func (e *SSHExecutor) getClient() (*ssh.Client, error) {
	if e.client != nil {
		return e.client, nil
	}
	user := e.cfg.SSH.User
	password := e.cfg.SSH.Password
	host := e.cfg.SSH.Host
	port := e.cfg.SSH.Port
	cC := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	switch e.cfg.SSH.Method {
	case config.SSHAuthMethod_PW:
		cC.Auth = append(cC.Auth, ssh.Password(password))
	case config.SSHAuthMethod_KEY:
		// 读取私钥文件
		privateKey, err := os.ReadFile(filepath.Join(config.SSH_KEY_DIR, e.cfg.SSH.PrivateKey))
		if err != nil {
			log.Fatalf("Failed to read private key: %v", err)
		}

		// 解析私钥
		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			log.Fatalf("Failed to parse private key: %v", err)
		}
		cC.Auth = append(cC.Auth, ssh.PublicKeys(signer))
	default:
		return nil, fmt.Errorf("unsupported methd: %v", e.cfg.SSH.Method)
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), cC)
	if err != nil {
		return nil, err
	}
	e.client = client
	return client, nil
}

func (e *SSHExecutor) NewSession() (*ssh.Session, error) {
	return e.newSession()
}

func (e *SSHExecutor) newSession() (*ssh.Session, error) {
	client, err := e.getClient()
	if err != nil {
		return nil, err
	}
	return client.NewSession()
}

func (e *SSHExecutor) Stop() error {
	if e.session != nil {
		err := e.session.Close()
		if err != nil {
			return err
		}
	}
	if e.client != nil {
		err := e.client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *SSHExecutor) ExecuteCommand(command string) ([]byte, error) {
	session, err := e.newSession()
	if err != nil {
		return nil, err
	}
	return session.CombinedOutput(command)
}

// // 本地文件路径
// localFilePath := "path/to/local/file"
// // 远程文件路径
// remoteFilePath := "path/to/remote/file"

// // 传输文件
// if err := transferFileWithResume(client, localFilePath, remoteFilePath); err != nil {
// 	fmt.Printf("传输文件时出错: %s\n", err)
// 	return
// }

// 断点续传文件
func TransferFileWithResume(client *ssh.Client, localFilePath, remoteFilePath string) error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("打开本地文件 %s 时出错: %s", localFilePath, err)
	}
	defer localFile.Close()

	localFileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("获取本地文件 %s 信息时出错: %s", localFilePath, err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH 会话时出错: %s", err)
	}
	defer session.Close()

	// 检查远程文件大小
	statCommand := fmt.Sprintf("stat -c %%s %s 2>/dev/null", remoteFilePath)
	output, err := session.CombinedOutput(statCommand)
	var remoteFileSize int64
	if err == nil {
		fmt.Sscanf(string(output), "%d", &remoteFileSize)
	}

	// 如果远程文件已存在且大小小于本地文件，则进行断点续传
	if remoteFileSize > 0 && remoteFileSize < localFileInfo.Size() {
		fmt.Printf("继续传输文件，已传输 %d 字节\n", remoteFileSize)
		if _, err := localFile.Seek(remoteFileSize, io.SeekStart); err != nil {
			return fmt.Errorf("定位本地文件 %s 时出错: %s", localFilePath, err)
		}
		appendCommand := fmt.Sprintf("dd of=%s oflag=append conv=notrunc", remoteFilePath)
		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("获取 SSH 会话标准输入管道时出错: %s", err)
		}
		if err := session.Start(appendCommand); err != nil {
			return fmt.Errorf("启动 SSH 会话执行追加命令时出错: %s", err)
		}
		if _, err := io.Copy(stdin, localFile); err != nil {
			return fmt.Errorf("复制文件到远程时出错: %s", err)
		}
		if err := stdin.Close(); err != nil {
			return fmt.Errorf("关闭 SSH 会话标准输入管道时出错: %s", err)
		}
		if err := session.Wait(); err != nil {
			return fmt.Errorf("等待 SSH 会话完成追加命令时出错: %s", err)
		}
	} else if remoteFileSize == 0 {
		// 远程文件不存在，从头开始传输
		fmt.Println("开始全新传输文件")
		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("获取 SSH 会话标准输入管道时出错: %s", err)
		}
		if err := session.Start(fmt.Sprintf("cat > %s", remoteFilePath)); err != nil {
			return fmt.Errorf("启动 SSH 会话执行复制命令时出错: %s", err)
		}
		if _, err := io.Copy(stdin, localFile); err != nil {
			return fmt.Errorf("复制文件到远程时出错: %s", err)
		}
		if err := stdin.Close(); err != nil {
			return fmt.Errorf("关闭 SSH 会话标准输入管道时出错: %s", err)
		}
		if err := session.Wait(); err != nil {
			return fmt.Errorf("等待 SSH 会话完成复制命令时出错: %s", err)
		}
	} else {
		fmt.Println("远程文件已完整，无需传输")
	}

	return nil
}
