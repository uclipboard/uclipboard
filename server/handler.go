package server

import (
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/gin-gonic/gin"
)

func HandlerPush(conf *model.Conf) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
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
}
func HandlerPull(conf *model.Conf) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		var clipboardArr []model.Clipboard
		if err := core.GetLatestClipboardRecord(&clipboardArr, conf.Server.HistorySize); err != nil {
			ctx.String(500, err.Error())
		}
		ctx.JSON(200, clipboardArr)
	}
}
func HandlerHistory(ctx *gin.Context) {
}
