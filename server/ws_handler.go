package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/uclipboard/uclipboard/model"
)

func HandlerWebSocket(uctx *model.UContext) gin.HandlerFunc {
	logger := model.NewModuleLogger("HandlerWebSocket")
	return func(ctx *gin.Context) {
		// default upgrader
		wsUpgrader := websocket.Upgrader{}
		ws, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			logger.Errorf("Failed to upgrade connection: %v", err)
			return
		}
		defer ws.Close()
		msgType, p, err := ws.ReadMessage()
		if err != nil {
			logger.Errorf("Failed to read message: %v", err)
			return
		}
		logger.Debugf("Received ping message: %s", string(p))
		switch msgType {
		case websocket.PingMessage:
			err = ws.WriteMessage(websocket.PongMessage, []byte("pong"))
			if err != nil {
				logger.Errorf("Failed to write pong message: %v", err)
				return
			}
		case websocket.PongMessage:
			err = ws.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				logger.Errorf("Failed to write ping message: %v", err)
				return
			}
		default:
			logger.Errorf("Unsupported message type: %d", msgType)
			return
		}
	}
}
