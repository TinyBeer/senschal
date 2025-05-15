package cmd

import (
	"fmt"
	"log"
	"seneschal/config"
	"strings"

	"github.com/spf13/cobra"
)

type IEnvMgr interface {
	GetName() string
	Check(c *config.SSHConfig) (any, error)
	Deploy(c *config.SSHConfig) error
}

func init() {
	// ecm, err := config.GetEnvConfigMap()
	// if err != nil {
	// 	panic(err)
	// }
	// for _, ec := range ecm {
	// 	if ec.Default {
	// 		if ec.Docker != nil {
	// 			envMgrList = append(envMgrList, NewEnvMgrDocker(ec))
	// 		}
	// 	}
	// }

	envCmd.AddCommand(envCheckCmd)
	envCmd.AddCommand(envDeployCmd)
	rootCmd.AddCommand(envCmd)
}

var envDeployCmd = &cobra.Command{
	Use:   "deploy <alias>",
	Short: "部署预制环境",
	Long:  "为指定机器部署预制环境",
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

		envAlias := args[0]
		ec, find := ecm[envAlias]
		if !find {
			log.Printf("未找到环境[%s]的配置信息\n", envAlias)
			return
		}
		var envMgrList []IEnvMgr
		envMgrList = append(envMgrList, NewEnvMgrDocker(ec))

		sshAliasList := strings.Split(args[1], ",")

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
		for _, alias := range args {
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

var envCheckCmd = &cobra.Command{
	Use:   "check <env> <alias1>[,alias2]...",
	Short: "检查哪些环境已经部署",
	Long:  "输出已经完成部署的内容",
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

		envAlias := args[0]
		ec, find := ecm[envAlias]
		if !find {
			log.Printf("未找到环境[%s]的配置信息\n", envAlias)
			return
		}
		var envMgrList []IEnvMgr
		envMgrList = append(envMgrList, NewEnvMgrDocker(ec))

		sshAliasList := strings.Split(args[1], ",")

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
			log.Printf("environment[%v] check machine[%v] start ...\n", envAlias, alias)
			c := scm[alias]
			for _, mgr := range envMgrList {
				log.Printf("environment manager[%v] checking ...\n", mgr.GetName())
				res, err := mgr.Check(c)
				if err != nil {
					log.Printf("check environment with config[%v] failed, err:%v\n", c.SSH, err)
				} else {
					if diagnosis, ok := res.(*DockerDiagnosis); ok {
						if !diagnosis.IsInstalled {
							fmt.Println("docker is not installed")
						} else {
							fmt.Printf("docker version: %s\n", diagnosis.Version)
							if len(diagnosis.MissingImageList) != 0 {
								fmt.Println("missing images: ", diagnosis.MissingImageList)
							} else {
								fmt.Println("all images is loaded")
							}
						}
					} else {
						log.Fatalf("failed to convert res[%v] to diagnosis", res)
						return
					}
				}
			}
		}

	},
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "环境配置",
	Long:  "列出当前环境配置",
	Run: func(cmd *cobra.Command, args []string) {
		ecm, err := config.GetEnvConfigMap()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("别名\t摘要\n")
		for alias, ec := range ecm {
			var abstract string
			if ec.Docker != nil && ec.Docker.Enable {
				abstract = "docker enable"
				if len(ec.Docker.ImageList) != 0 {
					abstract = fmt.Sprintf("docker with images: %v", ec.Docker.ImageList)
				}
			}
			fmt.Printf("%s\t%s\n", alias, abstract)
		}
	},
}
