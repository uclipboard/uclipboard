package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/sirupsen/logrus"
)

func mainLoop(cfg *model.Conf, adapter model.ClipboardCmdAdapter, client *http.Client, logger *logrus.Entry) {
	loopLogger := logger.WithField("section", "loop")
	var previousClipboard model.Clipboard
	for {
		time.Sleep(time.Duration(cfg.Client.Interval) * time.Millisecond) //sleep first to avoid the possible occured error then it skip sleep
		s, _ := adapter.Paste()
		loopLogger.Tracef("adapter.Paste %v", []byte(s))

		pullApi := model.UrlPullApi(cfg)
		resp, err := client.Get(pullApi)
		if err != nil {
			loopLogger.Warnf("err sending req: %s", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			loopLogger.Panicf("Error reading response body: %s", err)
		}

		var remoteClipboards []model.Clipboard
		loopLogger.Tracef("response body: %s", string(body))
		if err = json.Unmarshal(body, &remoteClipboards); err != nil {
			loopLogger.Panicf("cannot parse response body: %s", err.Error())
		}

		resp.Body.Close() //this function don't stop,so I'd better close it by myself

		previousClipboardHistoryidx := model.IndexClipboardArray(remoteClipboards, &previousClipboard)
		// now we have previousClipboard, remoteClipboards and current clipboard s
		//  TODO:in current,we just ignore the conflict when all of those are different

		if previousClipboard.Content == s && previousClipboardHistoryidx > 0 {
			loopLogger.Debugf("Pull from server: %v", remoteClipboards[0])
			previousClipboard = remoteClipboards[0]
			adapter.Copy(previousClipboard.Content)

		} else if previousClipboard.Content != s && previousClipboardHistoryidx == 0 {
			loopLogger.Tracef("previousClipboard.Content is %v\n", []byte(previousClipboard.Content))
			loopLogger.Tracef("s is %v\n", []byte(s))
			loopLogger.Debug("Push clipboard to server and update local previousClipboard")
			reqBody, wrappedClipboard := GenClipboardReqBody(s)
			// update current Clipboard
			loopLogger.Tracef("previousClipboard=wrappedClipboard: %v", wrappedClipboard)
			previousClipboard = *wrappedClipboard

			resp, err = client.Post(model.UrlPushApi(cfg),
				"application/json", bytes.NewReader(reqBody))

			if err != nil {
				loopLogger.Panicf("push clipboard error: %s", err.Error())
			}
			resp.Body.Close()

		} else if previousClipboardHistoryidx == -1 {
			loopLogger.Info("This is a new client, pulling from server...")
			previousClipboard = remoteClipboards[0]
			adapter.Copy(previousClipboard.Content)
			loopLogger.Tracef("previousClipboard.Content: %s", previousClipboard.Content)

		} else {
			loopLogger.Debug("Current clipboard is up-to-date")
		}

	}

}
