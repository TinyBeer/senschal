package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"seneschal/config"
	"seneschal/tool"
	"seneschal/tool/file"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
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
