package model

import (
	"os"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var Version string = "v0.0.0"

type Conf struct {
	Token            string `toml:"token"`
	MaxClipboardSize int    `toml:"max_clipboard_size"`
	Client           struct {
		ServerUrl  string `toml:"server_url"`
		Connect    string `toml:"connect"`
		Adapter    string `toml:"adapter"`
		Interval   int    `toml:"interval"`
		Timeout    int64  `toml:"timeout"`
		XSelection string `toml:"X_selection"`
	} `toml:"client"`

	Server struct {
		DBPath                   string `toml:"db_path"`
		PullHistorySize          int    `toml:"pull_history_size"`
		TmpPath                  string `toml:"tmp_path"`
		DefaultFileLife          int64  `toml:"default_file_life"`
		TimerInterval            int    `toml:"timer_interval"`
		ClipboardHistoryPageSize int    `toml:"clipboard_history_page_size"`
		Port                     int    `toml:"port"`
		CacheMaxAge              int    `toml:"cache_max_age"`
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
func NewConfWithDefault() *Conf {
	// I want to implement this by reflect and chain calling
	// What a pity! it is too hard to resolve nest structure
	// in a simple function.
	c := Conf{}

	c.MaxClipboardSize = 1024 * 50 // 50KB

	c.Client.Interval = 1000
	c.Client.Connect = "polling"
	c.Client.XSelection = "clipboard"
	c.Client.Timeout = 10000

	c.Server.DBPath = "./uclipboard.db"
	c.Server.TmpPath = "./tmp/"
	c.Server.TimerInterval = 60
	c.Server.PullHistorySize = 5
	c.Server.DefaultFileLife = 60 * 5 //ms 5min
	c.Server.Port = 4533
	c.Server.ClipboardHistoryPageSize = 20
	c.Server.CacheMaxAge = 60 * 60 * 24 * 30 // 30 days
	return &c
}

func LoadConf(conf *Conf) *Conf {
	logger := NewModuleLogger("config_loader")
	content, err := os.ReadFile(conf.Runtime.ConfPath)
	if err != nil {
		logger.Fatalf("Can't load config file: %s", err.Error())
	}
	err = toml.Unmarshal(content, conf)
	if err != nil {
		logger.Fatalf("Can't parse config file: %s", err.Error())
	}

	return conf
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
func FormatConf(conf *Conf) *Conf {
	logger := NewModuleLogger("config_formatter")
	// delete the last '/' of server url
	if len(conf.Client.ServerUrl) > 0 && conf.Client.ServerUrl[len(conf.Client.ServerUrl)-1] == '/' {
		conf.Client.ServerUrl = conf.Client.ServerUrl[:len(conf.Client.ServerUrl)-1]
	}

	lifetimeInt := ParseTimeStr(conf.Runtime.UploadFileLifetimeStr)
	conf.Runtime.UploadFileLifetime = lifetimeInt
	logger.Debugf("parsed lifetime: %d", lifetimeInt)
	conf.Runtime.TokenEncrypt = TokenEncrypt(conf.Token)
	return conf
}

func CheckConf(conf *Conf) {
	logger := NewModuleLogger("config_checker")
	if conf.Token == "" && !strings.Contains(conf.Runtime.Test, "t") {
		logger.Fatal("token is empty, please set token in conf file.")
	}
	if conf.Client.ServerUrl == "" && conf.Runtime.Mode == "client" {
		logger.Fatal("server url is empty, please set server url in conf file.")
	}

}
