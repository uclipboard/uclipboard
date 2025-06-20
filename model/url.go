package model

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	ApiPrefix                   = "api"
	ApiVersion                  = "v0"
	ApiVersion1                 = "v1"
	Api_Push                    = "push"
	Api_Pull                    = "pull"
	Api_History                 = "history"
	Api_Upload                  = "upload"
	Api_DownloadPure            = "download"
	Api_Download                = Api_DownloadPure + "/*filename"
	Api_WS                      = "ws"
	Api_DownloadWithAccessToken = Api_Download
)

func UrlPushApi(c *UContext) string {
	u, err := url.Parse(c.Client.Connect.Url)
	if err != nil {
		panic(err)
	}
	u = u.JoinPath(ApiPrefix, ApiVersion, Api_Push)
	return u.String()
}

func UrlPullApi(c *UContext) string {
	u, err := url.Parse(c.Client.Connect.Url)
	if err != nil {
		panic(err)
	}
	u = u.JoinPath(ApiPrefix, ApiVersion, Api_Pull)
	return u.String()
}

func UrlUploadApi(c *UContext) string {
	u, err := url.Parse(c.Client.Connect.Url)
	if err != nil {
		panic(err)
	}
	u = u.JoinPath(ApiPrefix, ApiVersion, Api_Upload)
	if c.Runtime.UploadFileLifetime != 0 {
		q := u.Query()
		q.Set("lifetime", fmt.Sprintf("%d", c.Runtime.UploadFileLifetime))
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func UrlDownloadApi(c *UContext, fileName string) string {
	u, err := url.Parse(c.Client.Connect.Url)
	if err != nil {
		panic(err)
	}
	u = u.JoinPath(ApiPrefix, ApiVersion1, Api_DownloadPure, fileName)
	return u.String()
}

func UrlWsApi(c *UContext) (string, http.Header) {
	u, err := url.Parse(c.Client.Connect.Url)
	if err != nil {
		panic(err)
	}

	// Modify the scheme directly for clarity and safety.
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		panic("url must start with http:// or https://")
	}

	u = u.JoinPath(ApiPrefix, ApiVersion1, Api_WS)
	return u.String(), http.Header{
		"token": []string{c.Runtime.TokenEncrypt},
	}
}
