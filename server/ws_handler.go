package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/uclipboard/uclipboard/model"
)

func wsServerPingPong(uctx *model.UContext, ws *websocket.Conn) {
	logger := model.NewModuleLogger("wsServerPing")
	ws.SetPongHandler(func(string) error {
		logger.Trace("Received pong message")
		return nil
	})
	logger.Debugf("ws ping interval: %dms", uctx.Server.Api.PingInterval)
	for {
		logger.Trace("Sending ping message")
		if err := ws.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
			if err == websocket.ErrCloseSent {
				logger.Debug("WebSocket connection closed")
				return
			}

			logger.Errorf("Failed to write ping message: %v", err)
			return
		}
		time.Sleep(time.Duration(uctx.Server.Api.PingInterval) * time.Millisecond)
	}
}

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
		defer ws.Close()
		// create a ping timer goroutine
		go wsServerPingPong(uctx, ws)

		msgType, p, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived, websocket.CloseServiceRestart,
				websocket.CloseTryAgainLater) {
				logger.Debug("Websocket closed.")
				return
			}
			logger.Errorf("Failed to read message: %v", err)
		}
		if msgType != websocket.BinaryMessage {
			logger.Errorf("Invalid message type: %d", msgType)
			return
		}
		logger.Debugf("Received message: %s", p)

	}
}
