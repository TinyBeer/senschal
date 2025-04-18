package cmd

import (
	"bytes"
	"fmt"
	"seneschal/config"
	"seneschal/tool"
)

type EnvMgrDocker struct {
	cnf *config.Docker
}

// Name implements IEnvMgr.
func (e *EnvMgrDocker) GetName() string {
	return "docker"
}

func NewEnvMgrDocker(ec *config.EnvConfig) *EnvMgrDocker {
	if ec == nil || ec.Docker == nil {
		return nil
	}
	return &EnvMgrDocker{
		cnf: ec.Docker,
	}
}

// Check implements IEnvMgr.
func (e *EnvMgrDocker) Check(c *config.SSHConfig) error {
	if e.cnf == nil {
		return nil
	}
	se, err := tool.NewSSHExecutor(c)
	if err != nil {
		return err
	}

	err = checkDocker(se, e.cnf)
	if err != nil {
		return err
	}

	return nil
}

func checkDocker(e *tool.SSHExecutor, dc *config.Docker) error {
	if dc != nil && dc.Enable {
		fmt.Println("check docker ...")
		output, err := e.ExecuteCommand("docker version --format 'docker: {{.Client.Version}}'")
		if err != nil {
			if err.Error() == "Process exited with status 127" {
				fmt.Println("docker not install")
				return nil
			} else {
				fmt.Println("failed to check docker installment")
				return err
			}
		}
		fmt.Println(string(bytes.Trim(output, " \n")))

		targetImageList := dc.ImageList
		if len(targetImageList) != 0 {
			fmt.Println("check docker images ...")
			output, err = e.ExecuteCommand(`docker images --format "{{.Repository}}:{{.Tag}}"`)
			if err != nil {
				fmt.Println("failed to check docker images")
				return err
			}
			bsList := bytes.Split(bytes.Trim(output, " \n"), []byte("\n"))
			imgTbl := make(map[string]struct{})
			for _, bs := range bsList {
				image := string(bs)
				imgTbl[image] = struct{}{}
			}

			var missingImageList []string
			for _, image := range targetImageList {
				if _, has := imgTbl[image]; !has {
					missingImageList = append(missingImageList, image)
				}
			}
			if len(missingImageList) == 0 {
				fmt.Println("ok")
			} else {
				fmt.Println("missing images:", missingImageList)
			}
		}
	}
	return nil
}

// Deploy implements IEnvMgr.
func (e *EnvMgrDocker) Deploy(c *config.SSHConfig) error {
	// 1. check docker installment
	// 1.1 try to install docker with net
	// 1.2 try to install docker with deb
	// 2. check docker images
	// 2.2 load missing images
	return nil
}

var _ IEnvMgr = new(EnvMgrDocker)
