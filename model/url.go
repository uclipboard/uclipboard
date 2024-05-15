package model

import "fmt"

const (
	ApiPrefix        = "api"
	ApiVersion       = "v0"
	Api_Push         = "push"
	Api_Pull         = "pull"
	Api_History      = "history"
	Api_Upload       = "upload"
	Api_Download     = "download/*filename"
	Api_DownloadPure = "download"
)

func UrlPushApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Push)
}

func UrlPullApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Pull)
}

func UrlUploadApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Upload)
}

func UrlDownloadApi(c *Conf, fileName string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_DownloadPure, fileName)
}
