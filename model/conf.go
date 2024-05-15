package model

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Conf struct {
	Client struct {
		ServerUrl  string `toml:"server_url"`
		Connect    string `toml:"connect"`
		Adapter    string `toml:"adapter"`
		Interval   int    `toml:"interval"`
		XSelection string `toml:"X_selection"`
	} `toml:"client"`

	Server struct {
		DBPath          string `toml:"db_path"`
		PullHistorySize int    `toml:"pull_history_size"`
		TmpPath         string `toml:"tmp_path"`
		DefaultFileLife int64  `toml:"default_file_life"`
		TimerInterval   int    `toml:"timer_interval"`
		MaxHistorySize  int    `toml:"max_history_size"`
	} `toml:"server"`

	Flags struct {
		Mode     string
		ConfPath string
		LogLevel string
		Msg      string
		Download string
		Upload   string
		Pull     bool
		Latest   bool
	}
}

// Why go-toml? because gin need it!
// Default tag is not supported by go-toml now,so I have to implement this
func NewConfWithDefault() *Conf {
	// I want to implement this by reflect and chain calling
	// What a pity! it is too hard to resolve nest structure
	// in a simple function.
	c := Conf{}
	c.Client.Connect = "http"
	c.Client.Interval = 1000
	c.Server.PullHistorySize = 5
	c.Server.DBPath = "./uclipboard.db"
	c.Client.XSelection = "clipboard"
	c.Server.TmpPath = "./tmp/"
	c.Server.TimerInterval = 60
	c.Server.DefaultFileLife = 60 * 5
	return &c
}

func LoadConf(conf *Conf) *Conf {
	logger := NewModuleLogger("config_loader")
	content, err := os.ReadFile(conf.Flags.ConfPath)
	if err != nil {
		logger.Fatalf("Can't load config file: %s", err.Error())
	}
	err = toml.Unmarshal(content, conf)
	if err != nil {
		logger.Fatalf("Can't parse config file: %s", err.Error())
	}

	return conf
}
