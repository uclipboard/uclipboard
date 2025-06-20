package model

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	ApiPrefix                   = "api"
	ApiVersion                  = "v0"
	ApiVersion1                 = "v1"
	Api_Push                    = "push"
	Api_Pull                    = "pull"
	Api_History                 = "history"
	Api_Upload                  = "upload"
	Api_Download                = "download/*filename"
	Api_DownloadPure            = "download"
	Api_WS                      = "ws"
	Api_DownloadWithAccessToken = "download/*filename"
)

func UrlPushApi(c *UContext) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_Push)
}

func UrlPullApi(c *UContext) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_Pull)
}

func UrlUploadApi(c *UContext) string {
	str := fmt.Sprintf("%s/%s/%s/%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_Upload)
	if c.Runtime.UploadFileLifetime != 0 {
		str += fmt.Sprintf("?lifetime=%d", c.Runtime.UploadFileLifetime)
	}
	return str
}

func UrlDownloadApi(c *UContext, fileName string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", c.Client.Connect.Url, ApiPrefix, ApiVersion1, Api_DownloadPure, fileName)
}

func UrlWsApi(c *UContext) (string, http.Header) {
	wsUrl := c.Client.Connect.Url
	if strings.HasPrefix(wsUrl, "http://") {
		wsUrl = strings.Replace(wsUrl, "http://", "ws://", 1)
	} else if strings.HasPrefix(wsUrl, "https://") {
		wsUrl = strings.Replace(wsUrl, "https://", "wss://", 1)
	} else {
		panic("url must start with http:// or https://")
	}
	return fmt.Sprintf("%s/%s/%s/%s", wsUrl, ApiPrefix, ApiVersion1, Api_WS), http.Header{
		"token": []string{c.Runtime.TokenEncrypt},
	}
}
