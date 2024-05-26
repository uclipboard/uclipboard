package main

import (
	"flag"
	"strings"

	"github.com/dangjinghao/uclipboard/client"
	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server"
)

func main() {
	// create default config struct
	conf := model.NewConfWithDefault()
	// modify the `run` config struct
	flag.StringVar(&conf.Runtime.Mode, "mode", "instant", "Specify the running mode. (client|server|instant)")
	flag.StringVar(&conf.Runtime.LogLevel, "log-level", "info", "logger level [info/debug/trace]")
	flag.StringVar(&conf.Runtime.ConfPath, "conf", "./conf.toml", "Specify the config path.")
	flag.StringVar(&conf.Runtime.Msg, "msg", "", "(instant mode) push clipboard data.")
	flag.StringVar(&conf.Runtime.Upload, "upload", "", "(instant mode) upload what ever file you want.")
	flag.StringVar(&conf.Runtime.Download, "download", "",
		"(instant mode) specify the file name you want to download."+
			"You can specify the file name to download the latest file."+
			"You can also specify file id by @id to download the file you want. e.g. -download @123")
	flag.BoolVar(&conf.Runtime.Pull, "pull", false, "(instant mode) pull clipboard data.")
	flag.BoolVar(&conf.Runtime.Latest, "latest", false, "(instant mode) download latest file.")
	flag.StringVar(&conf.Runtime.Test, "test", "", "componments test, [ct] `c`: allow all cors request. `t`: disable token check ... multi-char is allowed. e.g. -test ct")

	flag.Parse()
	model.LoggerInit(conf.Runtime.LogLevel)

	logger := model.NewModuleLogger("entry")

	// modify config struct by conf file
	conf = model.LoadConf(conf)
	conf = model.FormatConf(conf)

	if conf.Token == "" && !strings.Contains(conf.Runtime.Test, "t") {
		logger.Fatal("token is empty, please set token in conf file.")
	}

	conf.Runtime.TokenEncrypt = model.TokenEncrypt(conf.Token)

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
