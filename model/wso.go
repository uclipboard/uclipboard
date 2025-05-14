package model

import (
	"errors" // Added for error checking
	"fmt"
	"io" // Added for io.EOF, io.ErrUnexpectedEOF
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	WSMsgTypeErr   = "error"
	WSMsgTypeData  = "data"
	WSMsgTypePush  = "push"
	WSMsgTypePPush = "ppush"
	WSMsgTypePull  = "pull"

	// Constants for reconnection strategy
	initialReconnectDelay = 1 * time.Second
	maxReconnectDelay     = 30 * time.Second
)

type WsObject struct {
	ws     *websocket.Conn
	api    string // used for reconnecting
	dialer *websocket.Dialer
	wlock  sync.Mutex
	rlock  sync.Mutex
	logger *logrus.Entry
}

func NewWsObject(ws *websocket.Conn, dialer *websocket.Dialer, wsApi string) *WsObject {
	return &WsObject{
		ws:     ws,
		api:    wsApi,
		dialer: dialer,
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
	wsMsg := WSResponseMessage{
		Type: _type,
		ServerResponse: ServerResponse{
			Msg:  msg,
			Data: data,
		},
	}
	wso.WriteJSON(wsMsg)
	wso.logger.Debugf("Sent message: %v with type %v and msg %v", string(data), _type, msg)
	return nil
}

func (wso *WsObject) WriteJSON(msg any) error {
	wso.wlock.Lock()
	defer wso.wlock.Unlock()
	if err := wso.ws.WriteJSON(msg); err != nil {
		return err
	}
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

func (wso *WsObject) Reconnect() error {
	// Phase 1: Prepare for closing old connection and get necessary info under lock
	wso.wlock.Lock()
	wso.rlock.Lock()

	var oldWs *websocket.Conn
	if wso.ws != nil {
		oldWs = wso.ws
		wso.ws = nil // Prevent further use of the old connection object
	}
	apiToDial := wso.api // Copy API endpoint before unlocking

	// Unlock before potentially long-running close and dial operations
	wso.rlock.Unlock()
	wso.wlock.Unlock()

	// Phase 2: Close the old connection (if it existed)
	if oldWs != nil {
		wso.logger.Debugf("Closing existing WebSocket connection for reconnect.")
		// Log error on close but proceed, as the main goal is to establish a new connection
		if err := oldWs.Close(); err != nil {
			wso.logger.Warnf("Error closing existing WebSocket: %v. Proceeding with reconnect attempt.", err)
		}
	}

	// Phase 3: Dial new connection
	wso.logger.Infof("Attempting to dial WebSocket server at %s for reconnect.", apiToDial)
	newws, _, err := wso.dialer.Dial(apiToDial, nil)
	if err != nil {
		wso.logger.Errorf("Failed to dial WebSocket server during reconnect: %v", err)
		return err // Dialing failed
	}
	wso.logger.Info("Successfully re-established WebSocket connection.")

	// Phase 4: Update WsObject with the new connection under lock
	wso.wlock.Lock()
	defer wso.wlock.Unlock()
	wso.rlock.Lock()
	defer wso.rlock.Unlock()

	wso.ws = newws
	return nil
}

func (wso *WsObject) ServerErrorHandle(err error) error {
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,    // 1000
		websocket.CloseGoingAway,        // 1001
		websocket.CloseNoStatusReceived, // 1005 - Connection closed without a status code.
		websocket.CloseServiceRestart,   // 1012 - Server is restarting.
		websocket.CloseTryAgainLater) {  // 1013 - Temporary condition, try again.
		wso.logger.Infof("Server-initiated closure classified as normal/expected: %v", err)
		return nil
	}
	return err
}

func isRecoverableError(err error, logger *logrus.Entry) bool {
	if err == nil {
		return false
	}

	if websocket.IsUnexpectedCloseError(err) {
		logger.Debugf("Error is an UnexpectedCloseError, considered recoverable: %v", err)
		return true
	}

	if websocket.IsCloseError(err,
		websocket.CloseServiceRestart,   // 1012
		websocket.CloseTryAgainLater,    // 1013
		websocket.CloseNoStatusReceived, // 1005
		websocket.CloseAbnormalClosure,  // 1006
	) {
		logger.Debugf("Error is a recoverable WebSocket close code: %v", err)
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		logger.Debugf("Error is a net.Error (Timeout: %t, Temporary: %t), considered recoverable: %v", netErr.Timeout(), netErr.Temporary(), netErr)
		return true
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		logger.Debugf("Error is EOF or UnexpectedEOF, considered recoverable: %v", err)
		return true
	}

	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		logger.Debugf("Error is a normal closure (e.g. NormalClosure, GoingAway), considered recoverable by reconnecting: %v", err)
		return true
	}

	logger.Warnf("Error not explicitly classified as recoverable: %v (type: %T)", err, err)
	return false
}

// use a very simple strategy to handle the error
// reconnect until successful
// that means when this function returns, the connection is always ok
func (wso *WsObject) ClientErrorHandle(originalError error) error {
	wso.logger.Warnf("Attempting to handle client error: %v", originalError)

	if !isRecoverableError(originalError, wso.logger) {
		wso.logger.Errorf("Unrecoverable error detected, not attempting reconnect: %v", originalError)
		return originalError
	}

	wso.logger.Infof("Error deemed recoverable. Initiating reconnection process for: %v", originalError)

	var lastAttemptErr error = originalError
	currentDelay := initialReconnectDelay

	for i := 0; ; i++ {
		wso.logger.Infof("Reconnect attempt %d. Waiting for %v before trying. Previous error: %v", i+1, currentDelay, lastAttemptErr)
		time.Sleep(currentDelay)

		if reconnErr := wso.Reconnect(); reconnErr != nil {
			lastAttemptErr = reconnErr
			wso.logger.Errorf("Reconnect attempt %d failed: %v", i, reconnErr)

			currentDelay *= 2
			if currentDelay > maxReconnectDelay {
				currentDelay = maxReconnectDelay
			}
		} else {
			break
		}

	}
	return nil
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
