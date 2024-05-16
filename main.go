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
	flag.StringVar(&conf.Flags.Mode, "mode", "instant", "Specify the running mode. (client|server|instant)")
	flag.StringVar(&conf.Flags.LogLevel, "log-level", "info", "logger level [info/debug/trace]")
	flag.StringVar(&conf.Flags.ConfPath, "conf", "./conf.toml", "Specify the config path.")
	flag.StringVar(&conf.Flags.Msg, "msg", "", "(instant mode) push clipboard data.")
	flag.StringVar(&conf.Flags.Upload, "upload", "", "(instant mode) upload what ever file you want.")
	flag.StringVar(&conf.Flags.Download, "download", "",
		"(instant mode) specify the file name you want to download."+
			"You can specify the file name to download the latest file."+
			"You can also specify file id by @id to download the file you want. e.g. -download @123")
	flag.BoolVar(&conf.Flags.Pull, "pull", false, "(instant mode) pull clipboard data.")
	flag.BoolVar(&conf.Flags.Latest, "latest", false, "(instant mode) download latest file.")

	flag.Parse()
	model.LoggerInit(conf.Flags.LogLevel)

	logger := model.NewModuleLogger("entry")

	// modify config struct by conf file
	conf = model.LoadConf(conf)
	logger.Debugf("running Mode: %s", conf.Flags.Mode)
	logger.Debugf("config info:%v", conf)
	switch conf.Flags.Mode {
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
