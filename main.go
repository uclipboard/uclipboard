package main

import (
	"flag"

	"github.com/dangjinghao/uclipboard/client"
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server"
)

func main() {
	// create default config struct
	conf := model.NewConfWithDefault()
	// modify the `run` config struct
	flag.StringVar(&conf.Run.Mode, "mode", "server", "Specify the running mode. (client|server|instant)")
	flag.StringVar(&conf.Run.ConfPath, "conf", "./conf.toml", "Specify the config path.")
	flag.StringVar(&conf.Run.Msg, "msg", "", "(instant mode) push/pull clipboard data instantly.")
	flag.BoolVar(&conf.Run.Debug, "debug", false, "print debug message")
	flag.Parse()
	// modify config struct by conf file
	conf = model.LoadConf(conf)

	switch conf.Run.Mode {
	case "server":
		server.Run(conf)
	case "client":
		client.Run(conf)
	case "instant":
		client.Instant(conf)
	default:
		panic(model.ErrUnimplement)
	}
}
