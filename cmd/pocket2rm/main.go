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

// TODOs
// implement error handling
// - wrong/missing pocketCredentials
// - no internet
// images in files are not coming through
// -> external URLs. to support this, the xml should be parsed and files should be downloaded.
// local file system for debugging?
// - clean up repo structure: should not have to copies of utils
// - logs of service can grow very large?
// - what happens on first run with starting service? immediate action?
