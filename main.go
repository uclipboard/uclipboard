package main

import (
	"fmt"
	"os"

	"github.com/dangjinghao/uclipboard/client"
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server"
)

func main() {
	// create default config struct
	conf := model.NewConfWithDefault()

	model.InitFlags(conf)
	if conf.Runtime.ShowVersion {
		fmt.Println("uclipboard version: ", model.Version)
		os.Exit(0)
	}
	// before InitLogger, we can't use logger
	model.InitLogger(conf)
	// modify config struct by conf file
	conf = model.LoadConf(conf)
	conf = model.FormatConf(conf)
	model.CheckConf(conf)

	conf.Runtime.TokenEncrypt = model.TokenEncrypt(conf.Token)

	logger := model.NewModuleLogger("entry")

	logger.Debugf("running Mode: %s", conf.Runtime.Mode)
	logger.Debugf("config info:%v", conf)
	switch conf.Runtime.Mode {
	case "server":
		server.Run(conf)
	case "client":
		client.Run(conf)
	case "instant":
		client.Instant(conf)
	default:
		logger.Fatal("unknown running mode!")
	}
}
