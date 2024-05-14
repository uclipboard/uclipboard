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
	if err != nil {
		logger.Warnf("Can't get hostname:%v", err)
	} else {
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
	logger.Trace("into UploadStringData")
	reqBody, _ := GenClipboardReqBody(s, logger)
	resp, err := client.Post(model.UrlPushApi(c),
		"application/json", bytes.NewReader(reqBody))

	if err != nil {
		logger.Warnf("upload string data failed:%s", err.Error())
		return
	}
	defer resp.Body.Close()
}

func PullStringData(client *http.Client, c *model.Conf, logger *logrus.Entry) ([]byte, error) {
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
