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
	var previousClipboard model.Clipboard
	for {
		time.Sleep(time.Duration(cfg.Client.Interval) * time.Millisecond) //sleep first to avoid the possible occured error then it skip sleep
		s, _ := adapter.Paste()
		fmt.Printf("adapter.Paste %v\n", []byte(s))

		pullApi := model.UrlPullApi(cfg)
		resp, err := client.Get(pullApi)
		if err != nil {
			fmt.Println("err sending req:", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		var remoteClipboards []model.Clipboard
		println(string(body))
		if err = json.Unmarshal(body, &remoteClipboards); err != nil {
			panic(err)
		}

		resp.Body.Close() //this function may not be stop,so I'd better close it by myself

		previousClipboardHistoryidx := model.IndexClipboardArray(remoteClipboards, &previousClipboard)
		// now we have previousClipboard  ,remoteClipboards and current clipboard s

		// in current,we just ignore the conflict when all of those are different

		if previousClipboard.Content == s && previousClipboardHistoryidx > 0 {
			println("pull from server")
			previousClipboard = remoteClipboards[0]
			adapter.Copy(previousClipboard.Content)
		} else if previousClipboard.Content != s && previousClipboardHistoryidx == 0 {
			fmt.Printf("previousClipboard.Content %v\n", []byte(previousClipboard.Content))
			fmt.Printf("s %v\n", []byte(s))
			fmt.Printf("%v\n", previousClipboard.Content == s)

			println("push clipboard to server and update local previousClipboard")
			reqBody, wrappedClipboard := GenClipboardReqBody(s)
			// update current Clipboard
			previousClipboard = *wrappedClipboard
			resp, err = client.Post(model.UrlPushApi(cfg),
				"application/json", bytes.NewReader(reqBody))

			if err != nil {
				panic(err)
			}
			resp.Body.Close()

		} else if previousClipboardHistoryidx == -1 {
			println("a new client,pull from server")
			previousClipboard = remoteClipboards[0]
			adapter.Copy(previousClipboard.Content)
			println(previousClipboard.Content)

		} else {
			fmt.Println("latest clipboard")
		}

	}

}
