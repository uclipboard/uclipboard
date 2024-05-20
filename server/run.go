package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
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

func ginAuthMiddle(conf *model.Conf) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}
		if token != conf.Runtime.TokenEncrypt {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization failed"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func Run(c *model.Conf) {
	core.InitDB(c)
	go TimerGC(c)

	logger := model.NewModuleLogger("http")
	switch c.Runtime.LogLevel {
	case "debug":
		fallthrough
	case "trace":
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
		publicV0 := v0.Group(model.ApiPublic)

		v0.Use(ginAuthMiddle(c))
		v0.GET(model.Api_Pull, HandlerPull(c))
		v0.GET(model.Api_History, HandlerHistory(c))
		v0.POST(model.Api_Push, HandlerPush(c))
		v0.POST(model.Api_Upload, HandlerUpload(c))
		v0.GET(model.Api_Download, HandlerDownload(c))

		publicV0.GET(model.ApiPublicShare, HandlerPublicShare(c))
	}
	logger.Infof("Server is running on :%d", c.Server.Port)
	if err := r.Run(":" + strconv.Itoa(c.Server.Port)); err != nil {
		logger.Fatalf("Server run error: %s", err.Error())
	}
}
