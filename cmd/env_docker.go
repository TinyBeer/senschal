package cmd

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"seneschal/config"
	"seneschal/tool"
	"strconv"
	"strings"
)

type EnvMgrDocker struct {
	cnf *config.Docker
}

type DockerDiagnosis struct {
	IsInstalled       bool
	Version           string
	MatchVersion      bool
	HaveDockerGroup   bool
	UserInDockerGroup bool
	MissingImageList  []config.Image
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
		version := string(bytes.Trim(output, " \n"))
		res.Version = version

		if dc.Version != "" {
			ok, err := compareVersion(dc.Version, version)
			if err != nil {
				return nil, fmt.Errorf("failed to check docker version, err: %v", err)
			}
			res.MatchVersion = ok
		}

		if dc.CheckUserGroup {
			output, err = e.ExecuteCommand("getent group docker")
			if err != nil {
				return nil, fmt.Errorf("failed to check user group docker, err: %v", err)
			}
			hasGroup := bytes.HasPrefix(output, []byte("docker"))
			res.HaveDockerGroup = hasGroup
			if hasGroup {
				output, err = e.ExecuteCommand("groups")
				if err != nil {
					return nil, fmt.Errorf("failed to check user group docker, err: %v", err)
				}
				res.UserInDockerGroup = bytes.Contains(output, []byte("docker"))
			}
		}

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
				if _, has := imgTbl[string(image)]; !has {
					res.MissingImageList = append(res.MissingImageList, image)
				}
			}
		}
	}
	return res, nil
}

func compareVersion(expect, actual string) (bool, error) {
	versionUp := false
	upLabel := "^"
	if strings.HasPrefix(expect, upLabel) {
		versionUp = true
		expect = strings.TrimLeft(expect, upLabel)
	}
	ec, err := parseVersion(expect)
	if err != nil {
		return false, err
	}
	ac, err := parseVersion(actual)
	if err != nil {
		return false, err
	}

	if !versionUp {
		return reflect.DeepEqual(ec, ac), nil
	}

	if len(ec) > len(ac) {
		return false, fmt.Errorf("invalid docker version config[%v]", expect)
	}

	for i, c := range ec {
		if c == ac[i] {
			continue
		}
		return c < ac[i], nil
	}
	return true, nil
}

func parseVersion(v string) ([]int, error) {
	vCodeList := strings.Split(v, ".")
	res := make([]int, 0, len(vCodeList))
	for _, str := range vCodeList {
		code, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		res = append(res, code)
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
				if !image.LocalFileExist() {
					bs, err := tool.ExecuteCommand(fmt.Sprintf("docker pull %s", string(image)))
					if err != nil {
						return err
					}
					fmt.Println(string(bs))
					bs, err = tool.ExecuteCommand(fmt.Sprintf("docker save -o %s %s", image.LocalFilePath(), string(image)))
					if err != nil {
						return err
					}
					fmt.Println(string(bs))
				}
				err = tool.Copy(image.LocalFilePath(), c.Alias+":ops/docker_image/"+image.Name())
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
