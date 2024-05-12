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
		DBPath string `toml:"db_path"`
	} `toml:"server"`

	// passed by arguments
	Run struct {
		Mode     string
		ConfPath string
		Debug    bool
		Msg      string
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
	c.Server.DBPath = "./uclipboard.db"
	return &c
}

func LoadConf(conf *Conf) *Conf {
	content, err := os.ReadFile(conf.Run.ConfPath)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(content, conf)
	if err != nil {
		panic(err)
	}

	return conf
}
