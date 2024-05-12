package model

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Conf struct {
	Client struct {
		ServerUrl string `toml:"server_url"`
		Connect   string `toml:"connect"`
		Adapter   string `toml:"adapter"`
		Interval  int    `toml:"interval"`
	} `toml:"client"`

	Server struct {
		DBPath      string `toml:"db_path"`
		HistorySize int    `toml:"history_size"`
	} `toml:"server"`

	// passed by arguments
	Run struct {
		Mode     string
		ConfPath string
		LogInfo  string
		Msg      string
		Pull     bool
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
	c.Server.HistorySize = 5
	c.Server.DBPath = "./uclipboard.db"
	return &c
}

func LoadConf(conf *Conf) *Conf {
	logger := NewModuleLogger("config_loader")
	content, err := os.ReadFile(conf.Run.ConfPath)
	if err != nil {
		logger.Fatalf("Can't load config file: %s", err.Error())
	}
	err = toml.Unmarshal(content, conf)
	if err != nil {
		logger.Fatalf("Can't parse config file: %s", err.Error())
	}

	return conf
}
