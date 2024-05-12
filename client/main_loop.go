package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/model"
)

func mainLoop(cfg *model.Conf, adapter model.ClipboardCmdAdapter, client *http.Client) {
	for {
		s, _ := adapter.Paste()
		pullApi := model.UrlPullApi(cfg)
		resp, err := client.Get(pullApi)
		if err != nil {
			fmt.Println("err sending req:", err)
			continue
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		var RemoteClipboard model.Clipboard
		if err = json.Unmarshal(body, &RemoteClipboard); err != nil {
			panic(err)
		}
		resp.Body.Close() //this function may not be stop,so I'd better close it by myself
		if RemoteClipboard.Content != s {
			reqBody := GenClipboardReqBody(s)
			resp, err := client.Post(model.UrlPushApi(cfg),
				"application/json", bytes.NewReader(reqBody))

			if err != nil {
				panic(err)
			}

			resp.Body.Close()
		}
		time.Sleep(time.Duration(cfg.Client.Interval) * time.Millisecond)
	}

}
