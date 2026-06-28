package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"seneschal/config"
	"seneschal/internal/fsutil"
	envmgr "seneschal/internal/runner/env_mgr"
	"seneschal/pkg/util"

	"github.com/spf13/cobra"
)

func init() {
	hostAddCmd.Flags().StringP("user", "u", "", "ssh user")
	hostAddCmd.Flags().String("host", "", "ssh host address")
	hostAddCmd.Flags().IntP("port", "p", 22, "ssh port (default 22)")
	hostAddCmd.Flags().String("password", "", "ssh password")
	hostAddCmd.Flags().String("private", "", "ssh private key file path")

	hostCmd.AddCommand(hostAddCmd)
	hostCmd.AddCommand(hostListCmd)
	hostCmd.AddCommand(hostCheckCmd)
	hostCmd.AddCommand(hostDeployCmd)
	hostCmd.AddCommand(hostUpCmd)
	hostCmd.AddCommand(hostDownCmd)
	rootCmd.AddCommand(hostCmd)
}

var hostCmd = &cobra.Command{
	Use:     "host",
	Short:   "host manager tool",
	Example: "seneschal host [list|check|deploy|up|down]",
}

var hostAddCmd = &cobra.Command{
	Use:   "add <alias>",
	Short: "add host config",
	Example: "seneschal host add myserver -u root --host 192.168.1.100 -p 22 " +
		"--password mypassword\n" +
		"  seneschal host add myserver -u root --host 192.168.1.100 --private ~/.ssh/id_rsa",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]

		user, _ := cmd.Flags().GetString("user")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		password, _ := cmd.Flags().GetString("password")
		privateKey, _ := cmd.Flags().GetString("private")

		if user == "" {
			return fmt.Errorf("user is required (--user / -u)")
		}
		if host == "" {
			return fmt.Errorf("host is required (--host)")
		}
		if password == "" && privateKey == "" {
			return fmt.Errorf("password (--password) or private key (--private) is required")
		}

		// 确定认证方式
		var method config.SSHAuthMethod
		var privateKeyPath string

		if privateKey != "" {
			method = config.SSHAuthMethod_KEY

			// 确保 SSH_KEY_DIR 存在
			if err := os.MkdirAll(config.SSH_KEY_DIR, 0o700); err != nil {
				return fmt.Errorf("failed to create ssh key dir %s: %w", config.SSH_KEY_DIR, err)
			}

			// 拷贝密钥文件到 SSH_KEY_DIR
			keyFileName := filepath.Base(privateKey)
			dstKeyPath := filepath.Join(config.SSH_KEY_DIR, keyFileName)

			srcFile, err := os.Open(privateKey)
			if err != nil {
				return fmt.Errorf("failed to open private key %s: %w", privateKey, err)
			}
			defer srcFile.Close()

			dstFile, err := os.OpenFile(dstKeyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
			if err != nil {
				return fmt.Errorf("failed to create key file %s: %w", dstKeyPath, err)
			}
			defer dstFile.Close()

			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return fmt.Errorf("failed to copy private key: %w", err)
			}

			privateKeyPath = dstKeyPath
			log.Printf("private key copied to %s\n", dstKeyPath)
		} else {
			method = config.SSHAuthMethod_PW
		}

		// 写入配置文件
		cfg := &config.SSHConfig{
			Alias: alias,
			SSH: &config.SSH{
				User:       user,
				Host:       host,
				Port:       port,
				Method:     method,
				Password:   password,
				PrivateKey: privateKeyPath,
			},
		}
		if err := config.WriteSSHConfig(cfg); err != nil {
			return err
		}

		log.Printf("SSH config for %s saved\n", alias)
		return nil
	},
}

var hostListCmd = &cobra.Command{
	Use:     "list",
	Short:   "list hosts",
	Example: "seneschal host list",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}
		if len(m) == 0 {
			fmt.Println("没有找到可用的配置")
			return nil
		}
		var data [][]string
		data = append(data, []string{"alias", "host", "user"})
		for k, v := range m {
			data = append(data, []string{k, v.SSH.Host, v.SSH.User})
		}
		util.ShowTable(data)
		return nil
	},
}

// host upload file or dir
var hostUpCmd = &cobra.Command{
	Use:   "up <alias1>[,alias2]... <local_path> <remote_path>",
	Short: "upload file or dir to host",
	Long: `Upload file or directory to remote host(s).

Directory behavior:
  seneschal host up s1 ./deploy /opt/app    → creates /opt/app/deploy/ on remote
  seneschal host up s1 ./deploy/* /opt/app  → copies contents directly into /opt/app/`,
	Example: "seneschal host up host1,host2 test.txt ops",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliases := strings.Split(args[0], ",")
		rawLocalPath := args[1]
		rawRemotePath := args[2]

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}

		// 解析 /* 标记：dir/* 表示只拷贝目录内容，不保留外层目录名
		contentsOnly := strings.HasSuffix(rawLocalPath, "/*")
		localPath := strings.TrimSuffix(rawLocalPath, "/*")

		localRef, err := fsutil.Parse(localPath)
		if err != nil {
			return fmt.Errorf("invalid local path: %w", err)
		}
		localFs, err := fsutil.GetFS(localRef, scm)
		if err != nil {
			return fmt.Errorf("failed to get local fs: %w", err)
		}
		localStat, err := localFs.Stat(localRef)
		if err != nil {
			return fmt.Errorf("stat local path: %w", err)
		}
		if !localStat.Exist {
			return fmt.Errorf("local path not found: %s", localPath)
		}
		if contentsOnly && !localStat.IsDir {
			return fmt.Errorf("contents-only pattern /* requires a directory: %s", rawLocalPath)
		}

		transfer := fsutil.NewTransfer(scm)

		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}
			if _, ok := scm[alias]; !ok {
				log.Printf("跳过 %s：未找到配置信息\n", alias)
				continue
			}

			if contentsOnly {
				// dir/* 模式：逐条拷贝到远端，不保留外层目录名
				entries, err := localFs.ListDir(localRef)
				if err != nil {
					log.Printf("读取目录 %s 失败: %v", localPath, err)
					continue
				}
				for _, entry := range entries {
					srcEntry := filepath.Join(localPath, entry.Name)
					dstEntry := alias + ":" + path.Join(rawRemotePath, entry.Name)
					srcRef, _ := fsutil.Parse(srcEntry)
					dstRef, _ := fsutil.Parse(dstEntry)
					if entry.Stat.IsDir {
						if err := transfer.CopyDir(srcRef, dstRef); err != nil {
							log.Printf("上传子目录 %s 到 %s 失败: %v", entry.Name, alias, err)
						}
					} else {
						if err := transfer.CopyFile(srcRef, dstRef); err != nil {
							log.Printf("上传文件 %s 到 %s 失败: %v", entry.Name, alias, err)
						}
					}
				}
			} else if localStat.IsDir {
				// 目录拷贝：在远端创建同名目录
				dirName := filepath.Base(localRef.RawPath)
				dstPath := alias + ":" + path.Join(rawRemotePath, dirName)
				dstRef, _ := fsutil.Parse(dstPath)
				if err := transfer.CopyDir(localRef, dstRef); err != nil {
					log.Printf("上传目录到 %s 失败: %v", alias, err)
					continue
				}
			} else {
				// 单文件上传
				fullRemotePath := alias + ":" + rawRemotePath
				if err := transfer.Upload(localPath, fullRemotePath); err != nil {
					log.Printf("上传文件到 %s 失败: %v", alias, err)
					continue
				}
			}
			log.Printf("上传 %s → %s:%s 成功\n", rawLocalPath, alias, rawRemotePath)
		}
		return nil
	},
}

// host download file or dir
var hostDownCmd = &cobra.Command{
	Use:   "down <alias> <remote_path> <local_path>",
	Short: "download file or dir from host",
	Long: `Download file or directory from remote host.

Directory behavior:
  seneschal host down s1 /opt/data ./backup    → creates ./backup/data/ locally
  seneschal host down s1 /opt/data/* ./backup  → copies contents directly into ./backup/`,
	Example: "seneschal host down host1 test.txt .",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		rawRemotePath := args[1]
		localPath := args[2]

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}
		if _, ok := scm[alias]; !ok {
			return fmt.Errorf("未找到 %s 的配置信息", alias)
		}

		// 解析 /* 标记
		contentsOnly := strings.HasSuffix(rawRemotePath, "/*")
		remotePath := strings.TrimSuffix(rawRemotePath, "/*")

		fullRemotePath := alias + ":" + remotePath
		srcRef, err := fsutil.Parse(fullRemotePath)
		if err != nil {
			return fmt.Errorf("invalid remote path: %w", err)
		}

		remoteFs, err := fsutil.GetFS(srcRef, scm)
		if err != nil {
			return fmt.Errorf("failed to get remote fs: %w", err)
		}
		remoteStat, err := remoteFs.Stat(srcRef)
		if err != nil {
			return fmt.Errorf("stat remote path: %w", err)
		}
		if !remoteStat.Exist {
			return fmt.Errorf("remote path not found: %s:%s", alias, remotePath)
		}
		if contentsOnly && !remoteStat.IsDir {
			return fmt.Errorf("contents-only pattern /* requires a directory: %s", rawRemotePath)
		}

		transfer := fsutil.NewTransfer(scm)

		if contentsOnly {
			// /* 模式：逐条拷贝到本地，不保留外层目录名
			entries, err := remoteFs.ListDir(srcRef)
			if err != nil {
				return fmt.Errorf("list remote dir %s failed: %w", remotePath, err)
			}
			for _, entry := range entries {
				srcEntry := alias + ":" + path.Join(remotePath, entry.Name)
				dstEntry := filepath.Join(localPath, entry.Name)
				srcRef2, _ := fsutil.Parse(srcEntry)
				dstRef2, _ := fsutil.Parse(dstEntry)
				if entry.Stat.IsDir {
					if err := transfer.CopyDir(srcRef2, dstRef2); err != nil {
						log.Printf("下载子目录 %s 失败: %v", entry.Name, err)
					}
				} else {
					if err := transfer.CopyFile(srcRef2, dstRef2); err != nil {
						log.Printf("下载文件 %s 失败: %v", entry.Name, err)
					}
				}
			}
		} else if remoteStat.IsDir {
			// 目录拷贝：在本地创建同名目录
			dirName := filepath.Base(remotePath)
			dstRef, _ := fsutil.Parse(filepath.Join(localPath, dirName))
			if err := transfer.CopyDir(srcRef, dstRef); err != nil {
				return fmt.Errorf("下载目录失败: %w", err)
			}
		} else {
			if err := transfer.Download(fullRemotePath, localPath); err != nil {
				return fmt.Errorf("下载文件失败: %w", err)
			}
		}
		log.Printf("下载 %s:%s → %s 成功\n", alias, rawRemotePath, localPath)
		return nil
	},
}

// host check env
var hostCheckCmd = &cobra.Command{
	Use:     "check <alias1>[,alias2]... <env>",
	Short:   "check host environment",
	Example: "seneschal host check <alias1,alias2> <env>",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scm, sshAliasList, envMgrList, err := prepareHostEnv(args)
		if err != nil {
			return err
		}

		var data [][]string
		tblHead := []string{"alias"}

		t := reflect.TypeOf(envmgr.DockerDiagnosis{})
		for i := range t.NumField() {
			tblHead = append(tblHead, t.Field(i).Name)
		}
		data = append(data, tblHead)
		for _, alias := range sshAliasList {
			tblRow := []string{alias}
			c := scm[alias]
			for _, mgr := range envMgrList {
				res, err := mgr.Check(c)
				if err != nil {
					tblRow = append(tblRow, fmt.Sprintf("check environment with config[%v] failed, err:%v\n", c.SSH, err))
				} else {
					if diagnosis, ok := res.(*envmgr.DockerDiagnosis); ok {
						v := reflect.ValueOf(*diagnosis)
						for i := range v.NumField() {
							tblRow = append(tblRow, fmt.Sprintf("%v", v.Field(i)))
						}
					} else {
						return fmt.Errorf("failed to convert res[%v] to diagnosis", res)
					}
				}
			}
			data = append(data, tblRow)
		}
		util.ShowTable(data)
		return nil
	},
}

// host deploy
var hostDeployCmd = &cobra.Command{
	Use:     "deploy <alias1>[,alias2]... <env>",
	Short:   "deploy env on selected host",
	Example: "seneschal host deploy <alias1,alias2> <env>",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scm, sshAliasList, envMgrList, err := prepareHostEnv(args)
		if err != nil {
			return err
		}

		for _, alias := range sshAliasList {
			c := scm[alias]
			for _, mgr := range envMgrList {
				log.Printf("environment manager[%v] deploying ...\n", mgr.GetName())
				err := mgr.Deploy(c)
				if err != nil {
					log.Printf("deploy environment with config[%v] failed, err:%v\n", c, err)
				}
			}
		}
		return nil
	},
}

// prepareHostEnv loads env/ssh config and validates aliases.
// Shared by check and deploy commands.
func prepareHostEnv(args []string) (scm map[string]*config.SSHConfig, aliases []string, envMgrs []envmgr.IEnvMgr, err error) {
	ecm, err := config.GetEnvConfigMap()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get env config: %w", err)
	}

	scm, err = config.GetSSHConfigMap()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get ssh config: %w", err)
	}

	envAlias := args[1]
	ec, find := ecm[envAlias]
	if !find {
		return nil, nil, nil, fmt.Errorf("未找到环境[%s]的配置信息", envAlias)
	}

	envMgrs = append(envMgrs, envmgr.NewEnvMgrDocker(ec))

	aliases = strings.Split(args[0], ",")

	for _, alias := range aliases {
		if _, ok := scm[alias]; !ok {
			log.Printf("未找到%s的配置信息\n", alias)
			err = fmt.Errorf("存在未找到的机器配置")
		}
	}
	return scm, aliases, envMgrs, err
}
