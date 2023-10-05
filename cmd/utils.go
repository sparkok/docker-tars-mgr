package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ConfirmWithTip(tipMsg string) bool {
	fmt.Println(tipMsg)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text != "y" {
		fmt.Printf("您终止了操作\n")
		return true
	}
	return false
}
func ExtractFileName(fileFullPath string) string {
	//从文件路径中提取文件名
	return fileFullPath[strings.LastIndexAny(fileFullPath, "\\/")+1:]
}
func FormatTarName(tag string) string {
	tag = strings.Replace(tag, "/", ".", -1)
	tag = strings.Replace(tag, ":", ".", -1)
	//tag = strings.Replace(tag, ".", "_", -1)
	return tag
}

var reader *bufio.Reader

func InitReader() {
	reader = bufio.NewReader(os.Stdin)
}
func GetReader() *bufio.Reader {
	return reader
}
