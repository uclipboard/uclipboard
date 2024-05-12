package server

import (
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/gin-gonic/gin"
)

func HandlerPush(ctx *gin.Context) {
	clipboardData := model.NewClipoardWithDefault()
	if err := ctx.BindJSON(&clipboardData); err != nil {
		ctx.String(500, err.Error())
		return
	}
	if err := core.AddClipboardRecord(clipboardData); err != nil {
		ctx.String(500, err.Error())
		return
	}
	ctx.JSON(200, gin.H{"result": "ok"})
}
func HandlerPull(ctx *gin.Context) {
	clipboardData := model.NewClipoardWithDefault()
	if err := core.GetLatestClipboardRecord(clipboardData); err != nil {
		ctx.String(500, err.Error())
	}
	ctx.JSON(200, clipboardData)
}
func HandlerHistory(ctx *gin.Context) {
}
