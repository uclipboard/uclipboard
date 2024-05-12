package client

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dangjinghao/uclipboard/model"
)

func GenClipboardReqBody(c string) []byte {
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
		panic(err)
	}
	return reqBody
}
