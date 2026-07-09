package runner

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"seneschal/config"
	"seneschal/internal/ilog"

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

	var httpClient *http.Client
	if jc.InsecureSkipVerify {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}
	jenkins := gojenkins.CreateJenkins(httpClient, jc.Host, jc.UserName, jc.Password)
	ctx := ilog.NewTraceCtx(context.Background())
	_, err = jenkins.Init(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to jenkins failed, err: %w", err)
	}
	slog.InfoContext(ctx, "connected to jenkins successfully")
	return jenkins, ctx, nil
}
