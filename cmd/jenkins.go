package cmd

import (
	"context"
	"fmt"
	"log"
	"seneschal/config"
	"seneschal/pkg/util"

	"github.com/bndr/gojenkins"
	"github.com/spf13/cobra"
)

const (
	Jenkins_Base     = "http://10.242.1.3:8080"
	Jenkins_Account  = "joynova"
	Jenkins_Password = "Joynova1234"
)

func init() {
	jenkinsAddCmd.Flags().StringP("user", "u", "", "jenkins username (default: alias)")
	jenkinsAddCmd.Flags().String("host", "", "jenkins host address")

	jenkinsCmd.AddCommand(jenkinsListCmd)
	jenkinsCmd.AddCommand(jenkinsAddCmd)
	rootCmd.AddCommand(jenkinsCmd)
}

var jenkinsCmd = &cobra.Command{
	Use:     "jenkins",
	Short:   "list jenkins config",
	Example: "senechal jenkins",
	RunE: func(cmd *cobra.Command, args []string) error {
		jcm, err := config.GetJenkinsConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get jenkins config map, err: %w", err)
		}

		data := [][]string{{"Alias", "Host", "UserName"}}
		for _, jc := range jcm {
			data = append(data, []string{jc.Alias, jc.Host, jc.UserName})
		}
		util.ShowTable(data)
		return nil
	},
}

var jenkinsListCmd = &cobra.Command{
	Use:     "list <alias>",
	Short:   "simply list jenkins jobs",
	Long:    "simply list jenkins jobs with name",
	Example: "senechal jenkins list alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		jenkins := gojenkins.CreateJenkins(nil, Jenkins_Base, Jenkins_Account, Jenkins_Password)
		ctx := context.Background()
		_, err := jenkins.Init(ctx)
		if err != nil {
			return fmt.Errorf("connect to jenkins failed, err: %w", err)
		}
		log.Printf("connect to jenkins succeed!")

		innerJobs, err := jenkins.GetAllJobNames(ctx)
		if err != nil {
			return fmt.Errorf("failed to get all jenkins job names, err: %w", err)
		}

		data := [][]string{{"Name"}}
		for _, job := range innerJobs {
			data = append(data, []string{job.Name})
		}
		util.ShowTable(data)

		return nil
	},
}

var jenkinsAddCmd = &cobra.Command{
	Use:     "add <alias> <password|token>",
	Short:   "add jenkins config",
	Example: "seneschal jenkins add myjenkins mytoken -u admin --host http://jenkins.example.com:8080",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		password := args[1]
		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")

		if host == "" {
			return fmt.Errorf("host is required, use --host flag")
		}
		if user == "" {
			user = alias
		}

		cfg := &config.Jenkins{
			Alias:    alias,
			Host:     host,
			UserName: user,
			Password: password,
		}
		if err := config.WriteJenkinsConfig(cfg); err != nil {
			return fmt.Errorf("failed to write jenkins config: %w", err)
		}
		log.Printf("jenkins config [%s] saved successfully", alias)
		return nil
	}}
