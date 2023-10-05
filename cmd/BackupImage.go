package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type BackupImage struct {
	dockerImages []*Image
}

var backupImage *BackupImage

func GetBackupImage() *BackupImage {
	if backupImage == nil {
		backupImage = &BackupImage{}
	}
	return backupImage
}
func (this *BackupImage) Loop() {
	var err error
	this.dockerImages, err = this.ListTarsForBackupImages(GetConfig().GetBackupDir())
	if err != nil {
		fmt.Println("获取备份列表失败")
		return
	}
	GetListImage().DumpDockerImages(this.dockerImages)
	for {
		fmt.Print("备份> q - 返回上级,l - 列表 或 输入序号 - 进行备份。\n")
		text, _ := GetReader().ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "q" {
			break
		}
		if text == "l" {
			GetListImage().DumpDockerImages(this.dockerImages)
			continue
		}
		if backupNumber, err := strconv.Atoi(text); err == nil {
			this.Backup(backupNumber, this.dockerImages)
		} else {
			fmt.Println("无效的输入")
		}
	}
}

func (this *BackupImage) Backup(backupIndex int, dockerImages []*Image) {
	if backupIndex >= 0 {
		if len(dockerImages) <= backupIndex {
			fmt.Println("没有这个镜像")
			return
		}
		dockerImage := dockerImages[backupIndex]
		if dockerImage == nil {
			fmt.Println("没有这个镜像")
			return
		}
		if dockerImage.IsBackup == "Y" {
			fmt.Println("这个镜像已经备份了")
			return
		}
		fmt.Printf("备份目标:%s\n", dockerImage.Tag)
		if ConfirmWithTip("即将进行备份,确认请按 y 或 其他键 取消") {
			return
		}

		if err := this.SaveImageAsTar(dockerImage.ImageId, dockerImage.Tag, fmt.Sprintf("%s.tar", dockerImage.TarFile)); err != nil {
			fmt.Println("镜像备份失败,%s", err.Error())
			return
		}
		dockerImage.IsBackup = "Y"
	}
}

func (this *BackupImage) SaveImageAsTar(imageID, tagName, tarFile string) error {
	if _, err := os.Stat(tarFile); err == nil {
		return errors.New("文件已存在不必备份!")
	}

	//docker save -o ../docker-tars/postgres-9.6.tar chenjingdong/postgres:9.6
	// 执行 docker load 命令
	cmd := exec.Command("docker", "save", "-o", tarFile, tagName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("不能备份镜像: %v\n%s", err, output))
	}

	fmt.Println("成功备份镜像!")
	return nil

	//cli, err := client.NewClientWithOpts(client.FromEnv)
	//if err != nil {
	//	return err
	//}
	//cli.ImageTag(context.Background(), imageID, tagName)
	//
	//r, err := cli.ImageSave(context.Background(), []string{imageID})
	//
	//if err != nil {
	//	return err
	//}
	//defer r.Close()
	//
	//f, err := os.Create(tarFile)
	//if err != nil {
	//	return err
	//}
	//defer f.Close()
	//
	//_, err = io.Copy(f, r)
	//return err
}

func (this *BackupImage) ListTarsForBackupImages(outputDir string) ([]*Image, error) {
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
		tarFile := fmt.Sprintf("%s/%s", outputDir, FormatTarName(theTog))
		if _, err := os.Stat(fmt.Sprintf("%s.tar", tarFile)); err == nil {
			continue
		}

		dockerImage := &Image{
			Tag:      theTog,
			Tags:     tags,
			IsBackup: "N",
			ImageId:  image.ID,
		}
		dockerImages = append(dockerImages, dockerImage)
		dockerImage.TarFile = tarFile
		dockerImage.Name = theTog

	}

	return dockerImages, nil
}
