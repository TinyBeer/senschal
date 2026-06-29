package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"seneschal/config"
	"seneschal/internal/command/file"
	"seneschal/pkg/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FlagGenDir        = "gen_dir"
	FlagGenDir_S      = "d"
	FlagSettingFile   = "setting_file"
	FlagSettingFile_S = "s"
	FlagDirName       = "name"
	FlagDirName_S     = "n"
)

func init() {
	joyTplExecCmd.Flags().StringP(FlagGenDir, FlagGenDir_S, "", "set file generate dir")
	joyTplExecCmd.Flags().StringP(FlagSettingFile, FlagSettingFile_S, "", "set file where setting read from")
	joyTplExecCmd.Flags().StringP(FlagDirName, FlagDirName_S, "", "set dir name where files generate to")
	joyTplCmd.AddCommand(joyTplExecCmd)
	joyCmd.AddCommand(joyTplCmd)
	joyInterCmd.Flags().Bool("lobby", false, "register lobby interface at same time")
	joyCmd.AddCommand(joyInterCmd)
	joyCmd.Flags().BoolP("list", "l", false, "list project config")
	rootCmd.AddCommand(joyCmd)
}

var joyCmd = &cobra.Command{
	Use:     "joy",
	Short:   "joynova project tool",
	Example: "seneschal joy [-l]",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := config.GetProjectConfigList(config.ProjectDir)
		if err != nil {
			return fmt.Errorf("failed to get project config: %w", err)
		}
		util.ShowTableWithSlice(list)
		return nil
	},
}

var joyInterCmd = &cobra.Command{
	Use:     "inter <project> [flags] <service:api_name>",
	Short:   "register interface",
	Args:    cobra.ExactArgs(2),
	Example: "seneschal joy inter <project> [--lobby] <service:api_name>",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(strings.Split(args[1], ":")) != 2 {
			return errors.New("api_name 格式应为 service:api_name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		split := strings.Split(args[1], ":")
		serivce := split[0]
		apiName := split[1]

		log.Printf("register %v %v interface %v", projectName, serivce, apiName)
		list, err := config.GetProjectConfigList(config.ProjectDir)
		if err != nil {
			return fmt.Errorf("failed to get project config: %w", err)
		}
		var pc *config.ProjectConfig
		for _, c := range list {
			if c.Alias == projectName {
				pc = c
				break
			}
		}
		if pc == nil {
			return fmt.Errorf("not found project[%v]", projectName)
		}
		registerLobby, err := getBoolFlag(cmd, "lobby")
		if err != nil {
			return fmt.Errorf("failed to parse --lobby flag: %w", err)
		}

		if registerLobby && pc.LobbyRegisterWithTool && pc.LobbyRegisterFile == "" {
			return errors.New("no lobby register file config")
		}

		apiReq := apiName + "Req"
		apiRes := apiName + "Res"
		var targetLobbyProtoFile string
		if registerLobby {
			fileList, err := file.ListFileWithExt(pc.GetProtoDir(), file.Ext_PROTO)
			if err != nil {
				return fmt.Errorf("failed to list proto files: %w", err)
			}
			for _, f := range fileList {
				if strings.Contains(filepath.Base(f), serivce) {
					targetLobbyProtoFile = f
				}
			}

		}

		var targetServiceProtoFile string
		fileList, err := file.ListFileWithExt(pc.GetServiceDir(), file.Ext_PROTO)
		if err != nil {
			return fmt.Errorf("failed to list service proto files: %w", err)
		}
		for _, f := range fileList {
			if strings.Contains(filepath.Base(f), serivce) {
				targetServiceProtoFile = f
			}
		}

		reqs := []string{apiReq}
		rets := []string{apiRes}

		log.Printf("lobby proto file: %s\nservice proto file: %s\nlobby register file: %s\n", targetLobbyProtoFile, targetServiceProtoFile, pc.GetLobbyRegisterFile())

		if registerLobby {
			contain, err := file.FileContain(targetLobbyProtoFile, file.ReplaceProbe_Message.String())
			if err != nil {
				return fmt.Errorf("failed to check lobby proto file: %w", err)
			}
			if !contain {
				return fmt.Errorf("file[%s] not has probe[%s]", targetLobbyProtoFile, file.ReplaceProbe_Message.String())
			}
		}

		contain, err := file.FileContain(targetServiceProtoFile, file.ReplaceProbe_RPC.String())
		if err != nil {
			return fmt.Errorf("failed to check service proto file: %w", err)
		}
		if !contain {
			return fmt.Errorf("file[%s] not has probe[%s]", targetServiceProtoFile, file.ReplaceProbe_RPC.String())
		}
		if registerLobby && pc.LobbyRegisterWithTool {
			contain, err = file.FileContain(pc.GetLobbyRegisterFile(), file.ReplaceProbe_Func.String())
			if err != nil {
				return fmt.Errorf("failed to check lobby register file: %w", err)
			}
			if !contain {
				return fmt.Errorf("file[%s] not has probe[%s]", pc.GetLobbyRegisterFile(), file.ReplaceProbe_Func.String())
			}
		}
		if registerLobby {
			err = file.InsertCodeIntoFile(targetLobbyProtoFile, file.ReplaceProbe_Message, file.GenerateMessage(apiReq), file.GenerateMessage(apiRes))
			if err != nil {
				return fmt.Errorf("failed to insert code into lobby proto file: %w", err)
			}
		}

		err = file.InsertCodeIntoFile(targetServiceProtoFile, file.ReplaceProbe_RPC, file.GenerateRPC(apiName, reqs, rets))
		if err != nil {
			return fmt.Errorf("failed to insert code into service proto file: %w", err)
		}

		var reqMessageOpt []file.ProtoMessageOpt
		var resMessageOpt []file.ProtoMessageOpt
		if registerLobby {
			if pc.ServiceMessageTemplate != "" {
				tpl, err := template.New("service_message_template").Parse(pc.ServiceMessageTemplate)
				if err != nil {
					return fmt.Errorf("failed to parse service message template: %w", err)
				}
				var buf bytes.Buffer
				err = tpl.Execute(&buf, map[string]interface{}{
					"api_name": apiName,
				})
				if err != nil {
					return fmt.Errorf("failed to execute service message template: %w", err)
				}
				err = file.InsertCodeIntoFile(targetServiceProtoFile, file.ReplaceProbe_Message, buf.String())
				if err != nil {
					return fmt.Errorf("failed to insert message into service proto file: %w", err)
				}
			} else {
				reqMessageOpt = append(reqMessageOpt, file.ProtoMessageWithField("greenly_proto_server.RpcRoleInfo", "Role"))
				reqMessageOpt = append(reqMessageOpt, file.ProtoMessageWithField("greenly_proto_lobby."+apiReq, "Msg"))
				resMessageOpt = append(resMessageOpt, file.ProtoMessageWithField("greenly_proto_error.ErrCode", "ErrCode"))
				resMessageOpt = append(resMessageOpt, file.ProtoMessageWithField("greenly_proto_lobby."+apiRes, "Msg"))
				err = file.InsertCodeIntoFile(targetServiceProtoFile, file.ReplaceProbe_Message,
					file.GenerateMessage(apiReq, reqMessageOpt...), file.GenerateMessage(apiRes, resMessageOpt...))
				if err != nil {
					return fmt.Errorf("failed to insert message into service proto file: %w", err)
				}
			}
		}

		if registerLobby && pc.LobbyRegisterWithTool {
			err = file.InsertCodeIntoFile(pc.GetLobbyRegisterFile(), file.ReplaceProbe_Func, fmt.Sprintf("\tapi_%s.%s,", serivce, apiName))
			if err != nil {
				return fmt.Errorf("failed to insert code into lobby register file: %w", err)
			}
		}

		return nil
	},
}

var joyTplCmd = &cobra.Command{
	Use:     "tpl",
	Short:   "list template",
	Example: "seneschal joy tpl",
	RunE: func(cmd *cobra.Command, args []string) error {
		infoList, err := joyTplListTemplateInfo(config.TplDir)
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}
		util.ShowTableWithSlice(infoList)
		return nil
	},
}

type JoyTplTemplateInfo struct {
	Alias string
	Desc  string
	Path  string
}

func joyTplListTemplateInfo(dir string) ([]*JoyTplTemplateInfo, error) {
	dirList, err := file.ListDirName(dir)
	if err != nil {
		return nil, err
	}
	infoList := make([]*JoyTplTemplateInfo, 0, len(dirList))
	for _, dir := range dirList {
		info, err := joyTplGetTemplateInfo(filepath.Join(config.TplDir, dir))
		if err != nil {
			return nil, err
		}
		infoList = append(infoList, info)
	}
	return infoList, nil
}

func joyTplGetTemplateInfo(dir string) (*JoyTplTemplateInfo, error) {
	settingPath := filepath.Join(dir, config.TplSettingName+"."+file.Ext_TOML)
	dirName := filepath.Base(dir)

	info := &JoyTplTemplateInfo{
		Alias: dirName,
		Desc:  "missing",
		Path:  dir,
	}

	_, err := os.Stat(settingPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Print(settingPath)
			return info, nil
		}
		return nil, err
	}
	v := viper.New()
	v.SetConfigName(filepath.Base(settingPath))
	v.SetConfigType(file.Ext_TOML)
	v.AddConfigPath(filepath.Dir(settingPath))
	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	setting := v.AllSettings()

	alias := setting["__alias__"].(string)
	description := setting["__description__"].(string)
	if alias == "" {
		alias = dirName
	}
	info.Alias = alias
	info.Desc = description
	return info, nil
}

var joyTplExecCmd = &cobra.Command{
	Use:     "exec <tpl_name> [flags]",
	Short:   "execute tpl to generate files",
	Args:    cobra.ExactArgs(1),
	Long:    "execute tpl to generate files\nNotice: setting file variable name should be lower case",
	Example: "seneschal joy tpl exec <tpl_name> [-d gen_dir] [-s setting_file] [-n name]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("请指定模板名称")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tplName := args[0]
		infoList, err := joyTplListTemplateInfo(config.TplDir)
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}
		genDir, err := getStringFlag(cmd, FlagGenDir)
		if err != nil {
			return fmt.Errorf("failed to parse --gen-dir flag: %w", err)
		}
		settingFilePath, err := getStringFlag(cmd, FlagSettingFile)
		if err != nil {
			return fmt.Errorf("failed to parse --setting-file flag: %w", err)
		}
		dirName, err := getStringFlag(cmd, FlagDirName)
		if err != nil {
			return fmt.Errorf("failed to parse --name flag: %w", err)
		}
		for _, info := range infoList {
			if info.Alias == tplName {
				tplPath := info.Path
				if genDir == "" {
					genDir = filepath.Join(config.DefaultGenerateDir, tplName)
					if dirName != "" {
						genDir = filepath.Join(config.DefaultGenerateDir, dirName)
					}
				}
				if settingFilePath == "" {
					settingFilePath = filepath.Join(tplPath, config.TplSettingName+"."+file.Ext_TOML)
				}
				err = file.ExecuteTemplate(tplPath, genDir, settingFilePath)
				if err != nil {
					return fmt.Errorf("failed to execute template: %w", err)
				}
				return nil
			}
		}
		return fmt.Errorf("template[%s] not found", tplName)
	},
}
