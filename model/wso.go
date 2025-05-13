package model

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	WSMsgTypeErr   = "error"
	WSMsgTypeData  = "data"
	WSMsgTypePush  = "push"
	WSMsgTypePPush = "ppush"
	WSMsgTypePull  = "pull"
)

type WsObject struct {
	ws     *websocket.Conn
	wlock  sync.Mutex
	rlock  sync.Mutex
	logger *logrus.Entry
}

func NewWsObject(ws *websocket.Conn) *WsObject {
	return &WsObject{
		ws:     ws,
		wlock:  sync.Mutex{},
		rlock:  sync.Mutex{},
		logger: NewModuleLogger("wsObject"),
	}
}

func (wso *WsObject) ErrorMsg(fmtstr string, args ...any) {
	wso.logger.Errorf(fmtstr, args...)
	// format the error message with args
	fullErrorMsg := fmt.Sprintf(fmtstr, args...)
	if err := wso.ResponseMsg(WSMsgTypeErr, fullErrorMsg, nil); err != nil {
		wso.logger.Errorf("Failed to send error message: %v", err)
		return
	}

}

func (wso *WsObject) ResponseMsg(_type string, msg string, data []byte) error {
	wsMsg := WSMessage{
		Type: _type,
		ServerResponse: ServerResponse{
			Msg:  msg,
			Data: data,
		},
	}
	wso.wlock.Lock()
	defer wso.wlock.Unlock()

	if err := wso.ws.WriteJSON(wsMsg); err != nil {
		return err
	}
	wso.logger.Debugf("Sent message: %v", wsMsg)
	return nil
}

func (wso *WsObject) ReadMessage() (msgType int, p []byte, err error) {
	wso.rlock.Lock()
	defer wso.rlock.Unlock()
	msgType, p, err = wso.ws.ReadMessage()
	return
}

func (wso *WsObject) WritePing() error {
	wso.wlock.Lock()
	defer wso.wlock.Unlock()
	if err := wso.ws.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
		return err
	}
	return nil
}

func (wso *WsObject) SetPongHandler(f func(string) error) {
	wso.ws.SetPongHandler(f)
}

func (wso *WsObject) Close() error {
	wso.wlock.Lock()
	defer wso.wlock.Unlock()
	wso.rlock.Lock()
	defer wso.rlock.Unlock()

	if err := wso.ws.Close(); err != nil {
		return err
	}
	return nil
}
