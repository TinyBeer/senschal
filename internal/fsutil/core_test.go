package fsutil

import (
	"io"
	"os"
	"testing"

	"seneschal/config"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   string
		want    *PathRef
		wantErr bool
	}{
		{
			name:  "本地路径",
			input: "test",
			want: &PathRef{
				IsRemote: false,
				Alias:    "",
				RawPath:  "test",
			},
			wantErr: false,
		},
		{
			name:  "远端路径",
			input: "alias:test",
			want: &PathRef{
				IsRemote: true,
				Alias:    "alias",
				RawPath:  "test",
			},
			wantErr: false,
		},
		{
			name:    "错误路径",
			input:   "alias:test:xxx",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := Parse(tt.input)
			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			require.EqualExportedValues(t, tt.want, got)
		})
	}
}

func TestPathRef_IsLocal(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		input string
		want  bool
	}{
		{
			name:  "远端文件",
			input: "test:test",
			want:  false,
		},
		{
			name:  "本地文件",
			input: "test",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.input)
			require.NoError(t, err)
			got := p.IsLocal()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPathRef_Validate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		input   string
		wantErr bool
	}{
		{
			name:    "本地文件",
			input:   "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.input)
			require.NoError(t, err)

			gotErr := p.Validate()
			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func TestLocalFS_Stat(t *testing.T) {
	fs := &localFS{}
	tmpDir := t.TempDir()

	t.Run("文件不存在", func(t *testing.T) {
		ref := &PathRef{RawPath: tmpDir + "/not_exist.txt"}
		stat, err := fs.Stat(ref)
		require.NoError(t, err) // 无错误

		require.False(t, stat.Exist)
		require.False(t, stat.IsDir)
		require.EqualValues(t, 0, stat.Size)
	})

	t.Run("普通文件存在", func(t *testing.T) {
		path := tmpDir + "/test.txt"
		err := os.WriteFile(path, []byte("12345"), 0o644)
		require.NoError(t, err)

		stat, err := fs.Stat(&PathRef{RawPath: path})
		require.NoError(t, err)

		require.True(t, stat.Exist)
		require.False(t, stat.IsDir)
		require.EqualValues(t, 5, stat.Size)
	})
}

func TestLocalFS_OpenRead(t *testing.T) {
	fs := &localFS{}
	tmpDir := t.TempDir()
	filePath := tmpDir + "/read.txt"
	content := []byte("hello read")
	_ = os.WriteFile(filePath, content, 0o644)

	t.Run("正常读取文件", func(t *testing.T) {
		rc, err := fs.OpenRead(&PathRef{RawPath: filePath})
		require.NoError(t, err)

		defer rc.Close()

		data, err := io.ReadAll(rc)
		require.NoError(t, err)

		require.Equal(t, content, data)
	})

	t.Run("读取不存在文件返回错误", func(t *testing.T) {
		_, err := fs.OpenRead(&PathRef{RawPath: tmpDir + "/no.txt"})
		require.Error(t, err)
	})
}

func TestLocalFS_OpenWrite(t *testing.T) {
	fs := &localFS{}
	tmpDir := t.TempDir()
	writePath := tmpDir + "/write.txt"
	content := []byte("test write")

	t.Run("创建并写入文件", func(t *testing.T) {
		wc, err := fs.OpenWrite(&PathRef{RawPath: writePath})
		require.NoError(t, err)

		_, _ = wc.Write(content)
		// 必须关闭才落盘
		err = wc.Close()
		require.NoError(t, err)

		// 校验文件存在+大小
		stat, err := os.Stat(writePath)
		require.NoError(t, err)
		require.EqualValues(t, stat.Size(), len(content))
	})
}

func TestLocalFS_MkdirAll(t *testing.T) {
	fs := &localFS{}
	tmpDir := t.TempDir()

	t.Run("创建新目录", func(t *testing.T) {
		newDir := tmpDir + "/new/sub/dir"
		err := fs.MkdirAll(&PathRef{RawPath: newDir})
		require.NoError(t, err)

		stat, err := os.Stat(newDir)
		require.NoError(t, err)
		require.True(t, stat.IsDir())
	})

	t.Run("创建已存在目录不报错", func(t *testing.T) {
		err := fs.MkdirAll(&PathRef{RawPath: tmpDir})
		require.NoError(t, err)
	})
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name     string
		isRemote bool
		elems    []string
		want     string
	}{
		{
			name:     "本地路径使用filepath.Join",
			isRemote: false,
			elems:    []string{"/base", "sub", "file.txt"},
			want:     "/base/sub/file.txt",
		},
		{
			name:     "远端路径使用path.Join",
			isRemote: true,
			elems:    []string{"/base", "sub", "file.txt"},
			want:     "/base/sub/file.txt",
		},
		{
			name:     "空元素",
			isRemote: false,
			elems:    []string{},
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinPath(tt.isRemote, tt.elems...)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewTransfer(t *testing.T) {
	confMap := map[string]*config.SSHConfig{
		"dev": {Alias: "dev"},
	}
	tr := NewTransfer(confMap)
	require.NotNil(t, tr)
	require.NotNil(t, tr.sshConfigMap)
	require.Equal(t, "dev", tr.sshConfigMap["dev"].Alias)
}

func TestGetFS_Local(t *testing.T) {
	ref := &PathRef{IsRemote: false, RawPath: t.TempDir()}
	fs, err := GetFS(ref, nil)
	require.NoError(t, err)
	require.NotNil(t, fs)

	_, ok := fs.(*localFS)
	require.True(t, ok, "expected localFS for local path")
}

func TestTransfer_CopyFile(t *testing.T) {
	tr := NewTransfer(nil)
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := srcDir + "/source.txt"
	content := []byte("hello copy file")
	err := os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	t.Run("本地文件复制成功", func(t *testing.T) {
		dstPath := dstDir + "/copied.txt"
		srcRef := &PathRef{RawPath: srcPath}
		dstRef := &PathRef{RawPath: dstPath}
		err := tr.CopyFile(srcRef, dstRef)
		require.NoError(t, err)

		copied, err := os.ReadFile(dstPath)
		require.NoError(t, err)
		require.Equal(t, content, copied)
	})

	t.Run("源文件不存在返回错误", func(t *testing.T) {
		srcRef := &PathRef{RawPath: srcDir + "/nonexistent.txt"}
		dstRef := &PathRef{RawPath: dstDir + "/out.txt"}
		err := tr.CopyFile(srcRef, dstRef)
		require.Error(t, err)
	})
}

func TestTransfer_CopyDir(t *testing.T) {
	tr := NewTransfer(nil)
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// 准备目录结构
	err := os.MkdirAll(srcDir+"/sub", 0o755)
	require.NoError(t, err)
	err = os.WriteFile(srcDir+"/a.txt", []byte("aaa"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(srcDir+"/sub/b.txt", []byte("bbb"), 0o644)
	require.NoError(t, err)

	srcRef := &PathRef{RawPath: srcDir}
	dstRef := &PathRef{RawPath: dstDir + "/copied"}

	err = tr.CopyDir(srcRef, dstRef)
	require.NoError(t, err)

	// 验证文件已递归复制
	contentA, err := os.ReadFile(dstDir + "/copied/a.txt")
	require.NoError(t, err)
	require.Equal(t, "aaa", string(contentA))

	contentB, err := os.ReadFile(dstDir + "/copied/sub/b.txt")
	require.NoError(t, err)
	require.Equal(t, "bbb", string(contentB))
}

func TestTransfer_Upload(t *testing.T) {
	tr := NewTransfer(nil)
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := srcDir + "/local.txt"
	content := []byte("upload content")
	err := os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	err = tr.Upload(srcPath, dstDir+"/uploaded.txt")
	require.NoError(t, err)

	uploaded, err := os.ReadFile(dstDir + "/uploaded.txt")
	require.NoError(t, err)
	require.Equal(t, content, uploaded)
}

func TestTransfer_Download(t *testing.T) {
	tr := NewTransfer(nil)
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := srcDir + "/remote.txt"
	content := []byte("download content")
	err := os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	err = tr.Download(srcPath, dstDir+"/downloaded.txt")
	require.NoError(t, err)

	downloaded, err := os.ReadFile(dstDir + "/downloaded.txt")
	require.NoError(t, err)
	require.Equal(t, content, downloaded)
}

func TestLocalFS_ListDir(t *testing.T) {
	fs := &localFS{}
	tmpDir := t.TempDir()

	// 准备目录结构
	// tmp/
	// ├─ file1.txt
	// ├─ file2.txt
	// └─ subdir/
	content1 := []byte("aaa")
	_ = os.WriteFile(tmpDir+"/file1.txt", content1, 0o644)
	content2 := []byte("bbbb")
	_ = os.WriteFile(tmpDir+"/file2.txt", content2, 0o644)
	_ = os.Mkdir(tmpDir+"/subdir", 0o755)

	t.Run("列出混合文件+目录", func(t *testing.T) {
		entries, err := fs.ListDir(&PathRef{RawPath: tmpDir})
		require.NoError(t, err)
		require.Equal(t, 3, len(entries), "预期3个条目，实际%d", len(entries))

		// 简单校验每个entry信息
		nameMap := make(map[string]DirEntry)
		for _, e := range entries {
			nameMap[e.Name] = e
		}

		// 校验文件
		f1 := nameMap["file1.txt"]
		require.False(t, f1.Stat.IsDir)
		require.EqualValues(t, len(content1), f1.Stat.Size)

		// 校验目录
		sub := nameMap["subdir"]
		require.True(t, sub.Stat.IsDir)
	})

	t.Run("空目录遍历", func(t *testing.T) {
		emptyDir := tmpDir + "/empty"
		_ = os.Mkdir(emptyDir, 0o755)
		entries, err := fs.ListDir(&PathRef{RawPath: emptyDir})
		require.NoError(t, err)
		require.Equal(t, 0, len(entries))
	})

	t.Run("遍历不存在目录报错", func(t *testing.T) {
		_, err := fs.ListDir(&PathRef{RawPath: tmpDir + "/not_dir"})
		require.Error(t, err)
	})
}
