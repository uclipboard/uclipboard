package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/sirupsen/logrus"
)

func genClipboardReqBody(c string, logger *logrus.Entry) ([]byte, *model.Clipboard) {
	logger.Trace("into genClipboardReqBody")
	reqData := model.NewFullClipoard(c)
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		logger.Fatalf("parse response json body error: %s", err.Error())

	}
	return reqBody, reqData
}

func uploadStringData(s string, client *http.Client, c *model.Conf, logger *logrus.Entry) {
	logger.Trace("into UploadStringData")
	reqBody, _ := genClipboardReqBody(s, logger)
	resp, err := client.Post(model.UrlPushApi(c),
		"application/json", bytes.NewReader(reqBody))

	if err != nil {
		logger.Warnf("upload string data failed:%s", err.Error())
		return
	}
	defer resp.Body.Close()
}

func pullStringData(client *http.Client, c *model.Conf, logger *logrus.Entry) ([]byte, error) {
	logger.Trace("into pullStringData")
	pullApi := model.UrlPullApi(c)
	logger.Tracef("pullApi: %s", pullApi)
	resp, err := client.Get(pullApi)
	if err != nil {
		logger.Warnf("error sending req: %s", err)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("error reading response body: %s", err)
		return nil, err
	}

	resp.Body.Close()

	return body, nil
}

func newUClipboardHttpClient() *http.Client {
	return &http.Client{}
}

func uploadFile(filePath string, client *http.Client, c *model.Conf, logger *logrus.Entry) {
	logger.Trace("into UploadFile")
	file, err := os.Open(filePath)
	fileName := filepath.Base(filePath) //file.Name doesn't work in the right way on Windows platform
	if err != nil {
		logger.Fatalf("open file error: %v", err)
	}
	defer file.Close()
	// TODO:read file content type
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	logger.Tracef("upload file name: %s", fileName)
	part, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		logger.Fatalf("CreateFormField error: %v", err)
	}
	num, err := io.Copy(part, file)
	if err != nil {
		logger.Fatalf("copy file error: %v", err)
	}
	logger.Tracef("copy file size: %d", num)

	err = bodyWriter.Close()
	if err != nil {
		logger.Fatalf("close bodyWriter error: %v", err)
	}

	req, err := http.NewRequest("POST", model.UrlUploadApi(c), bodyBuf)
	if err != nil {
		logger.Fatalf("NewRequest error: %v", err)
	}
	fileContentType := bodyWriter.FormDataContentType()
	req.Header.Set("Content-Type", fileContentType)

	logger.Tracef("Content-Type: %s", fileContentType)
	hostname, err := os.Hostname()
	if err != nil {
		logger.Warnf("Get hostname error: %v", err)
	}
	// set hostname to header
	req.Header.Set("Hostname", hostname)

	resp, err := client.Do(req)
	if err != nil {
		logger.Fatalf("Upload file error: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("Read response body error: %v", err)
	}
	logger.Tracef("Response body: %s", respBody)
	fmt.Println(string(respBody))
}

func parseContentDisposition(contentDisposition string) string {
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return ""
	}
	return params["filename"]
}

func downloadFile(fileId string, client *http.Client, c *model.Conf, logger *logrus.Entry) {
	logger.Trace("into DownloadFile")
	logger.Tracef("fileId: %s", fileId)
	downloadUrl := model.UrlDownloadApi(c, fileId)
	logger.Tracef("downloadApi: %s", downloadUrl)
	res, err := client.Get(downloadUrl)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Download file failed:%d", res.StatusCode)
	}
	contentDisposition := res.Header.Get("Content-Disposition")
	logger.Tracef("Content-Disposition: %s", contentDisposition)
	fileName := parseContentDisposition(contentDisposition)
	file, err := os.Create(fileName)
	if err != nil {
		logger.Fatalf("Create file error: %v", err)
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return
	}

}

// if this clipboard content is a binary file, return the download url
func deteckAndconcatClipboardFileUrl(conf *model.Conf, clipboard *model.Clipboard) string {
	content := clipboard.Content
	if clipboard.ContentType == "binary" {
		content = model.UrlDownloadApi(conf, content)
	}
	return content
}
