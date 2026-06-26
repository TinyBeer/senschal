package fsutil

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSSHClient 模拟 sshClient，脱离真实SSH连接
type MockSSHClient struct {
	mock.Mock
}

func (m *MockSSHClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSSHClient) Stat(remotePath string) (*RemoteStat, error) {
	args := m.Called(remotePath)
	stat, ok := args.Get(0).(*RemoteStat)
	if !ok {
		return nil, args.Error(1)
	}
	return stat, args.Error(1)
}

func (m *MockSSHClient) OpenReader(remotePath string) (io.ReadCloser, error) {
	args := m.Called(remotePath)
	rc, ok := args.Get(0).(io.ReadCloser)
	if !ok {
		return nil, args.Error(1)
	}
	return rc, args.Error(1)
}

func (m *MockSSHClient) OpenWriter(remotePath string) (io.WriteCloser, error) {
	args := m.Called(remotePath)
	wc, ok := args.Get(0).(io.WriteCloser)
	if !ok {
		return nil, args.Error(1)
	}
	return wc, args.Error(1)
}

func (m *MockSSHClient) ListDir(remoteDir string) ([]RemoteDirEntry, error) {
	args := m.Called(remoteDir)
	names, ok := args.Get(0).([]RemoteDirEntry)
	if !ok {
		return nil, args.Error(1)
	}
	return names, args.Error(1)
}

func (m *MockSSHClient) MkdirAll(remotePath string) error {
	args := m.Called(remotePath)
	return args.Error(0)
}

// 辅助：空实现 ReadCloser / WriteCloser
type nopReadCloser struct {
	io.Reader
}

func (nopReadCloser) Close() error { return nil }

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func TestRemoteFS_Stat(t *testing.T) {
	path := "/data/file.txt"
	ref := &PathRef{RawPath: path}

	t.Run("文件存在普通文件", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		mockStat := &RemoteStat{
			Exist: true,
			IsDir: false,
			Size:  1024,
		}
		mockCli.On("Stat", path).Return(mockStat, nil)

		outStat, err := fs.Stat(ref)
		require.NoError(t, err)
		require.True(t, outStat.Exist)
		require.False(t, outStat.IsDir)
		require.EqualValues(t, 1024, outStat.Size)

		mockCli.AssertExpectations(t)
	})

	t.Run("路径不存在", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		mockStat := &RemoteStat{
			Exist: false,
			IsDir: false,
			Size:  0,
		}
		mockCli.On("Stat", path).Return(mockStat, nil)

		outStat, err := fs.Stat(ref)
		t.Log(outStat)
		require.NoError(t, err)
		require.False(t, outStat.Exist)
		require.EqualValues(t, 0, outStat.Size)

		mockCli.AssertExpectations(t)
	})

	t.Run("ssh stat 返回错误", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		testErr := errors.New("ssh connect failed")
		mockCli.On("Stat", path).Return((*RemoteStat)(nil), testErr)

		_, err := fs.Stat(ref)
		require.ErrorIs(t, err, testErr)
		mockCli.AssertExpectations(t)
	})
}

func TestRemoteFS_OpenRead(t *testing.T) {
	path := "/read.txt"
	ref := &PathRef{RawPath: path}

	t.Run("正常获取读流", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		rc := nopReadCloser{}
		mockCli.On("OpenReader", path).Return(rc, nil)

		r, err := fs.OpenRead(ref)
		require.NoError(t, err)
		require.NotNil(t, r)
		mockCli.AssertExpectations(t)
	})

	t.Run("打开失败返回错误", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		testErr := errors.New("no such file")
		mockCli.On("OpenReader", path).Return(nil, testErr)

		_, err := fs.OpenRead(ref)
		require.ErrorIs(t, err, testErr)
		mockCli.AssertExpectations(t)
	})
}

func TestRemoteFS_OpenWrite(t *testing.T) {
	path := "/write.txt"
	ref := &PathRef{RawPath: path}

	t.Run("正常获取写流", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		wc := nopWriteCloser{}
		mockCli.On("OpenWriter", path).Return(wc, nil)

		w, err := fs.OpenWrite(ref)
		require.NoError(t, err)
		require.NotNil(t, w)
		mockCli.AssertExpectations(t)
	})

	t.Run("创建文件失败", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		testErr := errors.New("permission denied")
		mockCli.On("OpenWriter", path).Return(nil, testErr)

		_, err := fs.OpenWrite(ref)
		require.ErrorIs(t, err, testErr)
		mockCli.AssertExpectations(t)
	})
}

func TestRemoteFS_ListDir(t *testing.T) {
	dir := "/data/"
	ref := &PathRef{RawPath: dir}

	t.Run("目录包含文件与文件夹", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		// sshClient.listDir 只返回名称列表
		content := []byte("a.txt")
		result := []RemoteDirEntry{{
			Name: "a.txt",
			Stat: FileStat{
				Exist: true,
				IsDir: false,
				Size:  int64(len(content)),
			},
		}, {
			Name: "subdir",
			Stat: FileStat{
				Exist: true,
				IsDir: true,
				Size:  0,
			},
		}}
		mockCli.On("ListDir", dir).Return(result, nil)

		entries, err := fs.ListDir(ref)
		require.NoError(t, err)
		require.Len(t, entries, 2)

		nameSet := map[string]bool{}
		for _, e := range entries {
			nameSet[e.Name] = true
			require.True(t, e.Stat.Exist)
		}
		require.True(t, nameSet["a.txt"])
		require.True(t, nameSet["subdir"])

		mockCli.AssertExpectations(t)
	})

	t.Run("空目录", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		mockCli.On("ListDir", dir).Return([]string{}, nil)
		entries, err := fs.ListDir(ref)
		require.NoError(t, err)
		require.Len(t, entries, 0)
		mockCli.AssertExpectations(t)
	})

	t.Run("ListDir 调用异常", func(t *testing.T) {
		mockCli := new(MockSSHClient)
		fs := &remoteFS{cli: mockCli}
		testErr := errors.New("read dir failed")
		mockCli.On("ListDir", dir).Return(nil, testErr)

		_, err := fs.ListDir(ref)
		require.ErrorIs(t, err, testErr)
		mockCli.AssertExpectations(t)
	})
}
