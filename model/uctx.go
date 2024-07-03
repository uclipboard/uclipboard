package model

import (
	"os"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var Version string = "vNOVERSION"

type UContext struct {
	Token              string `toml:"token"`
	ContentLengthLimit int    `toml:"content_length_limit"`
	Client             struct {
		Connect struct {
			Type          string `toml:"type"`
			Interval      int    `toml:"interval"`
			Url           string `toml:"url"`
			Timeout       int    `toml:"timeout"`
			UploadTimeout int    `toml:"upload_timeout"`
		} `toml:"connect"`

		Adapter struct {
			Type       string `toml:"type"`
			XSelection string `toml:"X_selection"`
		} `toml:"adapter"`
	} `toml:"client"`

	Server struct {
		Api struct {
			PullSize        int `toml:"pull_size"`
			HistoryPageSize int `toml:"history_page_size"`
			Port            int `toml:"port"`
			CacheMaxAge     int `toml:"cache_max_age"`
		} `toml:"api"`

		Store struct {
			DBPath                   string `toml:"db_path"`
			TmpPath                  string `toml:"tmp_path"`
			DefaultFileLife          int64  `toml:"default_file_life"`
			MaxClipboardRecordNumber int    `toml:"max_clipboard_record_number"`
		} `toml:"store"`

		TimerInterval int `toml:"timer_interval"`
	} `toml:"server"`
	// All struct should be read-only except runtime after LoadConf
	Runtime struct {
		Mode                  string
		ConfPath              string
		LogLevel              string
		PushMsg               string
		Download              string
		Upload                string
		Pull                  bool
		Latest                bool
		Test                  string
		TokenEncrypt          string
		ShowVersion           bool
		LogPath               string
		ShowHelp              bool
		UploadFileLifetime    int64
		UploadFileLifetimeStr string
	}
}

// Why go-toml? because gin need it!
// Default tag is not supported by go-toml now,so I have to implement this
func NewUCtxWithDefault() *UContext {
	// I want to implement this by reflect and chain calling
	// What a pity! it is too hard to resolve nest structure
	// in a simple function.
	c := UContext{}

	c.ContentLengthLimit = 1024 * 50 // 50KB

	c.Client.Connect.Interval = 1000
	c.Client.Connect.Type = "polling"
	c.Client.Connect.Timeout = 10000
	c.Client.Connect.UploadTimeout = 300
	c.Client.Adapter.XSelection = "clipboard"

	c.Server.TimerInterval = 60
	c.Server.Store.DBPath = "./uclipboard.db"
	c.Server.Store.TmpPath = "./tmp/"
	c.Server.Store.DefaultFileLife = 60 * 5 //s 3m
	c.Server.Store.MaxClipboardRecordNumber = 0
	c.Server.Api.PullSize = 5
	c.Server.Api.Port = 4533
	c.Server.Api.HistoryPageSize = 20
	c.Server.Api.CacheMaxAge = 60 * 60 * 24 * 30 // 30 days
	return &c
}

func LoadConf(uctx *UContext) *UContext {
	logger := NewModuleLogger("config_loader")
	content, err := os.ReadFile(uctx.Runtime.ConfPath)
	if err != nil {
		logger.Fatalf("Can't load config file: %s", err.Error())
	}
	err = toml.Unmarshal(content, uctx)
	if err != nil {
		logger.Fatalf("Can't parse config file: %s", err.Error())
	}

	return uctx
}

// return secs
func ParseTimeStr(t string) int64 {
	logger := NewModuleLogger("time_parser")
	if t == "" {
		return 0
	}
	var unit int64
	parseLen := len(t) - 1
	switch t[len(t)-1] {
	case 's':
		unit = 1
	case 'm':
		unit = 60
	case 'h':
		unit = 60 * 60
	case 'd':
		unit = 60 * 60 * 24
	default:
		unit = 1
		parseLen += 1
	}
	numberNoUnit, err := strconv.ParseInt(t[:parseLen], 10, 64)
	if err != nil {
		logger.Warnf("invalid lifetime: %s", t)
		return 0
	}
	return numberNoUnit * unit

}
func FormatConf(uctx *UContext) *UContext {
	// logger := NewModuleLogger("config_formatter")
	// delete the last '/' of server url
	if len(uctx.Client.Connect.Url) > 0 && uctx.Client.Connect.Url[len(uctx.Client.Connect.Url)-1] == '/' {
		uctx.Client.Connect.Url = uctx.Client.Connect.Url[:len(uctx.Client.Connect.Url)-1]
	}

	lifetimeInt := ParseTimeStr(uctx.Runtime.UploadFileLifetimeStr)
	uctx.Runtime.UploadFileLifetime = lifetimeInt
	// logger.Debugf("default file lifetime: %d", lifetimeInt)
	uctx.Runtime.TokenEncrypt = TokenEncrypt(uctx.Token)
	return uctx
}

func CheckConf(uctx *UContext) {
	logger := NewModuleLogger("config_checker")
	if uctx.Token == "" && !strings.Contains(uctx.Runtime.Test, "t") {
		logger.Fatal("token is empty, please set token in conf file.")
	}
	if uctx.Client.Connect.Url == "" && uctx.Runtime.Mode == "client" {
		logger.Fatal("server url is empty, please set server url in conf file.")
	}

}
