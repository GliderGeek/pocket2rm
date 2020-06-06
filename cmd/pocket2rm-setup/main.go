package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/motemen/go-pocket/auth"
	"gopkg.in/yaml.v3"
)

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

	<-ch //block until request comes in

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
	setup(credentialsPath)
}
