package main

import (
	"bufio"
	. "docker-tars-mgr/cmd"
	"fmt"
	"os"
	"strings"
)

var backupDir string
var reader *bufio.Reader

func runLoop() {
	reader = bufio.NewReader(os.Stdin)
	for {
		fmt.Print("q - 退出; l - 列出备份情况; b - 进入备份指令。r - 进入还原模块\n")
		fmt.Print("输入指令: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch text {
		case "l":
			GetListImage().List()
		case "b":
			GetBackupImage().Loop()
		case "q":
			quit()
			return
		case "r":
			GetRestoreImage().Loop()
		default:
			fmt.Println("未知命令")
		}
	}
}

func quit() {
	fmt.Println("退出...\n")
	os.Exit(0)
}

func main() {
	InitReader()
	if len(backupDir) == 0 {
		backupDir = "./backup"
	}
	os.MkdirAll(backupDir, 0755)
	runLoop()
}
