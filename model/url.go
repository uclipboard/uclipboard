package model

import (
	"fmt"
	"strings"
)

const (
	ApiPrefix        = "api"
	ApiVersion       = "v0"
	ApiVersion1      = "v1"
	Api_Push         = "push"
	Api_Pull         = "pull"
	Api_History      = "history"
	Api_Upload       = "upload"
	Api_Download     = "download/*filename"
	Api_DownloadPure = "download"
	Api_WS           = "ws"
)

func UrlPushApi(c *UContext) string {
	return fmt.Sprintf("%s/%s/%s/%s?token=%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_Push, c.Runtime.TokenEncrypt)
}

func UrlPullApi(c *UContext) string {
	return fmt.Sprintf("%s/%s/%s/%s?token=%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_Pull, c.Runtime.TokenEncrypt)
}

func UrlUploadApi(c *UContext) string {
	str := fmt.Sprintf("%s/%s/%s/%s?token=%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_Upload, c.Runtime.TokenEncrypt)
	if c.Runtime.UploadFileLifetime != 0 {
		str += fmt.Sprintf("&lifetime=%d", c.Runtime.UploadFileLifetime)
	}
	return str
}

func UrlDownloadApi(c *UContext, fileName string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s?token=%s", c.Client.Connect.Url, ApiPrefix, ApiVersion, Api_DownloadPure, fileName, c.Runtime.TokenEncrypt)
}

func UrlWsApi(c *UContext) string {
	wsUrl := c.Client.Connect.Url
	if strings.HasPrefix(wsUrl, "http://") {
		wsUrl = strings.Replace(wsUrl, "http://", "ws://", 1)
	} else if strings.HasPrefix(wsUrl, "https://") {
		wsUrl = strings.Replace(wsUrl, "https://", "wss://", 1)
	} else {
		panic("url must start with http:// or https://")
	}
	return fmt.Sprintf("%s/%s/%s/%s?token=%s", wsUrl, ApiPrefix, ApiVersion1, Api_WS, c.Runtime.TokenEncrypt)
}
