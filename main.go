package main

import (
	"flag"

	"github.com/dangjinghao/uclipboard/client"
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server"
)

var (
	runMode   string
	confPath  string
	debugMode bool
	msg       string
)

func main() {

	flag.StringVar(&runMode, "mode", "server", "Specify the running mode. (client|server|instant)")
	flag.StringVar(&confPath, "conf", "./conf.ini", "Specify the config path.")
	flag.StringVar(&msg, "msg", "", "(instant mode) push/pull clipboard data instantly.")
	flag.BoolVar(&debugMode, "debug", false, "print debug message")
	flag.Parse()
	conf := model.LoadConf(confPath)
	switch runMode {
	case "server":
		server.Run(conf)
	case "client":
		client.Run(conf)
	case "instant":
		client.Instant(conf, msg)
	default:
		panic(model.ErrUnimplement)
	}
}
