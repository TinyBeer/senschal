package fsutil

import "io"

// localStat 本地文件元数据内部结构
type localStat struct {
	Exist bool
	IsDir bool
	Size  int64
}

// localClient 本地文件操作底层封装
type localClient struct{}

func newLocalClient() *localClient { return &localClient{} }

func (c *localClient) stat(path string) (*localStat, error)           { return nil, nil }
func (c *localClient) openReader(path string) (io.ReadCloser, error)  { return nil, nil }
func (c *localClient) openWriter(path string) (io.WriteCloser, error) { return nil, nil }
func (c *localClient) listDir(dir string) ([]string, error)           { return nil, nil }
