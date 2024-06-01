package client

import (
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func mainLoop(conf *model.Conf, adapterObj adapter.ClipboardCmdAdapter, client *http.Client) {
	logger := model.NewModuleLogger("loop")
	logger.Tracef("into mainLoop")
	var previousClipboard model.Clipboard
	dynamicSleepTime := time.Duration(conf.Client.Interval) * time.Millisecond
	logger.Debugf("default sleep time: %v", dynamicSleepTime)
	for {
		logger.Tracef("sleep %v begin", dynamicSleepTime)
		time.Sleep(dynamicSleepTime) //sleep first to avoid the possible occured error then it skip sleep
		logger.Trace("sleep end")
		body, err := SendPullReq(client, conf)
		if err != nil {
			if err == ErrUnexpectRespStatus {
				logger.Tracef("the response body when ErrUnexpectRespStatus occured: %q", string(body))
				serverMsg, err := ExtractErrorMsg(body)
				if err != nil {
					logger.Warnf("extrace error msg error: %v", err)
					continue
				}
				if serverMsg == "" {
					logger.Warn("server msg is empty, please check the server log for more information")
					continue
				}
				logger.Warnf("receive server msg: %s", serverMsg)

			} else {
				logger.Warnf("send pull request error: %v", err)
				if dynamicSleepTime <= 60*time.Second {
					dynamicSleepTime *= 2
					logger.Debugf("increase sleep time to %v", dynamicSleepTime)
				} else {
					logger.Debugf("interval time has been arrived the maxmium value:%v", dynamicSleepTime)
				}
			}

			continue
		}

		logger.Tracef("current response body: %q", string(body))
		if dynamicSleepTime != time.Duration(conf.Client.Interval)*time.Millisecond {
			logger.Debug("reset sleep time because the connection is activated.")
			dynamicSleepTime = time.Duration(conf.Client.Interval) * time.Millisecond
		}
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
		clipboardContentIfIsFile := DetectAndConcatFileUrl(conf, &remoteClipboards[0])

		if previousClipboardHistoryidx == -1 {
			previousClipboard = remoteClipboards[0]
			logger.Info("This is a new client, synchronizing from server...")
			E := adapterObj.Copy(clipboardContentIfIsFile)

			if E != nil {
				logger.Warnf("adapter.Copy error:%v", E)
				continue
			}
			logger.Info("synchronize data completed.")
			logger.Tracef("previousClipboard.Content: %s", previousClipboard.Content)
			continue
		} else if previousClipboardHistoryidx > 0 {
			logger.Debugf("(PULL <=) %v [%v]", remoteClipboards[0].Content, remoteClipboards[0].Hostname)
			previousClipboard = remoteClipboards[0]
			E := adapterObj.Copy(clipboardContentIfIsFile)
			if E != nil {
				logger.Warnf("adapter.Copy error:%v", E)
				continue
			}
			continue
		}
		// else: previousClipboardHistoryidx == 0, detect whether the current clipboard is updated.
		currentClipboard, E := adapterObj.Paste()
		if E != nil {
			if E == adapter.ErrEmptyClipboard {
				logger.Debugf(`adapter.Paste error:%v ,set empty string clipboard.`, E)
				currentClipboard = ""
			} else {
				logger.Warnf("adapter.Paste error:%v", E)
				continue
			}

		}

		if currentClipboard == "" {
			logger.Debug("skip current loop because current clipboard is empty")
			continue
		}

		logger.Tracef("previousClipboard.Content %s[%v]\n", previousClipboard.Content, []byte(previousClipboard.Content))
		logger.Tracef("currentClipboard %s[%v]\n", currentClipboard, []byte(currentClipboard))
		if previousClipboard.Content != currentClipboard {
			if currentClipboard == clipboardContentIfIsFile {
				logger.Debug("currentClipboard is file url clipboard, skip push")
				continue
			}
			logger.Debugf("(PUSH =>) %s", currentClipboard)
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
