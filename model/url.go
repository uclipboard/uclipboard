package model

import "fmt"

const (
	ApiPrefix        = "api"
	ApiVersion       = "v0"
	ApiPublic        = "public"
	Api_Push         = "push"
	Api_Pull         = "pull"
	Api_History      = "history"
	Api_Upload       = "upload"
	Api_Download     = "download/*filename"
	Api_DownloadPure = "download"

	ApiPublicShare = "share"
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
