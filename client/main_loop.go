package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/model"
)

func mainLoop(cfg *model.Conf, adapter model.ClipboardCmdAdapter, client *http.Client) {
	logger := model.NewModuleLogger("loop")
	var previousClipboard model.Clipboard
	for {
		time.Sleep(time.Duration(cfg.Client.Interval) * time.Millisecond) //sleep first to avoid the possible occured error then it skip sleep
		s, E := adapter.Paste()
		if E != nil {
			logger.Panicf("adapter.Paste error:%v", E)
		}
		logger.Tracef("adapter.Paste %v", []byte(s))

		// It's not a good idea to use PullStringData
		// because I need the error infomation to skip current look
		pullApi := model.UrlPullApi(cfg)
		resp, err := client.Get(pullApi)
		if err != nil {
			logger.Warnf("err sending req: %s", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Panicf("Error reading response body: %s", err)
		}

		var remoteClipboards []model.Clipboard
		logger.Tracef("response body: %s", string(body))
		if err := json.Unmarshal(body, &remoteClipboards); err != nil {
			logger.Panicf("cannot parse response body: %s", err.Error())
		}

		previousClipboardHistoryidx := model.IndexClipboardArray(remoteClipboards, &previousClipboard)
		// now we have previousClipboard, remoteClipboards and current clipboard s
		//  TODO:in current,we just ignore the conflict when all of those are different
		// why we need `previousClipboardHistoryidx`?
		// In fulture websocket connection mode, `previousClipboardHistoryidx` stores server pushed from remote server
		// And it will be used to sync remote data.
		if previousClipboardHistoryidx > 0 {
			logger.Debugf("Pull from server: %v[%v]", remoteClipboards[0].Content, remoteClipboards[0].Hostname)
			previousClipboard = remoteClipboards[0]
			E := adapter.Copy(previousClipboard.Content)
			if E != nil {
				logger.Panicf("adapter.Copy error:%v", E)

			}

		} else if previousClipboard.Content != s && previousClipboardHistoryidx == 0 {
			logger.Tracef("previousClipboard.Content is %v\n", []byte(previousClipboard.Content))
			logger.Tracef("s is %v\n", []byte(s))
			logger.Debug("Push clipboard to server and update local previousClipboard")
			// It's not good idea to use UploadStringData function because I need wrappedClipboard
			reqBody, wrappedClipboard := GenClipboardReqBody(s, logger)
			// update current Clipboard
			logger.Tracef("previousClipboard=wrappedClipboard: %v", wrappedClipboard)
			previousClipboard = *wrappedClipboard

			resp, err := client.Post(model.UrlPushApi(cfg),
				"application/json", bytes.NewReader(reqBody))

			if err != nil {
				logger.Panicf("push clipboard error: %s", err.Error())
			}
			resp.Body.Close()

		} else if previousClipboardHistoryidx == -1 {
			logger.Info("This is a new client, pulling from server...")
			previousClipboard = remoteClipboards[0]
			E := adapter.Copy(previousClipboard.Content)
			if E != nil {
				logger.Panicf("adapter.Copy error:%v", E)

			}
			logger.Tracef("previousClipboard.Content: %s", previousClipboard.Content)

		} else {
			logger.Debug("Current clipboard is up-to-date")
		}

	}

}
