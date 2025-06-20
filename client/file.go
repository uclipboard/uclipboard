// Those file operate is so big that I extract them from utils.go
package client

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/uclipboard/uclipboard/model"
)

func UploadFile(filePath string, client *HeaderHttpClient, c *model.UContext, logger *logrus.Entry) {
	logger.Trace("into UploadFile")
	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatalf("open file error: %v", err)
	}

	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		logger.Fatalf("stat file error: %v", err)
	}
	fileName := fileStat.Name()

	fmt.Printf("uploading file: %s\t file_size: %vKiB\n", fileName, float32(fileStat.Size()/1024))

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	logger.Debugf("upload file name: %s", fileName)
	part, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		logger.Fatalf("CreateFormFile error: %v", err)
	}
	num, err := io.Copy(part, file)
	if err != nil {
		logger.Fatalf("copy file to form buffer error: %v", err)
	}
	logger.Debugf("copy file size: %db", num)

	err = bodyWriter.Close()
	if err != nil {
		logger.Fatalf("close bodyWriter error: %v", err)
	}

	req, err := http.NewRequest("POST", model.UrlUploadApi(c), bodyBuf)
	if err != nil {
		logger.Fatalf("NewRequest error: %v", err)
	}
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	hostname, err := os.Hostname()
	if err != nil {
		logger.Warnf("Get hostname error: %v,reset to uclipboard_client", err)
		hostname = "uclipboard_client"
	}
	// add hostname to header
	req.Header.Set("Hostname", hostname)
	resp, err := client.Do(req)
	if err != nil {
		logger.Fatalf("Upload file error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Fatalf("Upload file failed: %s", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("Read response body error: %v", err)
	}
	logger.Tracef("Response body: %s", respBody)
	respData, err := ParseUploadInfomation(respBody)
	if err != nil {
		logger.Fatalf("parse upload response body error: %v", err)
	}

	fmt.Printf("upload file success, id: %v, name: %v, lifetime: %ss, access token: %s\n",
		respData.Id, respData.Name, strconv.FormatFloat(float64(respData.LifeTime), 'f', -1, 64),
		respData.Token)

}

func parseContentDisposition(contentDisposition string) string {
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return ""
	}
	return params["filename"]
}

func DownloadFile(requiredFileName string, client *HeaderHttpClient, c *model.UContext, logger *logrus.Entry) {
	logger.Trace("into DownloadFile")
	logger.Tracef("fileId: %s", requiredFileName)
	downloadUrl := model.UrlDownloadApi(c, requiredFileName)
	logger.Debugf("file url: %s", downloadUrl)
	res, err := client.Get(downloadUrl)
	if err != nil {
		logger.Fatalf("download file error: %v", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		// print response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Fatalf("read response body error when response stats code is not OK: %v", err)
		}
		logger.Tracef("response body: %s", body)
		msg, err := ExtractErrorMsg(body)
		if err != nil {
			logger.Fatalf("extract error message error: %v", err)
		}
		logger.Fatalf("download file failed: %v, error msg: %v", res.Status, msg)
	}
	contentDisposition := res.Header.Get("Content-Disposition")
	logger.Tracef("Content-Disposition: %s", contentDisposition)
	realFileName := parseContentDisposition(contentDisposition)
	logger.Debugf("real file name: %s", realFileName)
	file, err := os.Create(realFileName)
	if err != nil {
		logger.Fatalf("create file error: %v", err)
	}
	defer file.Close()
	N, err := io.Copy(file, res.Body)
	if err != nil {
		logger.Fatalf("copy file error: %v", err)
	}
	fmt.Printf("download file success: %s, file_size: %vKib\n", realFileName, float64(N/1024))

}
