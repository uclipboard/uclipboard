package model

import (
	"log"

	"github.com/go-ini/ini"
)

type Conf struct {
	Client_ServerUrl string
	Client_Connect   string
	Client_Adapter   string
	Client_Interval  int
}

func LoadConf(path string) *Conf {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatal("Loading config error:", err)
	}
	serverUrl := cfg.Section("client").Key("server_url").MustString("http://localhost:4700")
	connect := cfg.Section("client").Key("connect").MustString("http")
	adapter := cfg.Section("client").Key("adapter").MustString("wl")
	interval := cfg.Section("client").Key("interval").MustInt(1000)

	return &Conf{
		Client_ServerUrl: serverUrl,
		Client_Connect:   connect,
		Client_Adapter:   adapter,
		Client_Interval:  interval,
	}
}
