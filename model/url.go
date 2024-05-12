package model

import "fmt"

const (
	ApiPrefix   = "api"
	ApiVersion  = "v0"
	Api_Push    = "push"
	Api_Pull    = "pull"
	Api_History = "history"
	Api_Upload  = "upload"
)

func UrlPushApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Push)
}

func UrlPullApi(c *Conf) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.Client.ServerUrl, ApiPrefix, ApiVersion, Api_Pull)
}
