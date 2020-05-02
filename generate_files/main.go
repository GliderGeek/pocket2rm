package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

//ExtraMetaData silence lint
type ExtraMetaData struct {
}

//Config silence lint
type Config struct {
	ConsumerKey      string `yaml:"consumerKey"`
	AccessToken      string `yaml:"accessToken"`
	ReloadUUID       string `yaml:"reloadUUID"`
	PocketFolderUUID string `yaml:"pocketFolderUUID"`
}

//MetaData silence lint
type MetaData struct {
	Deleted          bool   `json:"deleted"`
	LastModified     uint   `json:"lastModified"`
	Metadatamodified bool   `json:"metadatamodified"`
	Modified         bool   `json:"modified"`
	Parent           string `json:"parent"` //uuid
	Pinned           bool   `json:"pinned"`
	Synced           bool   `json:"synced"`
	Type             string `json:"type"`
	Version          int    `json:"version"`
	VisibleName      string `json:"visibleName"`
}

//Transform silence lint
type Transform struct {
	M11 int `json:"m11"`
	M12 int `json:"m12"`
	M13 int `json:"m13"`
	M21 int `json:"m21"`
	M22 int `json:"m22"`
	M23 int `json:"m23"`
	M31 int `json:"m31"`
	M32 int `json:"m32"`
	M33 int `json:"m33"`
}

// DocumentContent silence lint
type DocumentContent struct {
	ExtraMetadata  ExtraMetaData `json:"extraMetadata"`
	FileType       string        `json:"fileType"`
	FontName       string        `json:"fontName"`
	LastOpenedPage int           `json:"lastOpenedPage"`
	LineHeight     int           `json:"lineHeight"`
	Margins        int           `json:"margins"`
	Orientation    string        `json:"orientation"`
	PageCount      int           `json:"pageCount"`
	TextScale      int           `json:"textScale"`
	Transform      Transform     `json:"transform"`
}

func articeFolderPath() string {
	return "/home/root/.local/share/remarkable/xochitl/"
}

func getConfigPath() string {
	return "/home/root/.pocket2rm"
}
func writeFile(fileName string, fileContent []byte) {

	// write the whole body at once
	err := ioutil.WriteFile(fileName, fileContent, 0644)
	if err != nil {
		panic(err)
	}
}

func getDotContentContent(fileType string) []byte {
	transform := Transform{1, 0, 0, 0, 1, 0, 0, 0, 1}
	docContent := DocumentContent{ExtraMetaData{}, fileType, "", 0, -1, 100, "portrait", 1, 1, transform}
	content, _ := json.Marshal(docContent)
	return content
}

func getMetadataContent(visibleName string, parentUUID string, lastModified uint) []byte {
	metadataContent := MetaData{false, lastModified, false, false, parentUUID, false, false, "DocumentType", 1, visibleName}
	content, _ := json.Marshal(metadataContent)
	return content
}

//check both if file is present and if metadata deleted=false
func pdfIsPresent(uuid string) bool {

	pdfPath := filepath.Join(articeFolderPath(), uuid+".pdf")
	metadaPath := filepath.Join(articeFolderPath(), uuid+".metadata")
	_, err := os.Stat(pdfPath)

	if os.IsNotExist(err) {
		return false
	}

	fileContent, _ := ioutil.ReadFile(metadaPath)
	var metadata MetaData
	json.Unmarshal(fileContent, &metadata)
	return !metadata.Deleted
}

func reloadFileExists() bool {
	config := getConfig()
	return pdfIsPresent(config.ReloadUUID)
}

func writeReloadUUID(newReloadUUID string) {
	configPath := getConfigPath()
	config := getConfig()
	config.ReloadUUID = newReloadUUID
	ymlContent, _ := yaml.Marshal(config)
	_ = ioutil.WriteFile(configPath, ymlContent, os.ModePerm)
}

func generateReloadFile() {
	fmt.Println("writing reloadfile")
	fileContent, _ := ioutil.ReadFile("sample.pdf")
	reloadFileUUID := generatePDF("remove to sync", fileContent)
	writeReloadUUID(reloadFileUUID)
}

// uuid is returned
func generateEpub(visibleName string, fileContent []byte) string {

	var lastModified uint = 1 //TODO number too big. maybe need custom marshal: http://choly.ca/post/go-json-marshalling/

	config := getConfig()
	fileUUID := uuid.New().String()

	fileName := filepath.Join(articeFolderPath(), fileUUID+".epub")
	writeFile(fileName, fileContent)

	fileContent = getDotContentContent("epub")
	fileName = filepath.Join(articeFolderPath(), fileUUID+".content")
	writeFile(fileName, fileContent)

	fileContent = getMetadataContent(visibleName, config.PocketFolderUUID, lastModified)
	fileName = filepath.Join(articeFolderPath(), fileUUID+".metadata")
	writeFile(fileName, fileContent)

	return fileUUID
}

func generatePDF(visibleName string, fileContent []byte) string {

	var lastModified uint = 1 //TODO number too big. maybe need custom marshal: http://choly.ca/post/go-json-marshalling/

	config := getConfig()
	fileUUID := uuid.New().String()

	fileName := filepath.Join(articeFolderPath(), fileUUID+".pdf")
	writeFile(fileName, fileContent)

	fileContent = getDotContentContent("pdf")
	fileName = filepath.Join(articeFolderPath(), fileUUID+".content")
	writeFile(fileName, fileContent)

	fileContent = getMetadataContent(visibleName, config.PocketFolderUUID, lastModified)
	fileName = filepath.Join(articeFolderPath(), fileUUID+".metadata")
	writeFile(fileName, fileContent)

	return fileUUID
}

func getConfig() Config {
	fileContent, _ := ioutil.ReadFile(getConfigPath())
	var config Config
	yaml.Unmarshal(fileContent, &config)
	return config
}

func generateFiles() {
	fmt.Println("generating files")
	fileNameInput := "sample.pdf"
	fileContent, _ := ioutil.ReadFile(fileNameInput)
	visibleName := "sample pdf " + uuid.New().String()[0:4]
	generatePDF(visibleName, fileContent)

	fileNameInput = "sample.epub"
	fileContent, _ = ioutil.ReadFile(fileNameInput)
	visibleName = "sample epub " + uuid.New().String()[0:4]
	generateEpub(visibleName, fileContent)
}

func stopXochitl() {
	cmd := exec.Command("systemctl", "stop", "xochitl")
	cmd.Run()
}

func restartXochitl() {
	cmd := exec.Command("systemctl", "restart", "xochitl")
	cmd.Run()
}

func main() {
	fmt.Println("start programm")
	time.Sleep(3 * time.Second)

	for {
		fmt.Println("sleep for 5 secs")
		time.Sleep(5 * time.Second)
		if reloadFileExists() {
			fmt.Println("reload file exists")
		} else {
			fmt.Println("no reload file")
			stopXochitl()
			generateFiles()
			generateReloadFile()
			restartXochitl()
		}
	}
}

//current flow:
// - enable ssh
// - make ~/.pocket2rm with "consumerKey", "accessToken", "reloadUUID", "pocketFolderUUID"
// - inside generate_files folder: `GOOS=linux GOARCH=arm GOARM=5 go build -o pocket2rm.arm`
// - scp ~/.pocket2rm root@10.11.99.1:/home/root/.
// - scp pocket2rm.arm root@10.11.99.1:/home/root/.
// - ssh@10.11.99.1
// - ./pocket2rm
// - remove sync file

// TODOs
// debug issue with reloadFile. seems to not work after power cycle
// - reload file is there, but still workflow is started
// - is uuid changed?
// get actual pocket articles
// - limit amount
// - keep track of what has been processed (URLs in config?)
// local file mimic for debugging?
// - golang command for local folder
// run as service

// to find folder UUID:
//grep \"visibleName\":\ \"Pocket\" *.metadata -l -i
