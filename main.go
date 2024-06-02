package main

import (
	"fmt"
	"os"

	"flag"

	"github.com/uclipboard/uclipboard/client"
	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/server"
)

func main() {
	// create default config struct
	conf := model.NewConfWithDefault()

	model.InitFlags(conf)
	if conf.Runtime.ShowVersion {
		fmt.Println("uclipboard version: ", model.Version)
		os.Exit(0)
	}
	if conf.Runtime.ShowHelp {
		fmt.Println("uclipboard version: ", model.Version)
		fmt.Println("Usage of uclipboard:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	// before InitLogger, we can't use logger
	model.InitLogger(conf)
	// modify config struct by conf file
	conf = model.LoadConf(conf)
	conf = model.FormatConf(conf)
	model.CheckConf(conf)

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
