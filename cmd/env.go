package cmd

import (
	"fmt"
	"log"
	"seneschal/config"

	"github.com/spf13/cobra"
)

type IEnvMgr interface {
	GetName() string
	Check(c *config.SSHConfig) (any, error)
	Deploy(c *config.SSHConfig) error
}

var envMgrList []IEnvMgr

func init() {
	ec, err := config.GetEnvConfig()
	if err != nil {
		panic(err)
	}
	if ec.Docker != nil {
		envMgrList = append(envMgrList, NewEnvMgrDocker(ec))
	}

	envCmd.AddCommand(envCheckCmd)
	envCmd.AddCommand(envDeployCmd)
	rootCmd.AddCommand(envCmd)
}

var envDeployCmd = &cobra.Command{
	Use:   "deploy <alias>",
	Short: "部署预制环境",
	Long:  "为指定机器部署预制环境",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			log.Println(err)
			return
		}
		if len(args) == 0 {
			log.Println("请输入需要检查的机器别名")
			return
		}
		missingCfg := false
		for _, alias := range args {
			if _, ok := m[alias]; !ok {
				log.Printf("未找到%s的配置信息\n", alias)
				missingCfg = true
			}
		}
		if missingCfg {
			return
		}
		for _, alias := range args {
			c := m[alias]
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
	Use:   "check <alias>",
	Short: "检查哪些环境已经部署",
	Long:  "输出已经完成部署的内容",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := config.GetSSHConfigMap()
		if err != nil {
			log.Println(err)
			return
		}
		if len(args) == 0 {
			log.Println("请输入需要检查的机器别名")
			return
		}
		missingCfg := false
		for _, alias := range args {
			if _, ok := m[alias]; !ok {
				log.Printf("未找到%s的配置信息\n", alias)
				missingCfg = true
			}
		}
		if missingCfg {
			return
		}
		for _, alias := range args {
			c := m[alias]
			for _, mgr := range envMgrList {
				log.Printf("environment manager[%v] checking ...\n", mgr.GetName())
				res, err := mgr.Check(c)
				if err != nil {
					log.Printf("check environment with config[%v] failed, err:%v\n", c, err)
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
		ec, err := config.GetEnvConfig()
		if err != nil {
			fmt.Println(err)
			return
		}
		if ec.Docker != nil {
			fmt.Print("docker")
			if len(ec.Docker.ImageList) != 0 {
				fmt.Println(" with image list:")
				for idx, image := range ec.Docker.ImageList {
					fmt.Printf("%d:\t%s\n", idx+1, image)
				}
			}
		}
	},
}
