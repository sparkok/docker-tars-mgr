package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os"
	"strings"
	"text/tabwriter"
)

type Image struct {
	Tag       string
	BackupTag string
	TarFile   string
	IsBackup  string
	Name      string
	Tags      []string
	ImageId   string
}
type ListImage struct {
}

var listImage *ListImage

func GetListImage() *ListImage {
	if listImage == nil {
		listImage = &ListImage{}
	}
	return listImage
}
func (this *ListImage) List() {
	dockerImages, err := this.ListTarsForImages(GetConfig().GetBackupDir())
	if err != nil {
		panic(err)
	}
	this.DumpDockerImages(dockerImages)
}

func (this *ListImage) DumpDockerImages(dockerImages []*Image) {
	writer := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', 0)
	fmt.Fprintf(writer, "序号\t 已备份 \t 备份路径 \t 标签\t\n")
	for i, dockerImage := range dockerImages {
		for _, tag := range dockerImage.Tags {
			fmt.Fprintf(writer, "%d \t %s\t %s.tar\t %s\n", i, dockerImage.IsBackup, dockerImage.TarFile, tag)
		}
	}
	writer.Flush()
}

// 创建镜像列表
func (this *ListImage) ListTarsForImages(outputDir string) ([]*Image, error) {
	var dockerImages = []*Image{}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		tags := image.RepoTags
		if len(tags) == 0 {
			continue
		}
		isSystemImage := false
		for _, tag := range tags {
			if strings.Index(tag, "system") == 0 ||
				strings.Index(tag, "hubproxy.") == 0 ||
				strings.Index(tag, "k8s.") == 0 {
				//fmt.Printf("system is %s\n", tag)
				isSystemImage = true
				break
			}
		}
		if isSystemImage {
			continue
		}
		theTog := tags[0]
		dockerImage := &Image{
			Tag:      theTog,
			Tags:     tags,
			IsBackup: "N",
			ImageId:  image.ID,
		}
		dockerImages = append(dockerImages, dockerImage)
		dockerImage.TarFile = fmt.Sprintf("%s/%s", outputDir, FormatTarName(theTog))
		dockerImage.Name = theTog
		if _, err := os.Stat(fmt.Sprintf("%s.tar", dockerImage.TarFile)); err == nil {
			dockerImage.IsBackup = "Y"

		}
	}

	return dockerImages, nil
}
