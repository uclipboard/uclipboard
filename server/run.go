package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/server/core"
	"github.com/uclipboard/uclipboard/server/frontend"
)

// ref: https://github.com/toorop/gin-logrus
func ginLoggerMiddleware() gin.HandlerFunc {
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

func ginAuthMiddleware(uctx *model.UContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, model.NewDefaultServeRes("unauthorized", nil))
			c.Abort()
			return
		}
		if token != uctx.Runtime.TokenEncrypt {
			c.JSON(http.StatusUnauthorized, model.NewDefaultServeRes("token is incorrect", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

func cacheMiddleware(conf *model.UContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", conf.Server.Api.CacheMaxAge)) // 30 days
		c.Next()

	}
}

func removeCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Next()
	}
}

func Run(c *model.UContext) {
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

	if c.Server.AccessLog {
		r.Use(ginLoggerMiddleware())
	}

	if strings.Contains(c.Runtime.Test, "c") {
		logger.Warn("Allow all cors request")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "hostname")
		r.Use(cors.New(corsConfig))
	}

	if !strings.Contains(c.Runtime.Test, "f") {
		r.Use(cacheMiddleware(c))
		// icon
		r.StaticFileFS("/favicon.ico", "favicon.ico", frontend.FrontendRootFS())
		// index or anything
		r.NoRoute(func(c *gin.Context) {
			c.FileFromFS("/", frontend.FrontendRootFS())
		})
		// assets
		r.StaticFS("/assets", frontend.AssetsFS())
	} else {
		logger.Warn("Frontend is disabled")

	}
	// api
	api := r.Group(model.ApiPrefix)
	api.Use(removeCacheMiddleware())
	{
		v0 := api.Group(model.ApiVersion)
		v1 := api.Group(model.ApiVersion1)

		if !strings.Contains(c.Runtime.Test, "t") {
			logger.Debugf("Token is `%s` and server will use it", c.Runtime.TokenEncrypt)
			v0.Use(ginAuthMiddleware(c))
		} else {
			logger.Warnf("Token checker is disabled")
		}

		v0.GET(model.Api_Pull, HandlerPull(c))
		v0.GET(model.Api_History, HandlerHistory(c))
		v0.POST(model.Api_Push, HandlerPush(c))
		v0.POST(model.Api_Upload, HandlerUpload(c))
		v0.GET(model.Api_Download, HandlerDownload(c))

		v1.GET(model.Api_WS, HandlerWebSocket(c))
	}
	logger.Infof("Server is running on :%d", c.Server.Api.Port)
	if err := r.Run(":" + strconv.Itoa(c.Server.Api.Port)); err != nil {
		logger.Fatalf("Server run error: %v", err)
	}
}
