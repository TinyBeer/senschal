package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"seneschal/config"
	"seneschal/internal/runner"
	"seneschal/pkg/util"

	"github.com/bndr/gojenkins"
	"github.com/spf13/cobra"
)

func init() {
	jenkinsAddCmd.Flags().StringP("user", "u", "", "jenkins username (default: alias)")
	jenkinsAddCmd.Flags().String("host", "", "jenkins host address")

	jenkinsCreateCmd.Flags().StringP("file", "f", "", "xml config file path for creating a single job")
	jenkinsCreateCmd.Flags().StringP("dir", "d", "", "directory containing job configs (dir/{job_name}/xml)")
	jenkinsCreateCmd.Flags().String("name", "", "job name (required when using --file)")
	jenkinsCreateCmd.Flags().Bool("overwrite", false, "overwrite job config if job already exists")

	jenkinsCmd.AddCommand(jenkinsListCmd)
	jenkinsCmd.AddCommand(jenkinsAddCmd)
	jenkinsCmd.AddCommand(jenkinsCreateCmd)
	rootCmd.AddCommand(jenkinsCmd)
}

// ensureJob 检查 job 是否存在，根据 overwrite 决定创建或更新
func ensureJob(ctx context.Context, jenkins *gojenkins.Jenkins, uniqueTbl map[string]struct{}, jobName, xmlContent string, overwrite bool) error {
	if _, exists := uniqueTbl[jobName]; exists {
		if !overwrite {
			log.Printf("job [%s] already exists, skip", jobName)
			return nil
		}
		job, err := jenkins.GetJob(ctx, jobName)
		if err != nil {
			return fmt.Errorf("failed to get existing job [%s], err: %w", jobName, err)
		}
		if err = job.UpdateConfig(ctx, xmlContent); err != nil {
			return fmt.Errorf("failed to update job [%s], err: %w", jobName, err)
		}
		log.Printf("job [%s] updated successfully", jobName)
		return nil
	}

	_, err := jenkins.CreateJob(ctx, xmlContent, jobName)
	if err != nil {
		return fmt.Errorf("failed to create job [%s], err: %w", jobName, err)
	}
	log.Printf("job [%s] created successfully", jobName)
	return nil
}

var jenkinsCmd = &cobra.Command{
	Use:     "jenkins",
	Short:   "list jenkins config",
	Example: "seneschal jenkins",
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
	Example: "seneschal jenkins list alias",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jenkins, ctx, err := runner.NewJenkinsClient(args[0])
		if err != nil {
			return err
		}

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
	},
}

// jenkinsCreateCmd 创建jenkins job，支持两种模式：
//   --file <xml> --name <jobName>  通过xml文件创建单个job
//   --dir  <path>                   通过文件夹批量创建，结构为 dir/{job_name}/*.xml
// 默认job已存在则跳过，使用 --overwrite 覆盖配置。

var jenkinsCreateCmd = &cobra.Command{
	Use:     "create <alias>",
	Short:   "create jenkins job(s) from xml config",
	Example: "  seneschal jenkins create myjenkins --name myJob --file ./job.xml\n  seneschal jenkins create myjenkins --dir ./jobs",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jenkins, ctx, err := runner.NewJenkinsClient(args[0])
		if err != nil {
			return err
		}

		innerJobs, err := jenkins.GetAllJobNames(ctx)
		if err != nil {
			return fmt.Errorf("failed to get all jenkins job names, err: %w", err)
		}

		uniqueTbl := make(map[string]struct{}, len(innerJobs))
		for _, job := range innerJobs {
			uniqueTbl[job.Name] = struct{}{}
		}

		filePath, _ := cmd.Flags().GetString("file")
		dirPath, _ := cmd.Flags().GetString("dir")

		if filePath == "" && dirPath == "" {
			return fmt.Errorf("either --file or --dir must be provided")
		}
		if filePath != "" && dirPath != "" {
			return fmt.Errorf("--file and --dir are mutually exclusive")
		}

		overwrite, _ := cmd.Flags().GetBool("overwrite")

		if filePath != "" {
			jobName, _ := cmd.Flags().GetString("name")
			if jobName == "" {
				return fmt.Errorf("--name is required when using --file")
			}

			xmlContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read xml config file, err: %w", err)
			}

			if err = ensureJob(ctx, jenkins, uniqueTbl, jobName, string(xmlContent), overwrite); err != nil {
				return err
			}
		} else {
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				return fmt.Errorf("failed to read directory [%s], err: %w", dirPath, err)
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				jobName := entry.Name()

				jobDir := filepath.Join(dirPath, jobName)
				xmlFiles, err := filepath.Glob(filepath.Join(jobDir, "*.xml"))
				if err != nil || len(xmlFiles) == 0 {
					log.Printf("no xml config found in directory [%s], skip", jobName)
					continue
				}

				xmlContent, err := os.ReadFile(xmlFiles[0])
				if err != nil {
					log.Printf("failed to read xml config for job [%s], err: %v", jobName, err)
					continue
				}

				if err = ensureJob(ctx, jenkins, uniqueTbl, jobName, string(xmlContent), overwrite); err != nil {
					log.Printf("%v, skip", err)
					continue
				}
			}
		}

		return nil
	},
}
