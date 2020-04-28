package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
)

//ExtraMetaData silence lint
type ExtraMetaData struct {
}

//MetaData silence lint
type MetaData struct {
	Deleted          bool   `json:"deleted"`
	LastModified     int    `json:"lastModified"` //maybe uint? see golang example of json marshal
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

func getMetadataContent(visibleName string, parentUUID string, lastModified int) []byte {

	metadataContent := MetaData{false, lastModified, false, false, parentUUID, false, false, "DocumentType", 1, visibleName}
	content, _ := json.Marshal(metadataContent)
	return content
}

func main() {

	parentUUID := "75d25724-2cde-4872-8bf8-24e289b52476"
	lastModified := 1588012868000
	fileNameInput := "sample.pdf"
	visibleName := "sample"

	fmt.Println("hello there")
	fileUUID := uuid.New().String()
	fmt.Println(fileUUID)

	fileContent, _ := ioutil.ReadFile(fileNameInput)
	writeFile(fileUUID+".pdf", fileContent)

	fileContent = getDotContentContent("pdf")
	writeFile(fileUUID+".content", fileContent)

	fileContent = getMetadataContent(visibleName, parentUUID, lastModified)
	writeFile(fileUUID+".metadata", fileContent)
}
