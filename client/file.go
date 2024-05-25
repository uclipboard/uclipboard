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

	"github.com/dangjinghao/uclipboard/model"
	"github.com/sirupsen/logrus"
)

func UploadFile(filePath string, client *http.Client, c *model.Conf, logger *logrus.Entry) {
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
	if resp.StatusCode != http.StatusOK {
		logger.Fatalf("Upload file failed, status: %s", resp.Status)
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

	fmt.Printf("upload file success, file_id: %v, file_name: %v, life_time: %vs\n",
		respData["file_id"], respData["file_name"], respData["life_time"])

}

func parseContentDisposition(contentDisposition string) string {
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return ""
	}
	return params["filename"]
}

func DownloadFile(fileId string, client *http.Client, c *model.Conf, logger *logrus.Entry) {
	logger.Trace("into DownloadFile")
	logger.Tracef("fileId: %s", fileId)
	downloadUrl := model.UrlDownloadApi(c, fileId)
	logger.Tracef("downloadApi: %s", downloadUrl)
	res, err := client.Get(downloadUrl)
	if err != nil {
		logger.Fatalf("download file error: %v", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		fmt.Printf("file not found!\n")
		return
	} else if res.StatusCode != http.StatusOK {
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
		logger.Fatalf("download file failed, status_code %d, msg: %v", res.StatusCode, msg)
	}
	contentDisposition := res.Header.Get("Content-Disposition")
	logger.Tracef("Content-Disposition: %s", contentDisposition)
	fileName := parseContentDisposition(contentDisposition)
	file, err := os.Create(fileName)
	if err != nil {
		logger.Fatalf("create file error: %v", err)
	}
	defer file.Close()
	N, err := io.Copy(file, res.Body)
	if err != nil {
		logger.Fatalf("copy file error: %v", err)
	}
	fmt.Printf("download file success: %s, file_size: %vKib\n", fileName, float32(N/1024))

}
