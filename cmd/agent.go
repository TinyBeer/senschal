package cmd

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"seneschal/config"
	"seneschal/internal/fsutil"
	"seneschal/internal/runner"
	envmgr "seneschal/internal/runner/env_mgr"
	"seneschal/pkg/util"

	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentCpCmd)
	agentCmd.AddCommand(agentCheckCmd)
	agentCmd.AddCommand(agentDeployCmd)
	agentCmd.AddCommand(agentUpCmd)
	agentCmd.AddCommand(agentDownCmd)
	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "agent manager tool",
	Example: "seneschal agent [list|cp|check|deploy]",
}

var agentListCmd = &cobra.Command{
	Use:     "list",
	Short:   "list agent",
	Example: "seneschal agent list",
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

// agent copy command
var agentCpCmd = &cobra.Command{
	Use:     "cp",
	Short:   "agent copy file",
	Long:    "copy file between agent(not support fold yet)",
	Example: "seneschal agent cp <src> <dst>",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("cp %s to %s\n", args[0], args[1])
		err := runner.Copy(args[0], args[1])
		if err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}
		return nil
	},
}

// agent upload file or dir
var agentUpCmd = &cobra.Command{
	Use:     "up <alias1>[,alias2]... <local_path> <remote_path>",
	Short:   "upload file or dir to agent",
	Example: "seneschal agent up agent1,agent2 test.txt ops",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliases := strings.Split(args[0], ",")
		localPath := args[1]
		remotePath := args[2]

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}

		// 检查本地路径是否存在
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

		transfer := fsutil.NewTransfer(scm)

		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}
			// 检查 SSH 配置
			if _, ok := scm[alias]; !ok {
				log.Printf("跳过 %s：未找到配置信息\n", alias)
				continue
			}

			fullRemotePath := alias + ":" + remotePath
			dstRef, err := fsutil.Parse(fullRemotePath)
			if err != nil {
				log.Printf("invalid remote path for %s: %v", alias, err)
				continue
			}

			if localStat.IsDir {
				if err := transfer.CopyDir(localRef, dstRef); err != nil {
					log.Printf("上传目录到 %s 失败: %v", alias, err)
					continue
				}
			} else {
				if err := transfer.Upload(localPath, fullRemotePath); err != nil {
					log.Printf("上传文件到 %s 失败: %v", alias, err)
					continue
				}
			}
			log.Printf("上传 %s → %s:%s 成功\n", localPath, alias, remotePath)
		}
		return nil
	},
}

// agent download file or dir
var agentDownCmd = &cobra.Command{
	Use:     "down <alias> <remote_path> <local_path>",
	Short:   "download file or dir from agent",
	Example: "seneschal agent down agent test.txt .",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		remotePath := args[1]
		localPath := args[2]

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}
		if _, ok := scm[alias]; !ok {
			return fmt.Errorf("未找到 %s 的配置信息", alias)
		}

		fullRemotePath := alias + ":" + remotePath
		srcRef, err := fsutil.Parse(fullRemotePath)
		if err != nil {
			return fmt.Errorf("invalid remote path: %w", err)
		}

		// 检查远端路径是否存在
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

		transfer := fsutil.NewTransfer(scm)
		dstRef, _ := fsutil.Parse(localPath)

		if remoteStat.IsDir {
			if err := transfer.CopyDir(srcRef, dstRef); err != nil {
				return fmt.Errorf("下载目录失败: %w", err)
			}
		} else {
			if err := transfer.Download(fullRemotePath, localPath); err != nil {
				return fmt.Errorf("下载文件失败: %w", err)
			}
		}
		log.Printf("下载 %s:%s → %s 成功\n", alias, remotePath, localPath)
		return nil
	},
}

// agent check env
var agentCheckCmd = &cobra.Command{
	Use:     "check <alias1>[,alias2]... <env>",
	Short:   "check agent environment",
	Example: "seneschal agent check <alias1,alias2> <env>",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get env config: %w", err)
		}
		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}

		envAlias := args[1]
		ec, find := ecm[envAlias]
		if !find {
			return fmt.Errorf("未找到环境[%s]的配置信息", envAlias)
		}
		var envMgrList []envmgr.IEnvMgr
		envMgrList = append(envMgrList, envmgr.NewEnvMgrDocker(ec))

		sshAliasList := strings.Split(args[0], ",")

		missingCfg := false
		for _, alias := range sshAliasList {
			if _, ok := scm[alias]; !ok {
				log.Printf("未找到%s的配置信息\n", alias)
				missingCfg = true
			}
		}
		if missingCfg {
			return errors.New("存在未找到的机器配置")
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

// agent deploy
var agentDeployCmd = &cobra.Command{
	Use:     "deploy <alias1>[,alias2]... <env>",
	Short:   "deploy env on selected agent",
	Example: "seneschal agent deploy <alias1,alias2> <env>",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get env config: %w", err)
		}

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get ssh config: %w", err)
		}

		envAlias := args[1]
		ec, find := ecm[envAlias]
		if !find {
			return fmt.Errorf("未找到环境[%s]的配置信息", envAlias)
		}
		var envMgrList []envmgr.IEnvMgr
		envMgrList = append(envMgrList, envmgr.NewEnvMgrDocker(ec))

		sshAliasList := strings.Split(args[0], ",")

		missingCfg := false
		for _, alias := range sshAliasList {
			if _, ok := scm[alias]; !ok {
				log.Printf("未找到%s的配置信息\n", alias)
				missingCfg = true
			}
		}
		if missingCfg {
			return errors.New("存在未找到的机器配置")
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
