package cmd

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"seneschal/config"
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
