package server

import (
	"net/http"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/gin-gonic/gin"
)

func HandlerPush(conf *model.Conf) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		clipboardData := model.NewClipoardWithDefault()
		if err := ctx.BindJSON(&clipboardData); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server interal Error"})
			return
		}
		if err := core.AddClipboardRecord(clipboardData); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server interal Error"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}
func HandlerPull(conf *model.Conf) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		var clipboardArr []model.Clipboard
		if err := core.GetLatestClipboardRecord(&clipboardArr, conf.Server.HistorySize); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
		}
		ctx.JSON(http.StatusOK, clipboardArr)
	}
}
func HandlerHistory(ctx *gin.Context) {
	// TODO:Designed for WebUI, pulling all data from server
}
