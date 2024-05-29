package model

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func InitFlags(conf *Conf) {
	// get the running program path
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	// modify the `Runtime` config struct
	flag.StringVar(&conf.Runtime.Mode, "mode", "instant", "Specify the running mode. (client|server|instant)")
	flag.StringVar(&conf.Runtime.LogLevel, "log-level", "info", "logger level [info/debug/trace]")
	flag.StringVar(&conf.Runtime.ConfPath, "conf", fmt.Sprintf("%s/conf.toml", exPath), "Specify the config path.")
	flag.StringVar(&conf.Runtime.Msg, "msg", "", "(instant mode) push clipboard data.")
	flag.StringVar(&conf.Runtime.Upload, "upload", "", "(instant mode) upload what ever file you want.")
	flag.StringVar(&conf.Runtime.LogPath, "log", "", "Specify the log path.")
	flag.StringVar(&conf.Runtime.Download, "download", "",
		"(instant mode) specify the file name you want to download."+
			"You can specify the file name to download the latest file."+
			"You can also specify file id by @id to download the file you want. e.g. -download @123")
	flag.StringVar(&conf.Runtime.Test, "test", "", "componments test, [ct] `c`: allow all cors request. `t`: disable token check ... multi-char is allowed. e.g. -test ct")
	flag.BoolVar(&conf.Runtime.Pull, "pull", false, "(instant mode) pull clipboard data.")
	flag.BoolVar(&conf.Runtime.Latest, "latest", false, "(instant mode) download latest file.")
	flag.BoolVar(&conf.Runtime.ShowVersion, "version", false, "show version")
	flag.Parse()
}
