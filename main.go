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
	flag.StringVar(&conf.Flags.Msg, "msg", "", "(instant mode) push clipboard data instantly.")
	flag.StringVar(&conf.Flags.Upload, "upload", "", "(instant mode) upload what ever file instantly.")
	flag.StringVar(&conf.Flags.Download, "download", "", "(instant mode) specify the file name you want to download instantly.")
	flag.BoolVar(&conf.Flags.Pull, "pull", false, "(instant mode) pull clipboard data instantly.")
	flag.BoolVar(&conf.Flags.Latest, "latest", false, "(instant mode) download latest file instantly.")

	flag.Parse()
	model.LoggerInit(conf.Flags.LogLevel)

	logger := model.NewModuleLogger("entry")

	// modify config struct by conf file
	conf = model.LoadConf(conf)
	logger.Debugf("conf.Run.Mode= %s", conf.Flags.Mode)

	switch conf.Flags.Mode {
	case "server":
		server.Run(conf)
	case "client":
		client.Run(conf)
	case "instant":
		client.Instant(conf)
	default:
		logger.Fatal("unknown running mod!")
	}
}
