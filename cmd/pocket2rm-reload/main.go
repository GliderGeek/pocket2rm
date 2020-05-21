package main

import (
	"fmt"
	"os/exec"
	"time"
)

func startPocket2rm() {
	cmd := exec.Command("systemctl", "restart", "pocket2rm")
	cmd.Run()
}

func main() {
	fmt.Println("start programm")
	for {
		fmt.Println("sleep for 10 secs")
		time.Sleep(10 * time.Second)
		if reloadFileExists() {
			fmt.Println("reload file exists")
		} else {
			fmt.Println("no reload file, starting pocket2rm")
			startPocket2rm()
		}
	}
}
