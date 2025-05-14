package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/server/core"
)

func wsServerPingPong(uctx *model.UContext, wso *model.WsObject) {
	logger := model.NewModuleLogger("wsServerPing")
	wso.SetPongHandler(func(string) error {
		logger.Trace("Received pong message")
		return nil
	})
	logger.Debugf("ws ping interval: %dms", uctx.Server.Api.PingInterval)
	for {
		time.Sleep(time.Duration(uctx.Server.Api.PingInterval) * time.Millisecond)
		logger.Trace("Sending ping message")
		if err := wso.WritePing(); err != nil {
			if err == websocket.ErrCloseSent {
				logger.Debug("Websocket has been closed.")
				return
			}
			logger.Errorf("Failed to send ping message: %v", err)
			return
		}
	}
}

func wsServerProactivePush(uctx *model.UContext, wso *model.WsObject) {
	logger := model.NewModuleLogger("wsServerPush")
	sub := uctx.Runtime.ClipboardUpdateNotify.Subscribe()
	defer uctx.Runtime.ClipboardUpdateNotify.Unsubscribe(sub)
	for {
		logger.Debug("Waiting for clipboard update notification")
		msg, ok := <-sub
		if !ok {
			logger.Debug("Clipboard update notification channel closed")
			return
		}
		logger.Debug("Received clipboard update notification")
		cew, ok := msg.(*model.ClipboardExcludeWso)
		if !ok {
			wso.ErrorMsg("Invalid message type from notify msgqueue")
			return
		}
		if wso == cew.Wso {
			logger.Debug("Skip the message from self")
			continue
		}
		// push the clipboard data to the client
		data, err := json.Marshal(cew.Cb)
		if err != nil {
			wso.ErrorMsg("Marshal clipboardData error: %v", err)
			return
		}
		if err := wso.ResponseMsg(model.WSMsgTypePPush, "ok", data); err != nil {
			logger.Errorf("Failed to send response message: %v", err)
			return
		}
	}
}

// ws api is used to pass the clipboard text. File manager should be managed by the http api
func HandlerWebSocket(uctx *model.UContext) gin.HandlerFunc {
	logger := model.NewModuleLogger("HandlerWebSocket")
	return func(ctx *gin.Context) {
		// default upgrader
		wsUpgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins
				return true
			},
		}
		logger.Debugf("WebSocket upgrade request from %s", ctx.Request.RemoteAddr)
		ws, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			logger.Errorf("Failed to upgrade connection: %v", err)
			return
		}
		// we don't need to reconnect the websocket connection
		// so those error handling arguments are not needed
		wso := model.NewWsObject(ws, nil, "")
		defer wso.Close()

		go wsServerPingPong(uctx, wso)

		go wsServerProactivePush(uctx, wso)

		for {
			msgType, p, err := wso.ReadMessage()
			if err != nil {
				if err = wso.ServerErrorHandle(err); err != nil {
					logger.Errorf("Failed to read message: %v", err)
					return
				}
			}
			if msgType != websocket.TextMessage {
				logger.Errorf("Invalid message type: %d", msgType)
				return
			}
			logger.Debugf("Received message: %v", string(p))
			wsMsg := model.WSBaseMessage{}
			if err := json.Unmarshal(p, &wsMsg); err != nil {
				wso.ErrorMsg("Unmarshal wsMsg error: %v", err)
				return
			}
			logger.Debugf("Received message type: %v", wsMsg.Type)
			switch wsMsg.Type {
			case model.WSMsgTypePush:
				// WSRequestPushMessage
				clipboardData := model.NewClipboardWithDefault()
				if err := json.Unmarshal(p, clipboardData); err != nil {
					wso.ErrorMsg("Unmarshal clipboardData error: %v", err)
					return
				}
				// we should not trust the client's timestamp
				clipboardData.Ts = time.Now().UnixMilli()
				logger.Debugf("Received clipboard data: %v", clipboardData)
				if len(clipboardData.Content) > uctx.ContentLengthLimit {
					wso.ErrorMsg("clipboard is too large: %dB > %dB]", len(clipboardData.Content), uctx.ContentLengthLimit)
					return
				}
				if clipboardData.Content == "" {
					wso.ErrorMsg("content is empty")
					return
				}
				if err := core.AddClipboardRecordAndNotify(uctx, &model.ClipboardExcludeWso{Cb: clipboardData, Wso: wso}); err != nil {
					wso.ErrorMsg("AddClipboardRecordAndNotify error: %v", err)
					return
				}
				if err := wso.ResponseMsg(model.WSMsgTypeData, "ok", nil); err != nil {
					logger.Errorf("Failed to send response message: %v", err)
					return
				}
			case model.WSMsgTypePull:
				// pull request contains only the type field, it should be WSBaseMessage
				clipboardArr, err := core.QueryLatestClipboardRecord(uctx.Server.Api.PullSize)
				if err != nil {
					wso.ErrorMsg("GetLatestClipboardRecord error: %v", err)
					return
				}
				clipboardsBytes, err := json.Marshal(clipboardArr)
				if err != nil {
					wso.ErrorMsg("Marshal clipboardArr error: %v", err)
					return
				}
				// send the clipboard data to the client
				if err := wso.ResponseMsg(model.WSMsgTypeData, "ok", clipboardsBytes); err != nil {
					logger.Errorf("Failed to send response message: %v", err)
					return
				}
			default:
				logger.Warnf("Unknown message type: %v", wsMsg.Type)
				return
			}
		}
	}
}
