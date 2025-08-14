package config

import "path/filepath"

const (
	Def_Data_Dir          = "data"
	Conf_Dir_Name         = "conf"
	Env_Conf_Dir_Name     = "env"
	Docker_Image_Dir_Name = "docker_image"
	Docker_Deb_Dir_Name   = "docker_deb"
	SSH_Conf_Dir_Name     = "ssh"
	SSH_Key_Dir_Name      = "ssh_key"
	Workout_Dir_Name      = "workout"
	Project_Dir_Name      = "project"
	Todo_Dir_Name         = "todo"
	Tpl_Dir_Name          = "tpl"
	Tpl_Gen_Dir_Name      = "_gen"
	Tpl_Template_Dir_Name = "template"
	Tpl_Setting_Name      = "setting"
)

var (
	CFG_DIR     = filepath.Join(Def_Data_Dir, Conf_Dir_Name)
	ENV_CFG_DIR = filepath.Join(CFG_DIR, Env_Conf_Dir_Name)

	DOCKER_DEB_DIR   = filepath.Join(Def_Data_Dir, Docker_Deb_Dir_Name)
	DOCKER_IMAGE_DIR = filepath.Join(Def_Data_Dir, Docker_Image_Dir_Name)
	SSH_CFG_DIR      = filepath.Join(CFG_DIR, SSH_Conf_Dir_Name)
	SSH_KEY_DIR      = filepath.Join(SSH_CFG_DIR, SSH_Key_Dir_Name)
	Workout_Dir      = filepath.Join(CFG_DIR, Workout_Dir_Name)
	Project_Dir      = filepath.Join(CFG_DIR, Project_Dir_Name)
	Todo_Dir         = filepath.Join(Def_Data_Dir, Todo_Dir_Name)
	Tpl_Dir          = filepath.Join(Def_Data_Dir, Tpl_Dir_Name)
	Tpl_Gen_Dir      = filepath.Join(Def_Data_Dir, Tpl_Gen_Dir_Name)
)
