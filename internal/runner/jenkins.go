package runner

import (
	"context"
	"fmt"
	"log"

	"seneschal/config"

	"github.com/bndr/gojenkins"
)

// NewJenkinsClient 根据 alias 获取 Jenkins 配置并建立连接
func NewJenkinsClient(alias string) (*gojenkins.Jenkins, context.Context, error) {
	jcm, err := config.GetJenkinsConfigMap()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get jenkins config map, err: %w", err)
	}
	jc, ok := jcm[alias]
	if !ok {
		return nil, nil, fmt.Errorf("missing jenkins config of %s", alias)
	}

	jenkins := gojenkins.CreateJenkins(nil, jc.Host, jc.UserName, jc.Password)
	ctx := context.Background()
	_, err = jenkins.Init(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to jenkins failed, err: %w", err)
	}
	log.Printf("connected to jenkins successfully")
	return jenkins, ctx, nil
}
