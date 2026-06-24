package sshclient

import "context"

type SSHClient interface {
	Connect(ctx context.Context)
	Upload(ctx context.Context, localPath, remotePath string) error
	Download(ctx context.Context, remotePath, localPath string) error
	Close() error
}
