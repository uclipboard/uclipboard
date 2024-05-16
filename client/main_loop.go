package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/model"
)

func mainLoop(conf *model.Conf, adapter model.ClipboardCmdAdapter, client *http.Client) {
	logger := model.NewModuleLogger("loop")
	var previousClipboard model.Clipboard
	for {
		time.Sleep(time.Duration(conf.Client.Interval) * time.Millisecond) //sleep first to avoid the possible occured error then it skip sleep
		currentClipboard, E := adapter.Paste()
		if E != nil {
			logger.Fatalf("adapter.Paste error:%v", E)
		}
		logger.Tracef("adapter.Paste %v", []byte(currentClipboard))

		// It's not a good idea to use PullStringData
		// because I need the error infomation to skip current look
		pullApi := model.UrlPullApi(conf)
		resp, err := client.Get(pullApi)
		if err != nil {
			logger.Warnf("error sending request: %s", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Fatalf("error reading response body: %s", err)
		}

		var remoteClipboards []model.Clipboard
		logger.Tracef("response body: %s", string(body))
		if err := json.Unmarshal(body, &remoteClipboards); err != nil {
			logger.Fatalf("cannot parse response body: %s", err.Error())
		}

		previousClipboardHistoryidx := model.IndexClipboardArray(remoteClipboards, &previousClipboard)
		clipboardContentIfIsFile := deteckAndconcatClipboardFileUrl(conf, &remoteClipboards[0])
		// Now we just ignore the conflict when all of those are different
		// Why do we need `previousClipboardHistoryidx`?
		// Consider the following situation
		// Copy text1 (remote whatever)
		// Copy text2 (remote text1)
		// Copy text1 (remote text2)
		// If there was not previousClipboard we created,
		// and just check the index of currentClipboard in remoteClipboard.
		// text2 will be synchronized to adapter again,
		// but what we need is a new copied text1
		// even though text1 is same as previous text1
		if previousClipboardHistoryidx > 0 {
			logger.Infof("Pull <= %v [%v]", remoteClipboards[0].Content, remoteClipboards[0].Hostname)
			previousClipboard = remoteClipboards[0]
			E := adapter.Copy(clipboardContentIfIsFile)
			if E != nil {
				logger.Fatalf("adapter.Copy error:%v", E)

			}

		} else if previousClipboard.Content != currentClipboard && previousClipboardHistoryidx == 0 {
			logger.Tracef("previousClipboard.Content is %v\n", []byte(previousClipboard.Content))
			logger.Tracef("s is %v\n", []byte(currentClipboard))
			logger.Infof("Push => %s", currentClipboard)
			// It's not good idea to use UploadStringData function because I need wrappedClipboard
			reqBody, wrappedClipboard := genClipboardReqBody(currentClipboard, logger)
			// update Clipboard
			logger.Tracef("previousClipboard=wrappedClipboard: %v", wrappedClipboard)
			previousClipboard = *wrappedClipboard

			resp, err := client.Post(model.UrlPushApi(conf),
				"application/json", bytes.NewReader(reqBody))

			if err != nil {
				logger.Fatalf("push clipboard error: %s", err.Error())
			}
			resp.Body.Close()

		} else if previousClipboardHistoryidx == -1 {
			previousClipboard = remoteClipboards[0]
			logger.Info("This is a new client, synchronizing from server...")
			E := adapter.Copy(clipboardContentIfIsFile)

			if E != nil {
				logger.Fatalf("adapter.Copy error:%v", E)

			}
			logger.Info("synchronize data completed.")
			logger.Tracef("previousClipboard.Content: %s", previousClipboard.Content)

		} else {
			logger.Debug("current clipboard is up-to-date")
		}

	}

}
