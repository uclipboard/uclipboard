package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/sirupsen/logrus"
)

func GenClipboardReqBody(c string, logger *logrus.Entry) ([]byte, *model.Clipboard) {
	reqData := model.NewClipoardWithDefault()
	reqData.Content = c
	hostname, err := os.Hostname()
	if err == nil {
		logger.Warnf("Can't get hostname:%v", err)
		reqData.Hostname = hostname
	}

	reqData.Ts = time.Now().Unix()
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		logger.Panicf("parse response json body error: %s", err.Error())

	}
	return reqBody, reqData
}

func UploadStringData(s string, client *http.Client, c *model.Conf, logger *logrus.Entry) {
	reqBody, _ := GenClipboardReqBody(s, logger)
	resp, err := client.Post(model.UrlPushApi(c),
		"application/json", bytes.NewReader(reqBody))

	if err != nil {
		logger.Error()
		return
	}
	defer resp.Body.Close()
}

func PullStringData(client *http.Client, c *model.Conf, logger *logrus.Entry) ([]byte, error) {
	pullApi := model.UrlPullApi(c)
	resp, err := client.Get(pullApi)
	if err != nil {
		logger.Warnf("Error sending req: %s", err)
		return nil, model.ErrRec
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("Error reading response body: %s", err)
		return nil, model.ErrRec
	}

	resp.Body.Close()

	return body, nil
}
