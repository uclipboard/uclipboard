package model

import (
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var Version string = "v0.0.0"

type Conf struct {
	Token  string `toml:"token"`
	Client struct {
		ServerUrl  string `toml:"server_url"`
		Connect    string `toml:"connect"`
		Adapter    string `toml:"adapter"`
		Interval   int    `toml:"interval"`
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

	Runtime struct {
		Mode         string
		ConfPath     string
		LogLevel     string
		Msg          string
		Download     string
		Upload       string
		Pull         bool
		Latest       bool
		Test         string
		TokenEncrypt string
		ShowVersion  bool
		LogPath      string
		ShowHelp     bool
	}
}

// Why go-toml? because gin need it!
// Default tag is not supported by go-toml now,so I have to implement this
func NewConfWithDefault() *Conf {
	// I want to implement this by reflect and chain calling
	// What a pity! it is too hard to resolve nest structure
	// in a simple function.
	c := Conf{}
	c.Client.Interval = 1000
	c.Client.Connect = "http"
	c.Client.XSelection = "clipboard"

	c.Server.DBPath = "./uclipboard.db"
	c.Server.TmpPath = "./tmp/"
	c.Server.TimerInterval = 60
	c.Server.PullHistorySize = 5
	c.Server.DefaultFileLife = 60 * 5 * 1000 //ms
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

func FormatConf(conf *Conf) *Conf {
	// delete the last '/' of server url
	if len(conf.Client.ServerUrl) > 0 && conf.Client.ServerUrl[len(conf.Client.ServerUrl)-1] == '/' {
		conf.Client.ServerUrl = conf.Client.ServerUrl[:len(conf.Client.ServerUrl)-1]
	}
	conf.Server.DefaultFileLife *= 1000
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
