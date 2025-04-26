package config

const (
	Ext_TOML = "toml"
)

var (
	CFG_ROOT         = "./conf"
	ENV_CFG_DIR      = CFG_ROOT + "/env"
	ENV_DOCKER_DIR   = ENV_CFG_DIR + "/docker"
	DOCKER_IMAGE_DIR = "./tmp/docker_images"
	ENV_CFG_Name     = "env"
	SSH_CFG_DIR      = CFG_ROOT + "/ssh"
)
