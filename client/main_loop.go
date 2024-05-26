package client

import (
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func mainLoop(conf *model.Conf, adapterObj adapter.ClipboardCmdAdapter, client *http.Client) {
	logger := model.NewModuleLogger("loop")
	var previousClipboard model.Clipboard
	for {
		time.Sleep(time.Duration(conf.Client.Interval) * time.Millisecond) //sleep first to avoid the possible occured error then it skip sleep

		body, err := SendPullReq(client, conf)
		if err != nil {
			logger.Warnf("send pull request error: %v", err)
			if err == ErrUnexpectRespStatus {
				serverMsg, err := ExtractErrorMsg(body)
				if err != nil {
					logger.Warnf("extrace error msg error: %s", serverMsg)
					continue
				}
				logger.Warnf("receive server msg: %s", serverMsg)
			}
			continue
		}

		logger.Tracef("response body: %s", string(body))

		remoteClipboards, err := ParsePullData(body)
		if err != nil {
			logger.Warnf("error parsing response body: %s", err)
			continue
		}
		logger.Tracef("remoteClipboards: %v", remoteClipboards)
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

		previousClipboardHistoryidx := model.IndexClipboardArray(remoteClipboards, &previousClipboard)
		clipboardContentIfIsFile := DeteckAndConcatFileUrl(conf, &remoteClipboards[0])

		if previousClipboardHistoryidx == -1 {
			previousClipboard = remoteClipboards[0]
			logger.Info("This is a new client, synchronizing from server...")
			E := adapterObj.Copy(clipboardContentIfIsFile)

			if E != nil {
				logger.Fatalf("adapter.Copy error:%v", E)

			}
			logger.Info("synchronize data completed.")
			logger.Tracef("previousClipboard.Content: %s", previousClipboard.Content)
			continue
		} else if previousClipboardHistoryidx > 0 {
			logger.Infof("(PULL <=) %v [%v]", remoteClipboards[0].Content, remoteClipboards[0].Hostname)
			previousClipboard = remoteClipboards[0]
			E := adapterObj.Copy(clipboardContentIfIsFile)
			if E != nil {
				logger.Fatalf("adapter.Copy error:%v", E)

			}
			continue
		}
		// else: previousClipboardHistoryidx == 0, detect whether the current clipboard is updated.
		currentClipboard, E := adapterObj.Paste()
		if E != nil {
			if E == adapter.ErrEmptyClipboard {
				logger.Infof(`adapter.Paste error:%v ,set empty string clipboard.`, E)
				currentClipboard = ""
			} else {
				logger.Fatalf("adapter.Paste error:%v", E)
			}

		}
		logger.Tracef("adapter.Paste %v", []byte(currentClipboard))

		if currentClipboard == "" {
			logger.Info("skip current loop because current clipboard is empty")
			continue
		}

		logger.Tracef("previousClipboard.Content bytes is %v\n", []byte(previousClipboard.Content))
		logger.Tracef("currentClipboard bytes is %v\n", []byte(currentClipboard))
		if previousClipboard.Content != currentClipboard {
			if currentClipboard == clipboardContentIfIsFile {
				logger.Debug("currentClipboard is file url clipboard, skip push")
				continue
			}
			logger.Infof("(PUSH =>) %s", currentClipboard)
			var wrappedCLipboard *model.Clipboard
			if wrappedCLipboard, err = SendPushReq(currentClipboard, client, conf); err != nil {
				logger.Warnf("send push request error: %v", err)
				continue
			}
			previousClipboard = *wrappedCLipboard
			continue
		}

		logger.Debug("current clipboard is up-to-date")

	}

}
