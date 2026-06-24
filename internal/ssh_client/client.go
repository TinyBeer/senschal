package sshclient

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"seneschal/config"
	"strings"

	"github.com/melbahja/goph"
)

type Client struct {
	c *goph.Client
}

// Close implements [SSHClient].
func (c *Client) Close() error {
	return c.c.Close()
}

// Connect implements [SSHClient].
func (c *Client) Connect(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	var (
		out   []byte
		cmd   string
		parts []string
	)

loop:
	for scanner.Scan() {
		cmd = scanner.Text()
		parts = strings.Split(cmd, " ")
		if len(parts) < 1 {
			continue
		}
		switch parts[0] {
		case "exit":
			fmt.Println("bye!")
			break loop
		default:
			command, err := c.c.Command(parts[0], parts[1:]...)
			if err != nil {
				panic(err)
			}
			out, err = command.CombinedOutput()
			fmt.Println(string(out), err)
		}
		fmt.Print("> ")
	}

}

// Download implements [SSHClient].
func (c *Client) Download(ctx context.Context, remotePath string, localPath string) error {
	return c.c.Download(remotePath, localPath)
}

// Upload implements [SSHClient].
func (c *Client) Upload(ctx context.Context, localPath string, remotePath string) error {
	return c.c.Upload(localPath, remotePath)
}

func New(conf *config.SSH) (*Client, error) {
	var (
		auth goph.Auth
		err  error
	)

	switch conf.Method {
	case config.SSHAuthMethod_PW:
		auth = goph.Password(conf.Password)
	case config.SSHAuthMethod_KEY:
		auth, err = goph.Key(conf.PrivateKey, "")
		if err != nil {
			return nil, err
		}
	}

	callback, err := goph.DefaultKnownHosts()
	if err != nil {
		return nil, err
	}

	client, err := goph.NewConn(&goph.Config{
		Auth:     auth,
		User:     conf.User,
		Addr:     conf.Host,
		Port:     uint(conf.Port),
		Timeout:  goph.DefaultTimeout,
		Callback: callback,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		c: client,
	}, nil
}

var _ SSHClient = (*Client)(nil)
