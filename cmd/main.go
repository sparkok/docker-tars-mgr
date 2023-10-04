package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

var backupDir string
var reader *bufio.Reader

// 创建镜像列表
func listTarsForImages(outputDir string) ([]*Image, error) {
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
			IsBackup: "X",
			ImageId:  image.ID,
		}
		dockerImages = append(dockerImages, dockerImage)
		dockerImage.TarFile = fmt.Sprintf("%s/%s", outputDir, formatTarName(theTog))
		dockerImage.Name = theTog
		if _, err := os.Stat(fmt.Sprintf("%s.tar", dockerImage.TarFile)); err == nil {
			dockerImage.IsBackup = "Y"

		}
	}

	return dockerImages, nil
}

func formatTarName(tag string) string {
	tag = strings.Replace(tag, "/", ".", -1)
	tag = strings.Replace(tag, ":", ".", -1)
	//tag = strings.Replace(tag, ".", "_", -1)
	return tag
}

func saveImageAsTar(imageID, tarFile string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	r, err := cli.ImageSave(context.Background(), []string{imageID})
	if err != nil {
		return err
	}
	defer r.Close()

	f, err := os.Create(tarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}
func DumpDockerImages(dockerImages []*Image) {
	writer := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', 0)
	fmt.Fprintf(writer, "index\t bak \t tar \t tag\t\n")
	for i, dockerImage := range dockerImages {
		for j, tag := range dockerImage.Tags {
			fmt.Fprintf(writer, "%d.%d \t %s\t %s.tar\t %s\n", i, j, dockerImage.IsBackup, dockerImage.TarFile, tag)
		}
	}
	writer.Flush()
}

type Options struct {
	List  string
	Tar   float64
	UnTar string
}

func parseOptions() (*Options, error) {

	var opts Options

	flag.StringVar(&opts.List, "list", "list", "list images")
	flag.Float64Var(&opts.Tar, "tar", -1, "tar image")
	flag.StringVar(&opts.UnTar, "untar", "", "untar image")

	// 解析参数
	flag.Parse()

	// 参数校验
	return &opts, nil

}
func runLoop() {
	reader = bufio.NewReader(os.Stdin)
	for {
		fmt.Print("q - 退出; l - 列出备份情况; t - 进入备份指令。\n")
		fmt.Print("输入指令: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch text {
		case "list", "l":
			list()
		case "tar", "t":
			tarLoop(reader)
		case "quit", "q":
			quit()
			return
		default:
			fmt.Println("未知命令")
		}
	}
}

func tarLoop(reader *bufio.Reader) {

	for {
		fmt.Print("备份> q - 退出,l - 列表 或 输入序号进行备份。\n")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "q" {
			break
		}
		if text == "l" {
			list()
			continue
		}
		if backupNumber, err := strconv.Atoi(text); err == nil {
			tar(backupNumber)
		} else {
			fmt.Println("Invalid input")
		}
	}

}

func list() {
	dockerImages, err := listTarsForImages("./backup")
	if err != nil {
		panic(err)
	}
	DumpDockerImages(dockerImages)
}

func tar(backupIndex int) {
	if backupIndex >= 0 {
		dockerImages, err := listTarsForImages("./backup")
		if err != nil {
			panic(err)
		}
		if len(dockerImages) <= backupIndex {
			fmt.Println("没有这个镜像")
			return
		}
		dockerImage := dockerImages[backupIndex]
		if dockerImage == nil {
			fmt.Println("没有这个镜像")
			return
		}
		fmt.Printf("备份目标:%s\n", dockerImage.Tag)
		fmt.Println("即将进行备份,确认请按 y 或 其他键 取消")

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text != "y" {
			fmt.Printf("您终止了操作\n")
			return
		}

		if err = saveImageAsTar(dockerImage.ImageId, dockerImage.TarFile); err != nil {
			fmt.Println("failed to save image as tar,%s", err.Error())
			return
		}
	}
}

func quit() {
	fmt.Println("Exiting...")
	os.Exit(0)
}

func main() {
	backupDir = "./backup"
	os.MkdirAll(backupDir, 0755)
	runLoop()
}
