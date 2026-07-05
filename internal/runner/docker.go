package runner

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"seneschal/config"
)

func DockerLoadImageList(ctx context.Context, imageList []config.Image) error {
	var needSaveList []config.Image
	for _, image := range imageList {
		if !image.LocalFileExist() {
			needSaveList = append(needSaveList, image)
		}
	}
	err := DockerPullImageList(ctx, needSaveList)
	if err != nil {
		return err
	}
	for _, image := range needSaveList {
		bs, err := ExecuteCommand(fmt.Sprintf("docker save -o %s %s", image.LocalFilePath(), string(image)))
		if err != nil {
			return err
		}
		slog.DebugContext(ctx, "output", "output", string(bs))
	}
	return nil
}

func DockerPullImageList(ctx context.Context, imageList []config.Image) error {
	output, err := ExecuteCommand(`docker images --format "{{.Repository}}:{{.Tag}}"`)
	if err != nil {
		return fmt.Errorf("failed to list docker images, err: %v", err)
	}
	bsList := bytes.Split(bytes.Trim(output, " \n"), []byte("\n"))
	imgTbl := make(map[string]bool)
	for _, bs := range bsList {
		image := string(bs)
		imgTbl[image] = true
	}

	for _, image := range imageList {
		if imgTbl[string(image)] {
			continue
		}
		bs, err := ExecuteCommand(fmt.Sprintf("docker pull %s -q", string(image)))
		if err != nil {
			return err
		}
		slog.DebugContext(ctx, "output", "output", string(bs))
	}
	return nil
}
