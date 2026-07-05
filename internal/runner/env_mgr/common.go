package envmgr

import (
	"context"

	"seneschal/config"
)

//go:generate stringer -type=Environment -linecomment

type Environment int

const (
	Environment_Docker Environment = iota // docker
)

type IEnvMgr interface {
	GetName() Environment
	Check(c *config.SSHConfig) (any, error)
	Deploy(ctx context.Context, c *config.SSHConfig) error
}
