package client

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/uclipboard/uclipboard/client/adapter"
	"github.com/uclipboard/uclipboard/model"
)

// because there are 2 or more goroutines to access the clipboard in persist mode,
// we need to use a mutex to lock the clipboard
// to protect the clipboard
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
		logger.Debugf("clipboard content changed: %s", newContent)
		// send the content to the channel
		notify <- newContent
	})
	// normally this will never happen
	if err != nil {
		logger.Fatalf("watch error: %v", err)
	}

}

func clipboardLocalChangeService(u *model.UContext, cl *clipboardLock, wso *model.WsObject, loopUpdateNotify chan any) {
	logger := model.NewModuleLogger("clipboardLocalChangeWatcher")
	// create a updateNotify channel
	updateNotify := make(chan string, 20)
	go clipboardLocalChangeWatchService(cl, updateNotify)
	// lockedWarningCounter := 0
	for {
		// wait for the clipboard content changed
		// this will block until the clipboard content changed
		currentClipboard := <-updateNotify
		// if there are any data in the loopNotify channel, we just continue
		select {
		case <-loopUpdateNotify:
			logger.Debugf("receive loop notify, that's means same clipboard notify, skip clipboard local change")
			continue
		default:
			// do nothing
		}
		logger.Debugf("clipboard content changed: %s", currentClipboard)
		// prev is used to compare the current clipboard content and prev content
		// to avoid this situation: ppush -> copy -> local change -> paste -> push -> ppush
		// if the current clipboard content is the same as the prev content,
		// we don't need to send the websocket push
		// if we don't do this, the websocket push will be sent after receiving the ppush message
		// and the push will be sent again after handling the ppush message
		// prev := cl.prev()
		// get the clipboard content
		// if err != nil {
		// 	if err == adapter.ErrEmptyClipboard {
		// 		logger.Debugf(`adapter.Paste error:%v, set empty string clipboard.`, err)
		// 		currentClipboard = ""
		// 	} else if err == adapter.ErrLockedClipboard {
		// 		if lockedWarningCounter < 3 {
		// 			logger.Info("clipboard is locked, skip push")
		// 			lockedWarningCounter++
		// 		} else {
		// 			logger.Debugf("clipboard is locked, skip push, but the warning counter has reached the maximum value: %v", lockedWarningCounter)
		// 		}
		// 		continue
		// 	} else if err == adapter.ErrClipboardDataTypeUnknown {
		// 		logger.Debugf("the content type of clipboard is unrecgnized.")
		// 		currentClipboard = ""
		// 	} else {
		// 		logger.Warnf("adapter.Paste error:%v", err)
		// 		continue
		// 	}

		// }
		// lockedWarningCounter = 0
		if currentClipboard == "" {
			logger.Debug("skip push detect because current clipboard is empty")
			continue
		}
		if len(currentClipboard) > u.ContentLengthLimit {
			logger.Debug("current clipboard size is too large, skip push")
			continue
		}
		// if prev == currentClipboard {
		// 	logger.Debugf("clipboard content is the same as before, skip push")
		// 	continue
		// }
		// send the clipboard content to the server
		if _, err := SendWebSocketPush(currentClipboard, wso); err != nil {
			logger.Errorf("send websocket push error: %v", err)
			continue
		}
		logger.Debugf("send websocket push success")
	}

}

func persistMainLoop(conf *model.UContext, theAdapter adapter.ClipboardCmdAdapter, _ *http.Client) {
	logger := model.NewModuleLogger("persist")
	logger.Tracef("into persist mainLoop")
	wso, err := CreateWsConn(conf)
	if err != nil {
		logger.Errorf("create ws connection error: %v", err)
		return
	}
	defer wso.Close()

	cl := newClipboardLock(theAdapter)

	// send the pull request to the server
	// then we can get the latest clipboard content in the
	// case model.WSMsgTypeData
	if err := SendWebSocketPull(wso); err != nil {
		logger.Errorf("init clipboard error: %v", err)
		return
	}
	loopUpdateNotify := make(chan any, 20)
	initWorkers := true
	for {
		msgType, content, err := wso.ReadMessage()
		if err != nil {
			logger.Errorf("read message error: %v", err)
			continue
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
			logger.Debugf("receive proactive push message: %v", data)
			// notify the local change worker
			select {
			case loopUpdateNotify <- 'U':
				logger.Debugf("send loop update notify")
			default:
				logger.Warn("loop update notify channel is full, skip send")
			}
			logger.Debugf("update clipboard")
			if err := cl.copy(data.Content); err != nil {
				logger.Errorf("set clipboard data error: %v", err)
				continue
			}

		case model.WSMsgTypeData:
			// data message called by pull request
			data := make([]model.Clipboard, 0)
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				logger.Errorf("unmarshal message data error: %v", err)
				continue
			}
			logger.Debugf("receive data message: %v", data)
			// get the latest clipboard content
			if len(data) == 0 {
				logger.Debug("no clipboard data")
				continue
			}
			theClipboard := data[0]
			if err := cl.copy(theClipboard.Content); err != nil {
				logger.Errorf("set clipboard data error: %v", err)
				continue
			}
			if initWorkers {
				logger.Debugf("init clipboard content: %s", theClipboard.Content)
				logger.Debugf("start clipboard local change worker for the first time")
				go clipboardLocalChangeService(conf, cl, wso, loopUpdateNotify)
				initWorkers = false
			} else {
				// we got clipboard synchronize data response, to avoid the local change worker to send the same data to server
				select {
				case loopUpdateNotify <- 'U':
					logger.Debugf("send loop update notify")
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
