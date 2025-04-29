package cmd

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"seneschal/config"
	"seneschal/tool"
	"strings"
)

type EnvMgrDocker struct {
	cnf *config.Docker
}

type DockerDiagnosis struct {
	IsInstalled      bool
	Version          string
	MissingImageList []string
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
func (e *EnvMgrDocker) Check(c *config.SSHConfig) (any, error) {
	if e.cnf == nil {
		return nil, nil
	}
	se, err := tool.NewSSHExecutor(c)
	if err != nil {
		return nil, err
	}

	return checkDocker(se, e.cnf)
}

func checkDocker(e *tool.SSHExecutor, dc *config.Docker) (*DockerDiagnosis, error) {
	res := new(DockerDiagnosis)
	if dc != nil && dc.Enable {
		output, err := e.ExecuteCommand("docker version --format '{{.Client.Version}}'")
		if err != nil {
			if err.Error() == "Process exited with status 127" {
				res.IsInstalled = false
				res.MissingImageList = dc.ImageList
				return res, nil
			} else {
				return nil, fmt.Errorf("failed to check docker installment, err: %v", err)
			}
		}
		res.IsInstalled = true
		res.Version = string(bytes.Trim(output, " \n"))

		targetImageList := dc.ImageList
		if len(targetImageList) != 0 {
			output, err = e.ExecuteCommand(`docker images --format "{{.Repository}}:{{.Tag}}"`)
			if err != nil {
				return nil, fmt.Errorf("failed to check docker images, err: %v", err)
			}
			bsList := bytes.Split(bytes.Trim(output, " \n"), []byte("\n"))
			imgTbl := make(map[string]struct{})
			for _, bs := range bsList {
				image := string(bs)
				imgTbl[image] = struct{}{}
			}

			for _, image := range targetImageList {
				if _, has := imgTbl[image]; !has {
					res.MissingImageList = append(res.MissingImageList, image)
				}
			}
		}
	}
	return res, nil
}

// Deploy implements IEnvMgr.
func (e *EnvMgrDocker) Deploy(c *config.SSHConfig) error {
	if e.cnf == nil && !e.cnf.Enable {
		return nil
	}
	se, err := tool.NewSSHExecutor(c)
	if err != nil {
		return err
	}

	fmt.Println("check docker ...")
	// 1. check docker installment
	res, err := e.Check(c)
	if err != nil {
		return fmt.Errorf("failed to check environment, err: %v", err)
	}
	if diagnosis, ok := res.(*DockerDiagnosis); !ok {
		return fmt.Errorf("failed to convert res[%v] to docker diagnosis", res)
	} else {
		// 1. copy docker environment files to target machine
		// 2. deploy docker invironment
		if !diagnosis.IsInstalled {
			// 1.1 try to install docker with net
			log.Println("install docker with internet ...")
			output, err := se.ExecuteCommand("curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun")
			if err != nil {
				return fmt.Errorf("failed to install docker with internet, err: %v", err)
			}
			result := string(bytes.Trim(output, " \n"))
			if strings.Contains(result, "Server: Docker Engine - Community") {
				log.Println("install docker with internet ok...")
			} else {
				// 1.2 try to install docker with deb
				err = tool.Copy(config.DOCKER_DEB_DIR, c.Alias+":ops")
				if err != nil {
					return err
				}

				output, err = se.ExecuteCommand("bash ./ops/docker_deb/install.sh")
				if err != nil {
					return err
				}
				fmt.Println(string(output))
			}

			// user group docker
			_, err = se.ExecuteCommand("getent group docker > /dev/null 2>&1")
			if err != nil {
				if err.Error() != "Process exited with status 2" {
					return err
				}
				fmt.Println("missing docker group")
				_, err = se.ExecuteCommand("sudo groupadd docker")
				if err != nil {
					return err
				}

			}

			_, err = se.ExecuteCommand("id -Gn | grep docker")
			if err != nil {
				if err.Error() != "Process exited with status 1" {
					return err
				}
				fmt.Println("user not in docker group")
				_, err = se.ExecuteCommand("sudo usermod -aG docker $USER & newgrp docker")
				if err != nil {
					return err
				}
			}
		}

		if len(diagnosis.MissingImageList) != 0 {
			//  1.3 load images
			log.Println("docker images loading ...")
			for _, image := range diagnosis.MissingImageList {
				imageTarName := strings.ReplaceAll(strings.ReplaceAll(image, ":", "_"), "/", "+") + ".tar"
				err = tool.Copy(filepath.Join(config.DOCKER_IMAGE_DIR, imageTarName), c.Alias+":ops/docker_image/"+imageTarName)
				if err != nil {
					return err
				}
			}
			err = tool.Copy(filepath.Join(config.DOCKER_IMAGE_DIR, "load.sh"), c.Alias+":ops/docker_image/load.sh")
			if err != nil {
				return err
			}
			output, err := se.ExecuteCommand("bash ./ops/docker_image/load.sh")
			if err != nil {
				return err
			}
			fmt.Println(string(output))

		}
	}

	return nil
}

var _ IEnvMgr = new(EnvMgrDocker)
