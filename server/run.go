package server

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ref: https://github.com/toorop/gin-logrus
func ginLoggerMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		status := c.Writer.Status()
		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		lat := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		dataLength := c.Writer.Size()

		entry := model.NewModuleLogger("gin").WithFields(logrus.Fields{
			"length":     dataLength,
			"user_agent": ua,
			"IP":         ip,
		})
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s [%d] %s %s (%dms)", ip, status, c.Request.Method, path, lat)
			if status >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if status >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}

func Run(c *model.Conf) {
	core.InitDB(c)
	switch c.Run.LogInfo {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "info":
		fallthrough //Unnecessary, but I want :)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(ginLoggerMiddle())
	api := r.Group(model.ApiPrefix)
	{
		v0 := api.Group(model.ApiVersion)
		v0.GET(model.Api_Pull, HandlerPull(c))
		v0.GET(model.Api_History, HandlerHistory)
		v0.POST(model.Api_Push, HandlerPush(c))
	}
	r.Run()
}
