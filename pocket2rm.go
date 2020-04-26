package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bmaupin/go-epub"
	"github.com/go-shiori/go-readability"
	"github.com/motemen/go-pocket/api"
	"github.com/motemen/go-pocket/auth"
	"gopkg.in/yaml.v3"
)

type pocketItem struct {
	id    string
	url   *url.URL
	added time.Time
	title string
}

func writeEpub(filePath string, title string, content string) error {
	e := epub.NewEpub(title)
	e.SetAuthor("pocket2rm")
	_, err := e.AddSection(content, title, "", "")

	if err != nil {
		return err
	}

	err = e.Write(filePath)

	if err != nil {
		return err
	}

	return nil
}

func getReadableArticle(url *url.URL) (string, string, error) {
	timeout, _ := time.ParseDuration("10s")
	article, err := readability.FromURL(url.String(), timeout)

	if err != nil {
		return "", "", err
	}

	return article.Title, article.Content, nil
}

//get interactive input. whitespace is stripped from return
func input(text string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(text)
	text, _ = reader.ReadString('\n')
	text = strings.Join(strings.Fields(text), "") //strip whitespace
	return text
}

//obtain consumerKey, accessToken and write to credentialsPath
func setup(credentialsPath string) error {

	consumerKey := input("Insert consumerKey: ")

	// setup local listener server for a confirmation
	ch := make(chan struct{})
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if req.URL.Path != "/" {
				http.Error(w, "Not Found", 404)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "Choice registered.")
			ch <- struct{}{}
		}))

	defer ts.Close()

	requestToken, err := auth.ObtainRequestToken(consumerKey, ts.URL)
	if err != nil {
		fmt.Println("Could not obtain request token: ", err)
		return err
	}

	//Open authorization URL in default browser for user to confirm application
	authorizationURL := auth.GenerateAuthorizationURL(requestToken, ts.URL)

	open(authorizationURL)

	<-ch //block untill request comes in

	authorization, err := auth.ObtainAccessToken(consumerKey, requestToken)
	if err != nil {
		fmt.Println("Could not obtain accessToken: ", err)
		return err
	}

	fmt.Println("Authorized.")

	credentials := make(map[string]string)
	credentials["consumerKey"] = consumerKey
	credentials["accessToken"] = authorization.AccessToken

	ymlContent, err := yaml.Marshal(credentials)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(credentialsPath, ymlContent, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Println("Setup successful. Wrote credentials to " + credentialsPath)
	return nil
}

func getCredentials(credentialsPath string) (string, string, error) {
	//return consumerKey, accessToken, error

	fileContent, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return "", "", err
	}

	var credentials map[string]string
	yaml.Unmarshal(fileContent, &credentials)

	return credentials["consumerKey"], credentials["accessToken"], nil
}

func getPocketItems(credentialsPath string) ([]pocketItem, error) {

	consumerKey, accessToken, err := getCredentials(credentialsPath)
	if err != nil {
		return []pocketItem{}, nil
	}

	client := api.NewClient(consumerKey, accessToken)
	var r *api.RetrieveOption
	retrieveResult, err := client.Retrieve(r)
	if err != nil {
		return []pocketItem{}, nil
	}

	var items []pocketItem
	for id, item := range retrieveResult.List {
		parsedURL, _ := url.Parse(item.ResolvedURL)
		items = append(items, pocketItem{id, parsedURL, time.Time(item.TimeAdded), item.Title()})
	}

	return items, nil
}

//generate filename from time added and title
func getFilename(timeAdded time.Time, title string, fileType string) string {
	// fileType: "epub" or "pdf"
	title = strings.Join(strings.Fields(title), "-")
	title = strings.Replace(title, "/", "_", -1)
	fileName := fmt.Sprintf("%s_%s", timeAdded.Format("20060102"), title)

	if fileType == "epub" && filepath.Ext(fileName) != ".epub" {
		fileName = fileName + ".epub"
	} else if fileType == "pdf" && filepath.Ext(fileName) != ".pdf" {
		fileName = fileName + ".pdf"
	}

	return fileName
}

func writePDF(filePath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)

	if err != nil {
		return err
	}

	return nil
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func main() {

	user, err := user.Current()
	if err != nil {
		fmt.Println("Could not get user")
		panic(1)
	}

	credentialsPath := filepath.Join(user.HomeDir, ".pocket2rm")

	argsWithProg := os.Args
	if len(argsWithProg) > 1 {
		if argsWithProg[1] == "setup" {
			setup(credentialsPath)
		}
		os.Exit(0)
	}

	articleFolder := "articles"
	err = os.MkdirAll(articleFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Could not create article folder: ", err)
	}

	pocketArticles, err := getPocketItems(credentialsPath)
	if err != nil {
		fmt.Println("Could not get pocket articles: ", err)
	}

	for i, pocketItem := range pocketArticles {
		fmt.Println(fmt.Sprintf("progress: %d/%d", i+1, len(pocketArticles)))

		extension := filepath.Ext(pocketItem.url.String())
		if extension == ".pdf" {
			fileName := getFilename(pocketItem.added, pocketItem.title, "pdf")
			filePath := filepath.Join(articleFolder, fileName)

			if _, err := os.Stat(filePath); err == nil {
				continue //file already exists
			}

			err := writePDF(filePath, pocketItem.url.String())
			if err != nil {
				fmt.Println("Could not get PDF for ", pocketItem.url.String(), err)
			}

		} else {
			fileName := getFilename(pocketItem.added, pocketItem.title, "epub")
			filePath := filepath.Join(articleFolder, fileName)

			if _, err := os.Stat(filePath); err == nil {
				continue //file already exists
			}

			title, content, err := getReadableArticle(pocketItem.url)
			if err != nil {
				fmt.Println("Could not get readable article for ", pocketItem.url.String(), err)
				continue
			}

			err = writeEpub(filePath, title, content)
			if err != nil {
				fmt.Println("Could not write epub for ", pocketItem.url.String())
			}
		}
	}

	fmt.Println("Finished proces. Files written to: " + articleFolder)
}
