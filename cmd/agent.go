package cmd

import (
	"fmt"
	"log"
	"reflect"
	"seneschal/config"
	"seneschal/tool"
	envmgr "seneschal/tool/env_mgr"
	"strings"

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
	Use:   "agent",
	Short: "agent manager tool",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "list agent",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(m) == 0 {
			fmt.Println("没有找到可用的配置")
			return
		}
		var data [][]string
		data = append(data, []string{"alias", "host", "user"})
		for k, v := range m {
			data = append(data, []string{k, v.SSH.Host, v.SSH.User})
		}
		tool.ShowTable(data)
	},
}

// agent copy command
var agentCpCmd = &cobra.Command{
	Use:   "cp",
	Short: "agent copy file",
	Long:  "copy file between agent(not support fold yet)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Println("请输入需要拷贝的文件以及要拷贝到的路径")
			return
		}
		fmt.Printf("cp %s to %s\n", args[0], args[1])
		err := tool.Copy(args[0], args[1])
		if err != nil {
			log.Println(err)
			return
		}
	},
}

// agent check env
var agentCheckCmd = &cobra.Command{
	Use:   "check <alias1>[,alias2]... <env>",
	Short: "check agent environment",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Println("请输入需要检查的 环境 和 机器别名")
			return
		}
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			log.Println(err)
			return
		}
		scm, err := config.GetSSHConfigMap()
		if err != nil {
			log.Println(err)
			return
		}

		envAlias := args[1]
		ec, find := ecm[envAlias]
		if !find {
			log.Printf("未找到环境[%s]的配置信息\n", envAlias)
			return
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
			return
		}

		var data [][]string
		tblHead := []string{"alias"}

		t := reflect.TypeOf(envmgr.DockerDiagnosis{})
		for i := range t.NumField() {
			tblHead = append(tblHead, t.Field(i).Name)
		}
		data = append(data, tblHead)
		for _, alias := range sshAliasList {
			// log.Printf("environment[%v] check machine[%v] start ...\n", envAlias, alias)
			tblRow := []string{alias}
			c := scm[alias]
			for _, mgr := range envMgrList {
				// log.Printf("environment manager[%v] checking ...\n", mgr.GetName())
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
						log.Fatalf("failed to convert res[%v] to diagnosis", res)
						return
					}
				}
			}
			data = append(data, tblRow)
		}
		tool.ShowTable(data)
	},
}

// agent deploy
var agentDeployCmd = &cobra.Command{
	Use:   "deploy <alias1>[,alias2]... <env>",
	Short: "deploy env on selected agent",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Println("请输入需要检查的 环境 和 机器别名")
			return
		}
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			log.Println(err)
			return
		}

		scm, err := config.GetSSHConfigMap()
		if err != nil {
			log.Println(err)
			return
		}

		envAlias := args[1]
		ec, find := ecm[envAlias]
		if !find {
			log.Printf("未找到环境[%s]的配置信息\n", envAlias)
			return
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
			return
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
	},
}
