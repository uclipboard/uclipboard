package client

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/uclipboard/uclipboard/client/adapter"
	"github.com/uclipboard/uclipboard/model"
)

// because there are 2 or more goroutines to access the clipboard in persist mode,
// we need to use a mutex to lock the clipboard
// to protect the clipboard
// after architecture update, it is not necessary to use a mutex
type clipboardLock struct {
	lock    sync.Mutex
	adapter adapter.ClipboardCmdAdapter
}

func (c *clipboardLock) copy(content string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := c.adapter.Copy(content)
	return err
}

func (c *clipboardLock) paste() (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	s, err := c.adapter.Paste()
	return s, err

}

func (c *clipboardLock) watch(onChange func(string)) error {
	return c.adapter.Watch(onChange)
}

func newClipboardLock(adapter adapter.ClipboardCmdAdapter) *clipboardLock {
	return &clipboardLock{
		adapter: adapter,
	}
}

func clipboardLocalChangeWatchService(cl *clipboardLock, notify chan string) {
	logger := model.NewModuleLogger("clipboardLocalChangeWatchService")
	// this is a blocking call
	// when the clipboard content changed, it will call the function
	err := cl.watch(func(newContent string) {
		logger.Debug("Clipboard content changed, notifying")
		// send the content to the channel
		notify <- newContent
	})
	// normally this will never happen
	if err != nil {
		logger.Fatalf("watch error: %v", err)
	}

}

func clipboardLocalChangeService(u *model.UContext, cl *clipboardLock, wso *model.WsObject, loopUpdateNotify chan string) {
	logger := model.NewModuleLogger("clipboardLocalChangeWatcher")
	// create a updateNotify channel
	updateNotify := make(chan string, 20)
	go clipboardLocalChangeWatchService(cl, updateNotify)
	for {
		func() {
			// wait for the clipboard content changed
			// this will block until the clipboard content changed
			updateCurrent := <-updateNotify
			defer func() {
				logger.Debugf("set prevContext to currentClipboard: %s", updateCurrent)
				u.Runtime.ClipboardCurrentContent = updateCurrent
			}()
			// Consume the loopUpdateNotify channel is to avoid the local change worker
			// detects the same clipboard content change which is proactively pushed by the server.
			// Consume the loopUpdateNotify channel until it is empty or the message is the same as updateCurrent,
			// this is to avoid the local change worker lose some clipboard content update but the server has already proactively pushed
		consumeLoop:
			for {
				select {
				case notifyMsg := <-loopUpdateNotify:
					if notifyMsg == updateCurrent {
						logger.Debugf("same clipboard notify: %s, skip local change", notifyMsg)
						return
					}
					logger.Warnf("unexpected different notify message received: %s, keep consuming...", notifyMsg)
				default:
					logger.Debug("no more messages to consume, stopping consume loop")
					break consumeLoop
				}
			}
			// it is not recommended to put this before the select statement
			// if we skip it, we can't consume the message
			// and that may cause the channel to be full
			if updateCurrent == u.Runtime.ClipboardCurrentContent {
				logger.Debug("current clipboard is same as previous, skip push")
				return
			}

			logger.Debugf("Receive clipboard content changed notify: %v", updateCurrent)
			if updateCurrent == "" {
				logger.Debug("skip push detect because current clipboard is empty")
				return
			} else if len(updateCurrent) > u.ContentLengthLimit {
				logger.Debug("current clipboard size is too large, skip push")
				updateCurrent = updateCurrent[:u.ContentLengthLimit]
				return
			}

			if _, err := SendWebSocketPush(updateCurrent, wso); err != nil {
				logger.Errorf("send websocket push error: %v", err)
				return
			}
			logger.Debugf("send websocket push success")
		}()
	}

}

func persistMainLoop(conf *model.UContext, theAdapter adapter.ClipboardCmdAdapter, _ *HeaderHttpClient) {
	logger := model.NewModuleLogger("persist")
	logger.Tracef("into persist mainLoop")
	wso, err := CreateWsConn(conf)
	if err != nil {
		logger.Errorf("create ws connection error: %v, try to re-connect and init again", err)
		// we can't use the wso.ClientErrorHandle here
		// because we don't have a wso yet
		model.MaxLimitExpoGrowthAlgo(logger, model.DefaultInitReconDelay, model.DefaultMaxReconnDelay, func() bool {
			wso, err = CreateWsConn(conf)
			if err != nil {
				logger.Errorf("re-create ws connection error: %v", err)
				return false
			}
			logger.Debug("re-create ws connection success")
			return true
		})
	}

	logger.Info("Successfully connected to server.")

	defer wso.Close()
	timeout := time.Duration(conf.Client.Connect.Timeout) * time.Millisecond
	wso.InitClientPingHandler(timeout)
	cl := newClipboardLock(theAdapter)

	loopUpdateNotify := make(chan string, 20)

	pasteContent, err := cl.paste()
	if err != nil {
		logger.Errorf("get clipboard content error at startup, skip this operation: %v", err)
	}
	logger.Debugf("set clipboard content buffer at startup: %v", pasteContent)
	conf.Runtime.ClipboardCurrentContent = pasteContent

	// send the pull request to the server
	// then we can get the latest clipboard content in the
	// case model.WSMsgTypeData
	if err := SendWebSocketPull(wso); err != nil {
		if err = wso.ClientErrorHandle(err); err != nil {
			logger.Errorf("read message error: %v", err)
			return
		}
	}
	initWorkers := true
	for {
		msgType, content, err := wso.ReadMessage()
		if err != nil {
			if err = wso.ClientErrorHandle(err); err != nil {
				logger.Errorf("read message error: %v", err)
				continue
			}
			logger.Debug("send pull request to sync the data after reconnect")
			if err := SendWebSocketPull(wso); err != nil {
				logger.Errorf("Failed to send pull request after reconnect: %v", err)
				continue
			}

			msgType, content, err = wso.ReadMessage()
			if err != nil {
				logger.Errorf("read message error after reconnect: %v", err)
				continue
			}
		}
		if msgType != websocket.TextMessage {
			logger.Errorf("message type error: %v", msgType)
			continue
		}
		if len(content) == 0 {
			logger.Errorf("message content is empty")
			continue
		}
		logger.Tracef("receive message: %s", content)
		msg := model.WSResponseMessage{}
		if err := json.Unmarshal(content, &msg); err != nil {
			logger.Errorf("unmarshal message error: %v", err)
			continue
		}
		switch msg.Type {
		case model.WSMsgTypePPush:
			// proactive push
			data := model.Clipboard{}
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				logger.Errorf("unmarshal message data error: %v", err)
				continue
			}
			logger.Tracef("receive proactive push message: %v", data)
			logger.Debugf("ppush update local clipboard %v", data.Content)
			s := DetectAndConcatFileUrl(conf, &data)
			logger.Tracef("after detect and concat file url: %s", s)
			if s == conf.Runtime.ClipboardCurrentContent {
				logger.Debugf("current clipboard is same as previous when receiving ppush message, skip copy")
				continue
			}
			if err := cl.copy(s); err != nil {
				logger.Errorf("set clipboard data error: %v", err)
				continue
			}
			logger.Debugf("set clipboard data success")
			// notify the local change worker
			select {
			case loopUpdateNotify <- s:
				logger.Debugf("send loop update notify with content: %s", s)
			default:
				logger.Warn("loop update notify channel is full, skip send")
			}

		case model.WSMsgTypeData:
			// data message called by pull request
			data := make([]model.Clipboard, 0)
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				logger.Errorf("unmarshal message data error: %v", err)
				continue
			}
			logger.Tracef("receive data message: %v", data)
			// get the latest clipboard content
			if len(data) == 0 {
				logger.Debugf("Receive message: %s", msg.Msg)
				continue
			}
			theClipboard := data[0]
			s := DetectAndConcatFileUrl(conf, &theClipboard)
			if s == conf.Runtime.ClipboardCurrentContent {
				logger.Debugf("current clipboard is same as previous when receiving data message, skip copy")
				if initWorkers {
					logger.Debugf("start clipboard local change worker for the first time with same clipboard")
					go clipboardLocalChangeService(conf, cl, wso, loopUpdateNotify)
					initWorkers = false
					logger.Info("Successfully start all clipboard client workers")
				}
				continue
			}
			if err := cl.copy(s); err != nil {
				logger.Errorf("set clipboard data error: %v", err)
				continue
			}
			if initWorkers {
				logger.Tracef("init clipboard content: %v", s)
				logger.Debugf("start clipboard local change worker for the first time")
				go clipboardLocalChangeService(conf, cl, wso, loopUpdateNotify)
				initWorkers = false
				logger.Info("Successfully start all clipboard client workers")
			} else {
				// we got clipboard synchronize data response.
				// To avoid the local change worker to send the same data to server,
				// send loopUpdateNotify message.
				select {
				case loopUpdateNotify <- s:
					logger.Debugf("send loop update notify with content: %s", s)
				default:
					logger.Warn("loop update notify channel is full, skip send")
				}
			}
		default:
			logger.Warnf("unknown message type: %v", msg.Type)
			continue
		}

	}

}
