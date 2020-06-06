package main

import (
	"fmt"
)

func main() {
	fmt.Println("start programm")
	var maxFiles uint = 10
	if reloadFileExists() {
		fmt.Println("reload file exists")
	} else {
		fmt.Println("no reload file")
		if !pocketFolderExists() {
			fmt.Println("no pocket folder")
			generatePocketFolder()
		}
		generateReloadFile()
		generateFiles(maxFiles)
	}
}
