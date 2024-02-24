package main

import (
	. "docker-tars-mgr/cmd"
	"fmt"
	"os"
	"strings"
)

func runLoop() {
	InitReader()
	for {
		fmt.Print("q - 退出; l - 列出备份情况; b - 进入备份指令。r - 进入还原模块\n")
		fmt.Print("输入指令: ")
		text, _ := GetReader().ReadString('\n')
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
	os.MkdirAll(GetConfig().GetBackupDir(), 0755)
	runLoop()
}
