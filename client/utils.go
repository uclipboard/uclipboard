package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

func GenClipboardReqBody(c string) ([]byte, *model.Clipboard) {
	reqData := model.NewClipoardWithDefault()
	reqData.Content = c
	hostname, err := os.Hostname()
	if err == nil {
		// TODO:log this
		reqData.Hostname = hostname
	}

	reqData.Ts = time.Now().Unix()
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		logger.Panicf("parse response jsonbody error: %s", err.Error())
	}
	return reqBody, reqData
}

func UploadStringData(s string, client *http.Client, c *model.Conf) {
	reqBody, _ := GenClipboardReqBody(s)
	resp, err := client.Post(model.UrlPushApi(c),
		"application/json", bytes.NewReader(reqBody))

	if err != nil {
		logger.Panic("upload")
	}
	defer resp.Body.Close()
}
