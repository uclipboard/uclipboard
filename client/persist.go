package client

import (
	"encoding/json"
	"net/http"
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
		logger.Debugf("Clipboard content changed, notifying")
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
	prevContext := ""
	for {
		func() {
			// wait for the clipboard content changed
			// this will block until the clipboard content changed
			currentClipboard := <-updateNotify
			defer func() {
				logger.Debugf("set prevContext to currentClipboard: %s", currentClipboard)
				prevContext = currentClipboard
			}()
			// if there are any data in the loopNotify channel, we just return
			select {
			case <-loopUpdateNotify:
				logger.Debugf("receive loop notify, that's means same clipboard notify, skip clipboard local change")
				// we still need to update the prevContext
				return
			default:
				// do nothing
			}
			// it is not recommended to put this before the select statement
			// if we skip it, we can't consume the message
			// and that may cause the channel to be full
			if currentClipboard == prevContext {
				logger.Debugf("current clipboard is same as previous, skip push")
				return
			}

			logger.Debugf("Receive clipboard content changed notify: %v", currentClipboard)
			if currentClipboard == "" {
				logger.Debug("skip push detect because current clipboard is empty")
				return
			} else if len(currentClipboard) > u.ContentLengthLimit {
				logger.Debug("current clipboard size is too large, skip push")
				currentClipboard = currentClipboard[:u.ContentLengthLimit]
				return
			}

			if _, err := SendWebSocketPush(currentClipboard, wso); err != nil {
				logger.Errorf("send websocket push error: %v", err)
				return
			}
			logger.Debugf("send websocket push success")
		}()
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

	logger.Info("Successfully connected to server.")

	defer wso.Close()
	timeout := time.Duration(conf.Client.Connect.Timeout) * time.Millisecond
	wso.InitClientPingHandler(timeout)
	cl := newClipboardLock(theAdapter)

	loopUpdateNotify := make(chan any, 20)

	// send the pull request to the server
	// then we can get the latest clipboard content in the
	// case model.WSMsgTypeData
	if err := SendWebSocketPull(wso); err != nil {
		logger.Errorf("init clipboard error: %v", err)
		return
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
			if err := cl.copy(s); err != nil {
				logger.Errorf("set clipboard data error: %v", err)
				continue
			}
			logger.Debugf("set clipboard data success")
			// notify the local change worker
			select {
			case loopUpdateNotify <- 'U':
				logger.Debugf("send loop update notify")
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
			if err := cl.copy(s); err != nil {
				logger.Errorf("set clipboard data error: %v", err)
				continue
			}
			if initWorkers {
				logger.Tracef("init clipboard content: %s", theClipboard.Content)
				logger.Debugf("start clipboard local change worker for the first time")
				go clipboardLocalChangeService(conf, cl, wso, loopUpdateNotify)
				initWorkers = false
				logger.Info("Successfully start all clipboard client workers")
			} else {
				// we got clipboard synchronize data response.
				// To avoid the local change worker to send the same data to server,
				// send loopUpdateNotify message.
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
