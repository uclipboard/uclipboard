package client

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uclipboard/uclipboard/client/adapter"
	"github.com/uclipboard/uclipboard/model"
)

type loopContenxt struct {
	uctx                 *model.UContext
	client               *http.Client
	adapter              adapter.ClipboardCmdAdapter
	logger               *logrus.Entry
	dynamicSleepTime     time.Duration
	lockedWarningCounter int
	previousClipboard    model.Clipboard //skip init the memory
}

func (ctx *loopContenxt) stagePull() ([]byte, bool) {
	logger := ctx.logger
	logger.Trace("into loopPullStage")

	body, pullErr := SendPullReq(ctx.client, ctx.uctx)
	if pullErr != nil {
		if pullErr == ErrUnexpectRespStatus {
			logger.Tracef("the response body when received unexpected response status occured: %q", string(body))
			serverMsg, extractErr := ExtractErrorMsg(body)
			if extractErr != nil {
				logger.Warnf("extract error msg error: %v", extractErr)
				return nil, false
			}
			if serverMsg == "" {
				logger.Warn("server msg is empty, please check the server log for more information")
				return nil, false
			}
			logger.Warnf("receive unexpceted server msg: %s", serverMsg)
		} else {
			logger.Warnf("send pull request error: %v", pullErr)
			// when my laptop wakes up from sleep, the connection would never be activated.
			// so I try to close ide connections and recreate a client
			logger.Warn("reset client because of connection error")
			logger.Debug("close idle connections")
			ctx.client.CloseIdleConnections()
			// it looks like that close idle connections is enough,
			// but maybe it is better to recreate a new client
			logger.Debug("create a new client")
			ctx.client = NewUClipboardHttpClient(ctx.uctx)

			if ctx.dynamicSleepTime <= 60*time.Second {
				ctx.dynamicSleepTime *= 2
				logger.Debugf("increase sleep time to %v", ctx.dynamicSleepTime)
			} else {
				logger.Debugf("interval time has been arrived the maxmium value:%v", ctx.dynamicSleepTime)
			}
		}

		return nil, false
	}

	if ctx.dynamicSleepTime != time.Duration(ctx.uctx.Client.Connect.Interval)*time.Millisecond {
		ctx.logger.Debug("reset sleep time because the connection is activated.")
		ctx.dynamicSleepTime = time.Duration(ctx.uctx.Client.Connect.Interval) * time.Millisecond
	}

	return body, true
}

func (ctx *loopContenxt) stageCopy(currentClipboard *model.Clipboard) bool {
	s := DetectAndConcatFileUrl(ctx.uctx, currentClipboard)
	ctx.logger.Tracef("(MOD |V|) %s => %s", currentClipboard.Content, s)
	E := ctx.adapter.Copy(s)
	if E != nil {
		ctx.logger.Warnf("adapter.Copy error: %v", E)
		return false
	}
	return true
}

// ret:
// 0: previous clipboard is up-to-date,should check/upload the system clipboard,
// others: previous clipboard is updated, should update the system clipboard
func (ctx *loopContenxt) stageRemoteDecision(remoteClipboards []model.Clipboard) int {
	previousClipboardHistoryidx := model.IndexClipboardArray(remoteClipboards, &ctx.previousClipboard)
	ctx.logger.Tracef("previousClipboard: %v", ctx.previousClipboard)
	if previousClipboardHistoryidx == -1 {
		ctx.logger.Debug("update previousClipboard as a new client")
		ctx.previousClipboard = remoteClipboards[0]
		return -1

	} else if previousClipboardHistoryidx > 0 {
		ctx.logger.Debug("update previousClipboard")
		ctx.previousClipboard = remoteClipboards[0]
		return 1
	}

	// else: previousClipboardHistoryidx == 0, detect whether the current clipboard is updated.
	return 0
}

func (ctx *loopContenxt) stagePaste() (string, bool) {
	ctx.logger.Trace("into stagePaste")
	currentClipboard, E := ctx.adapter.Paste()
	if E != nil {
		if E == adapter.ErrEmptyClipboard {
			ctx.logger.Debugf(`adapter.Paste error:%v, set empty string clipboard.`, E)
			currentClipboard = ""
		} else if E == adapter.ErrLockedClipboard {
			if ctx.lockedWarningCounter < 3 {
				ctx.logger.Info("clipboard is locked, skip push")
				ctx.lockedWarningCounter++
			} else {
				ctx.logger.Debugf("clipboard is locked, skip push, but the warning counter has reached the maximum value: %v", ctx.lockedWarningCounter)
			}
			return "", false
		} else {
			ctx.logger.Warnf("adapter.Paste error:%v", E)
			return "", false
		}

	}

	if ctx.lockedWarningCounter > 0 {
		ctx.logger.Info("clipboard is unlocked, reset the warning counter")
		ctx.lockedWarningCounter = 0
	}

	if currentClipboard == "" {
		ctx.logger.Debug("skip push detect because current clipboard is empty")
		return "", false
	}

	if len(currentClipboard) > ctx.uctx.ContentLengthLimit {
		ctx.logger.Debug("current clipboard size is too large, skip push")
		return "", false
	}

	return currentClipboard, true
}

func (ctx *loopContenxt) stageLocalDecision(currentClipboardContent string) (doPush bool) {
	ctx.logger.Trace("into stageLocalDecision")
	ctx.logger.Tracef("previousClipboard.Content %q[%v]\n", ctx.previousClipboard.Content, []byte(ctx.previousClipboard.Content))
	ctx.logger.Tracef("currentClipboard %q[%v]\n", currentClipboardContent, []byte(currentClipboardContent))
	if ctx.previousClipboard.Content != currentClipboardContent {
		clipboardContentIfIsFile := DetectAndConcatFileUrl(ctx.uctx, &ctx.previousClipboard)
		if currentClipboardContent == clipboardContentIfIsFile {
			return false
		}
		return true
	}
	return false
}

func (ctx *loopContenxt) stagePush(currentClipboard string) bool {
	ctx.logger.Debugf("(PUSH =>) %s", currentClipboard)
	var wrappedCLipboard *model.Clipboard
	var err error
	if wrappedCLipboard, err = SendPushReq(currentClipboard, ctx.client, ctx.uctx); err != nil {
		ctx.logger.Warnf("send push request error: %v", err)
		return false
	}
	ctx.previousClipboard = *wrappedCLipboard
	return true
}

func mainLoop(conf *model.UContext, theAdapter adapter.ClipboardCmdAdapter, client *http.Client) {
	logger := model.NewModuleLogger("loop")
	logger.Tracef("into mainLoop")

	dynamicSleepTime := time.Duration(conf.Client.Connect.Interval) * time.Millisecond
	logger.Debugf("default sleep time: %v", dynamicSleepTime)
	ctx := loopContenxt{
		uctx:             conf,
		client:           client,
		adapter:          theAdapter,
		logger:           logger,
		dynamicSleepTime: dynamicSleepTime,
	}

	for {
		logger.Tracef("sleep %v begin", ctx.dynamicSleepTime)
		time.Sleep(ctx.dynamicSleepTime) //sleep first to avoid the possible occured error make the loop skips sleep
		logger.Trace("sleep end")
		body, ok := ctx.stagePull()
		if !ok {
			continue
		}

		logger.Tracef("current response body: %s", string(body))
		remoteClipboards, err := ParsePullData(body)
		if err != nil {
			logger.Warnf("error parsing response body: %v", err)
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

		previousClipboardHistoryidx := ctx.stageRemoteDecision(remoteClipboards)
		switch previousClipboardHistoryidx {
		case -1:
			ctx.logger.Info("This is a new client, synchronizing from server...")
			if ok = ctx.stageCopy(&ctx.previousClipboard); !ok {
				continue
			}
			ctx.logger.Info("synchronize data completed.")
			continue
		case 1:
			ctx.logger.Debugf("(PULL <=) %q [%v]", ctx.previousClipboard.Content, ctx.previousClipboard.Hostname)
			if ok = ctx.stageCopy(&ctx.previousClipboard); !ok {
				continue
			}
			ctx.logger.Debug("pull data completed.")
			continue
		}
		// case 0: check system clipboard and maybe push

		currentClipboard, ok := ctx.stagePaste()
		if !ok {
			continue
		}

		if doPush := ctx.stageLocalDecision(currentClipboard); doPush {
			if ok = ctx.stagePush(currentClipboard); !ok {
				continue
			}
		}
		logger.Debug("current clipboard is up-to-date")

	}

}
