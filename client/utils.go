package client

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/sirupsen/logrus"
)

func genClipboardReqBody(c string, logger *logrus.Entry) ([]byte, *model.Clipboard) {
	reqData := model.NewFullClipoard(c)
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		logger.Panicf("parse response json body error: %s", err.Error())

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
	pullApi := model.UrlPullApi(c)
	logger.Tracef("pullApi:%s", pullApi)
	resp, err := client.Get(pullApi)
	if err != nil {
		logger.Warnf("Error sending req: %s", err)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("Error reading response body: %s", err)
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
	if err != nil {
		logger.Panicf("Open file error: %v", err)
		return
	}
	defer file.Close()
	// TODO:read file content type
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	part, err := bodyWriter.CreateFormFile("file", file.Name())
	if err != nil {
		logger.Panicf("CreateFormField error: %v", err)
	}
	num, err := io.Copy(part, file)
	if err != nil {
		logger.Panicf("Copy file error: %v", err)
	}
	err = bodyWriter.Close()
	if err != nil {
		logger.Panicf("Close bodyWriter error: %v", err)
	}

	logger.Tracef("Copy file size: %d", num)
	req, err := http.NewRequest("POST", model.UrlUploadApi(c), bodyBuf)
	if err != nil {
		logger.Panicf("NewRequest error: %v", err)
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
		logger.Panicf("Upload file error: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Panicf("Read response body error: %v", err)
	}
	logger.Tracef("Response body: %s", respBody)

}

// func downloadFile(fileName string, client *http.Client, c *model.Conf, logger *logrus.Entry) (*http.Response, error) {
// 	logger.Trace("into DownloadFile")
// 	var downloadApi string
// 	if fileName == "" {
// 		downloadApi = model.UrlLatestApi(c)
// 	} else {
// 		downloadApi = model.UrlDownloadApi(c, fileName)
// 	}
// 	logger.Tracef("downloadApi:%s", downloadApi)
// 	resp, err := client.Get(downloadApi)
// 	if err != nil {
// 		logger.Warnf("Error sending req: %s", err)
// 		return nil, err
// 	}
// 	return resp, nil
// }
