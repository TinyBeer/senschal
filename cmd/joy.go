package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seneschal/config"
	"seneschal/tool"
	"seneschal/tool/file"
	"strings"

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
	Use:   "joy",
	Short: "joynova project tool",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := config.GetProjectConfigList(config.Project_Dir)
		if err != nil {
			log.Fatal(err)
		}
		tool.ShowTableWithSlice(list)
	},
}

var joyInterCmd = &cobra.Command{
	Use:   "inter <project> [flags] <service:api_name>",
	Short: "register interface",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			return
		}
		projectName := args[0]
		split := strings.Split(args[1], ":")
		if len(split) != 2 {
			cmd.Usage()
			return
		}
		serivce := split[0]
		apiName := split[1]

		log.Printf("register %v %v interface %v", projectName, serivce, apiName)
		list, err := config.GetProjectConfigList(config.Project_Dir)
		if err != nil {
			panic(err)
		}
		var pc *config.ProjectConfig
		for _, c := range list {
			if c.Alias == projectName {
				pc = c
				break
			}
		}
		if pc == nil {
			fmt.Printf("not found projcet[%v]\n", projectName)
			return
		}
		registerLobby, err := cmd.Flags().GetBool("lobby")
		if err != nil {
			log.Fatal(err)
		}

		if registerLobby && pc.LobbyRegisterFile == "" {
			log.Fatal("no lobby register file config")
		}

		apiReq := apiName + "Req"
		apiRes := apiName + "Res"
		var targetLobbyProtoFile string
		if registerLobby {
			fileList, err := file.ListFileWithExt(pc.GetProtoDir(), file.Ext_PROTO)
			if err != nil {
				log.Fatal(err)
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
			log.Fatal(err)
		}
		for _, f := range fileList {
			if strings.Contains(filepath.Base(f), serivce) {
				targetServiceProtoFile = f
			}
		}

		reqs := []string{apiReq}
		rets := []string{apiRes}

		var reqMessageOpt []file.ProtoMessageOpt
		var resMessageOpt []file.ProtoMessageOpt
		if registerLobby {
			reqMessageOpt = append(reqMessageOpt, file.ProtoMessageWithField("greenly_proto_server.RpcRoleInfo", "Role"))
			reqMessageOpt = append(reqMessageOpt, file.ProtoMessageWithField("greenly_proto_lobby."+apiReq, "Msg"))
			resMessageOpt = append(resMessageOpt, file.ProtoMessageWithField("greenly_proto_error.ErrCode", "ErrCode"))
			resMessageOpt = append(resMessageOpt, file.ProtoMessageWithField("greenly_proto_lobby."+apiRes, "Msg"))
		}

		log.Printf("lobby proto file: %s\nservice proto file: %s\nlobby register file: %s\n", targetLobbyProtoFile, targetServiceProtoFile, pc.GetLobbyRegisterFile())

		if registerLobby {
			contain, err := file.FileContain(targetLobbyProtoFile, file.ReplaceProbe_Message.String())
			if err != nil {
				log.Fatal(err)
			}
			if !contain {
				log.Fatalf("file[%s] not has probe[%s]", targetLobbyProtoFile, file.ReplaceProbe_Message.String())
			}
		}

		contain, err := file.FileContain(targetServiceProtoFile, file.ReplaceProbe_RPC.String())
		if err != nil {
			log.Fatal(err)
		}
		if !contain {
			log.Fatalf("file[%s] not has probe[%s]", targetServiceProtoFile, file.ReplaceProbe_RPC.String())
		}
		if registerLobby {
			contain, err = file.FileContain(pc.GetLobbyRegisterFile(), file.ReplaceProbe_Func.String())
			if err != nil {
				log.Fatal(err)
			}
			if !contain {
				log.Fatalf("file[%s] not has probe[%s]", pc.GetLobbyRegisterFile(), file.ReplaceProbe_Func.String())
			}
		}
		if registerLobby {
			err = file.InsertCodeIntoFile(targetLobbyProtoFile, file.ReplaceProbe_Message, file.GenerateMessage(apiReq), file.GenerateMessage(apiRes))
			if err != nil {
				log.Fatal(err)
			}
		}

		err = file.InsertCodeIntoFile(targetServiceProtoFile, file.ReplaceProbe_RPC, file.GenerateRPC(apiName, reqs, rets))
		if err != nil {
			log.Fatal(err)
		}

		err = file.InsertCodeIntoFile(targetServiceProtoFile, file.ReplaceProbe_Message,
			file.GenerateMessage(apiReq, reqMessageOpt...), file.GenerateMessage(apiRes, resMessageOpt...))
		if err != nil {
			log.Fatal(err)
		}

		if registerLobby && pc.LobbyRegisterWithTool {
			err = file.InsertCodeIntoFile(pc.GetLobbyRegisterFile(), file.ReplaceProbe_Func, fmt.Sprintf("\tapi_%s.%s,", serivce, apiName))
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

var joyTplCmd = &cobra.Command{
	Use:   "tpl",
	Short: "list template",
	Run: func(cmd *cobra.Command, args []string) {
		infoList, err := joyTplListTemplateInfo(config.Tpl_Dir)
		if err != nil {
			log.Fatal(err)
		}
		tool.ShowTableWithSlice(infoList)
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
		log.Fatal(err)
	}
	infoList := make([]*JoyTplTemplateInfo, 0, len(dirList))
	for _, dir := range dirList {
		info, err := joyTplGetTemplateInfo(filepath.Join(config.Tpl_Dir, dir))
		if err != nil {
			return nil, err
		}
		infoList = append(infoList, info)
	}
	return infoList, nil
}

func joyTplGetTemplateInfo(dir string) (*JoyTplTemplateInfo, error) {
	settingPath := filepath.Join(dir, config.Tpl_Setting_Name+"."+file.Ext_TOML)
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
	Use:   "exec <tpl_name> [flags]",
	Short: "execute tpl to generate files",
	Long:  "execute tpl to generate files\nNotice: setting file variable name should be lower case",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		tplName := args[0]
		infoList, err := joyTplListTemplateInfo(config.Tpl_Dir)
		if err != nil {
			log.Fatal(err)
		}
		genDir, err := cmd.Flags().GetString(FlagGenDir)
		if err != nil {
			log.Fatal(err)
		}
		settingFilePath, err := cmd.Flags().GetString(FlagSettingFile)
		if err != nil {
			log.Fatal(err)
			return
		}
		dirName, err := cmd.Flags().GetString(FlagDirName)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, info := range infoList {
			if info.Alias == tplName {
				tplPath := info.Path
				if genDir == "" {
					genDir = filepath.Join(config.Tpl_Gen_Dir, tplName)
					if dirName != "" {
						genDir = filepath.Join(config.Tpl_Gen_Dir, dirName)
					}
				}
				if settingFilePath == "" {
					settingFilePath = filepath.Join(tplPath, config.Tpl_Setting_Name+"."+file.Ext_TOML)
				}
				err = file.ExecuteTemplate(tplPath, genDir, settingFilePath)
				if err != nil {
					log.Fatal(err)
				}
				return
			}
		}
		log.Printf("template[%s] not found", tplName)
	},
}
