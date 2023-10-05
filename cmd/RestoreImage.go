package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
)

type RestoreInfo struct {
	TarFile  string
	Restored string
}

type RestoreImage struct {
	allTarsInBackupDir []*RestoreInfo
}

var restoreImage *RestoreImage

func GetRestoreImage() *RestoreImage {
	if restoreImage == nil {
		restoreImage = &RestoreImage{}
	}
	return restoreImage
}

func (this *RestoreImage) restoreFromTar(backupNumber int, restoreInfo []*RestoreInfo) {
	if restoreInfo[backupNumber].Restored == "Y" {
		fmt.Printf("文件已经还原过,不必再操作 %s\n", restoreInfo[backupNumber].TarFile)
		return
	}
	tarFile := path.Join(GetConfig().GetBackupDir(), restoreInfo[backupNumber].TarFile)
	fmt.Printf("还原文件 %s", restoreInfo[backupNumber].TarFile)
	if ConfirmWithTip("确定还原,确认请按 y 或 其他键 取消") {
		return
	}
	tarFile = fmt.Sprintf("%s.tar", tarFile)
	fmt.Printf("还原文件 %s\n", tarFile)
	if err := this.loadImageFromTar(tarFile); err == nil {
		restoreInfo[backupNumber].Restored = "Y"
	}
}

func (this *RestoreImage) Loop() {
	dockerImages, err := GetListImage().ListTarsForImages(GetConfig().GetBackupDir())
	if err != nil {
		fmt.Println("获取备份列表失败")
		return
	}

	//从 allTarsInBackupDir 去除 dockerImages 中的 tar4Image 文件
	tarsOfExistedImages := map[string]string{}
	for _, dockerImage := range dockerImages {
		tarsOfExistedImages[ExtractFileName(dockerImage.TarFile)] = "Y"
	}
	this.allTarsInBackupDir = this.listTarsInBackupDir(GetConfig().GetBackupDir(), tarsOfExistedImages)
	this.dumpDockerRestoreInfo(this.allTarsInBackupDir)
	for {
		fmt.Print("还原> q - 返回上级,l - 列表 或 输入序号 - 进行还原。\n")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "q" {
			break
		}
		if text == "l" {
			this.dumpDockerRestoreInfo(this.allTarsInBackupDir)
			continue
		}
		if backupNumber, err := strconv.Atoi(text); err == nil {
			this.restoreFromTar(backupNumber, this.allTarsInBackupDir)
		} else {
			fmt.Println("无效的输入")
		}
	}
}
func (this *RestoreImage) dumpDockerRestoreInfo(restoreInfos []*RestoreInfo) {
	writer := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', 0)
	fmt.Fprintf(writer, "序号\t 已还原 \t 备份路径 \n")
	for i, restoreInfo := range restoreInfos {
		fmt.Fprintf(writer, "%d\t %s \t %s\n", i, restoreInfo.Restored, restoreInfo.TarFile)
	}
	writer.Flush()
}
func (this *RestoreImage) loadImageFromTar(tarFilePath string) error {
	// 检查 tar 文件是否存在
	if _, err := os.Stat(tarFilePath); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("压缩 %s 不存在", tarFilePath))
	}

	// 执行 docker load 命令
	cmd := exec.Command("docker", "load", "-i", tarFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("不能恢复镜像: %v\n%s", err, output))
	}

	fmt.Println("成功恢复镜像!")
	return nil
}
func (this *RestoreImage) listTarsInBackupDir(dir string, excludeTars map[string]string) []*RestoreInfo {
	//list4Image all file in dir
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("读取目录失败")
		return []*RestoreInfo{}
	}

	//只保留.tar作为扩展名的文件
	var restoreInfos []*RestoreInfo
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar") {
			prefixOfFileName := strings.Replace(file.Name(), ".tar", "", 1)
			if _, ok := excludeTars[prefixOfFileName]; !ok {
				restoreInfos = append(restoreInfos, &RestoreInfo{
					TarFile:  prefixOfFileName,
					Restored: "N",
				})
			}
		}
	}
	return restoreInfos
}
