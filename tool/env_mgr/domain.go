package envmgr

import "seneschal/config"

type IEnvMgr interface {
	GetName() string
	Check(c *config.SSHConfig) (any, error)
	Deploy(c *config.SSHConfig) error
}
