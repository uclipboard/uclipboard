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

func UrlPushApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s?token=%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Push, c.Runtime.TokenEncrypt)
}

func UrlPullApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s?token=%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Pull, c.Runtime.TokenEncrypt)
}

func UrlUploadApi(c *Conf) string {
	str := fmt.Sprintf("%s/%s/%s/%s?token=%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Upload, c.Runtime.TokenEncrypt)
	if c.Runtime.UploadFileLifetime != 0 {
		str += fmt.Sprintf("&lifetime=%d", c.Runtime.UploadFileLifetime)
	}
	return str
}

func UrlDownloadApi(c *Conf, fileName string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s?token=%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_DownloadPure, fileName, c.Runtime.TokenEncrypt)
}
