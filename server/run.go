package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/dangjinghao/uclipboard/server/frontend"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ref: https://github.com/toorop/gin-logrus
func ginLoggerMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		status := c.Writer.Status()

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
			c.JSON(http.StatusUnauthorized, model.NewDefaultServeRes("unauthorized", nil))
			c.Abort()
			return
		}
		if token != conf.Runtime.TokenEncrypt {
			c.JSON(http.StatusUnauthorized, model.NewDefaultServeRes("token is incorrect", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

func Run(c *model.Conf) {
	logger := model.NewModuleLogger("http")
	core.InitDB(c)
	frontend.InitFrontend()

	go TimerGC(c)

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
	if strings.Contains(c.Runtime.Test, "c") {
		logger.Warn("Allow all cors request")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "hostname")
		r.Use(cors.New(corsConfig))
	}

	// icon
	r.StaticFileFS("/favicon.ico", "favicon.ico", frontend.AssetsFS())
	// index or anything
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", frontend.FrontendRootFS())
	})
	// assets
	r.StaticFS("/assets", frontend.AssetsFS())

	// api
	api := r.Group(model.ApiPrefix)
	{
		v0 := api.Group(model.ApiVersion)
		publicV0 := v0.Group(model.ApiPublic)

		if !strings.Contains(c.Runtime.Test, "t") {
			logger.Debugf("Token: %s and server will use it", c.Runtime.TokenEncrypt)
			v0.Use(ginAuthMiddle(c))
		} else {
			logger.Warnf("Token check is disabled")
		}

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
