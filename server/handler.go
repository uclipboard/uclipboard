package server

import (
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/gin-gonic/gin"
)

func HandlerPush(ctx *gin.Context) {
	act := model.NewClipboardAction()
	act.Act = model.ActCmdPushClipboard
	act.Ctx = ctx
	var clipboard model.ClipboardItem
	if err := ctx.ShouldBindJSON(&clipboard); err != nil {
		ctx.String(500, err.Error())
		return
	}

	act.Clipboard = clipboard
	core.PushActionChan(act)

	ret := core.PullFromReturnChan()
	// this ctx maybe not be the previous ctx
	ret.Ctx.JSON(200, ret.Clipboard)
}
func HandlerPull(ctx *gin.Context) {
	act := model.NewClipboardAction()
	act.Act = model.ActCmdPullClipboard
	act.Ctx = ctx

	core.PushActionChan(act)
	ret := core.PullFromReturnChan()
	ret.Ctx.JSON(200, ret.Clipboard)

}
func HandlerHistory(ctx *gin.Context) {
}
