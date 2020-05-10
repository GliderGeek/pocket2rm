package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	pdf "github.com/balacode/one-file-pdf"
	"github.com/bmaupin/go-epub"
	"github.com/go-shiori/go-readability"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

//ExtraMetaData silence lint
type ExtraMetaData struct {
}

//Config silence lint
type Config struct {
	ConsumerKey      string   `yaml:"consumerKey"`
	AccessToken      string   `yaml:"accessToken"`
	ReloadUUID       string   `yaml:"reloadUUID"`
	PocketFolderUUID string   `yaml:"pocketFolderUUID"`
	HandledArticles  []string `yaml:"handledArticles"` //id of article //TODO: should be converted to set for better time complexity
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

type pocketItem struct {
	id    string
	url   *url.URL
	added time.Time
	title string
}

//ByAdded silence lint
type ByAdded []pocketItem

func (a ByAdded) Len() int           { return len(a) }
func (a ByAdded) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAdded) Less(i, j int) bool { return a[i].added.Before(a[j].added) }

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

// PocketResult silence lint
type PocketResult struct {
	List     map[string]Item
	Status   int
	Complete int
	Since    int
}

// Item silence lint
type Item struct {
	ItemID        string `json:"item_id"`
	ResolvedID    string `json:"resolved_id"`
	GivenURL      string `json:"given_url"`
	ResolvedURL   string `json:"resolved_url"`
	GivenTitle    string `json:"given_title"`
	ResolvedTitle string `json:"resolved_title"`
	IsArticle     int    `json:"is_article,string"`
	TimeAdded     Time   `json:"time_added"`
}

// PocketRetrieve silence lint
type PocketRetrieve struct {
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`
}

// Time silence lint
type Time time.Time

//UnmarshalJSON silence lint
func (t *Time) UnmarshalJSON(b []byte) error {
	i, err := strconv.ParseInt(string(bytes.Trim(b, `"`)), 10, 64)
	if err != nil {
		return err
	}

	*t = Time(time.Unix(i, 0))

	return nil
}

// Title returns ResolvedTitle or GivenTitle
func (item Item) Title() string {
	title := item.ResolvedTitle
	if title == "" {
		title = item.GivenTitle
	}
	return title
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

func writeConfig(config Config) {
	configPath := getConfigPath()
	ymlContent, _ := yaml.Marshal(config)
	_ = ioutil.WriteFile(configPath, ymlContent, os.ModePerm)
}

func generateReloadFile() {
	fmt.Println("writing reloadfile")
	var pdf = pdf.NewPDF("A4")

	pdf.SetUnits("cm").
		SetFont("Helvetica-Bold", 100).
		SetColor("Black")
	pdf.SetXY(3.5, 5).
		DrawText("Remove")
	pdf.SetXY(9, 10).
		DrawText("to")
	pdf.SetXY(6.5, 15).
		DrawText("Sync")
	fileContent := pdf.Bytes()

	reloadFileUUID := generatePDF("remove to sync", fileContent)
	config := getConfig()
	config.ReloadUUID = reloadFileUUID
	writeConfig(config)
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

func createPDFFileContent(url string) []byte {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return content
}

func createEpubFileContent(title string, content string) []byte {
	e := epub.NewEpub(title)
	e.SetAuthor("pocket2rm")
	_, _ = e.AddSection(content, title, "", "")

	tmpName := "/tmp/epub" + uuid.New().String()[0:5] + ".epub"
	_ = e.Write(tmpName)
	defer os.Remove(tmpName)

	fileContent, _ := ioutil.ReadFile(tmpName)
	return fileContent
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

func getPocketItems() ([]pocketItem, error) {
	// unfortunately cannot use github.com/motemen/go-pocket
	// because of 32bit architecture
	// Item.ItemID in github.com/motemen/go-pocket is int, which cannot store enough
	// therefore the necessary types and functions have been copied and adapted

	config := getConfig()

	retrieveResult := &PocketResult{}
	body, _ := json.Marshal(PocketRetrieve{config.ConsumerKey, config.AccessToken})

	req, _ := http.NewRequest("POST", "https://getpocket.com/v3/get", bytes.NewReader(body))
	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []pocketItem{}, err
	}

	if resp.StatusCode != 200 {
		return []pocketItem{}, fmt.Errorf("got response %d; X-Error=[%s]", resp.StatusCode, resp.Header.Get("X-Error"))
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(retrieveResult)
	if err != nil {
		return []pocketItem{}, err
	}

	var items []pocketItem
	for id, item := range retrieveResult.List {
		parsedURL, _ := url.Parse(item.ResolvedURL)
		items = append(items, pocketItem{id, parsedURL, time.Time(item.TimeAdded), item.Title()})
	}

	// sort by latest added article first
	sort.Sort(sort.Reverse(ByAdded(items)))
	return items, nil
}

//generate filename from time added and title
func getFilename(timeAdded time.Time, title string) string {
	// fileType: "epub" or "pdf"
	title = strings.Join(strings.Fields(title), "-")
	title = strings.Replace(title, "/", "_", -1)
	fileName := fmt.Sprintf("%s_%s", timeAdded.Format("20060102"), title)
	return fileName
}

func alreadyHandled(article pocketItem) bool {
	config := getConfig()
	for _, articleID := range config.HandledArticles {
		if article.id == articleID {
			return true
		}
	}
	return false
}

func registerHandled(article pocketItem) {
	config := getConfig()
	config.HandledArticles = append(config.HandledArticles, article.id)
	writeConfig(config)
}

func getReadableArticle(url *url.URL) (string, string, error) {
	timeout, _ := time.ParseDuration("10s")
	article, err := readability.FromURL(url.String(), timeout)

	if err != nil {
		return "", "", err
	}

	return article.Title, article.Content, nil
}

func generateFiles(maxArticles uint) error {
	fmt.Println("inside generateFiles")
	pocketArticles, err := getPocketItems()
	if err != nil {
		fmt.Println("Could not get pocket articles: ", err)
		return err
	}

	var processed uint = 0
	for _, pocketItem := range pocketArticles {
		if alreadyHandled(pocketItem) {
			fmt.Println("already handled")
			continue
		}

		fileName := getFilename(pocketItem.added, pocketItem.title)
		extension := filepath.Ext(pocketItem.url.String())
		if extension == ".pdf" {
			fileContent := createPDFFileContent(pocketItem.url.String())
			generatePDF(fileName, fileContent)
		} else {
			title, XMLcontent, err := getReadableArticle(pocketItem.url)
			if err != nil {
				fmt.Println("Could not get readable article")
				continue
			}
			fileContent := createEpubFileContent(title, XMLcontent)
			generateEpub(fileName, fileContent)
		}

		registerHandled(pocketItem)
		processed++
		fmt.Println(fmt.Sprintf("progress: %d/%d", processed, maxArticles))
		if processed == maxArticles {
			break
		}
	}

	return nil
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
	var maxFiles uint = 10
	for {
		fmt.Println("sleep for 10 secs")
		time.Sleep(10 * time.Second)
		if reloadFileExists() {
			fmt.Println("reload file exists")
		} else {
			fmt.Println("no reload file")
			stopXochitl()
			generateFiles(maxFiles)
			generateReloadFile()
			restartXochitl()
		}
	}
}

//current flow:
// Setup
// - enable ssh
// - make ~/.pocket2rm with "consumerKey", "accessToken", "reloadUUID", "pocketFolderUUID"
// - inside generate_files folder: `GOOS=linux GOARCH=arm GOARM=7 go build -o pocket2rm.arm`
// - scp ~/.pocket2rm root@10.11.99.1:/home/root/.
// - scp pocket2rm.arm root@10.11.99.1:/home/root/.

//Run
// - ssh@10.11.99.1
// - ./pocket2rm
// - remove sync file

// TODOs
// debug issue with reloadFile. sometimes double file? after power cycle?
// - reload file is there, but still workflow is started
// - is uuid changed? seems not
// implement error handling
// - wrong/missing pocketCredentials
// - no internet
// logging instead of printing. enabled/disabled with flag?
// local file system for debugging?
// - golang command for local folder
// run as service
// images in files are not coming through
// title on top of file?

// to find folder UUID:
//grep \"visibleName\":\ \"Pocket\" *.metadata -l -i
