package server

import (
	"github.com/dangjinghao/uclipboard/model"
	"github.com/gin-gonic/gin"
)

func Run(c *model.Conf) {
	r := gin.Default()
	api := r.Group(model.ApiPrefix)
	{
		v0 := api.Group(model.ApiVersion)
		v0.GET(model.Api_Pull, HandlerPull)
		v0.GET(model.Api_History, HandlerHistory)
		v0.POST(model.Api_Push, HandlerPush)
	}
	r.Run()
}
