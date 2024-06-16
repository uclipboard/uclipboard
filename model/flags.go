package model

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func InitFlags(c *UContext) {
	// get the running program path
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	// modify the `Runtime` config struct
	flag.StringVar(&c.Runtime.Mode, "mode", "instant", "Specify the running mode. (client|server|instant)")
	flag.StringVar(&c.Runtime.LogLevel, "log-level", "info", "logger level [info/debug/trace]")
	flag.StringVar(&c.Runtime.ConfPath, "conf", fmt.Sprintf("%s/conf.toml", exPath), "Specify the config path.")
	flag.StringVar(&c.Runtime.LogPath, "log", "", "Specify the log path.")
	flag.BoolVar(&c.Runtime.ShowVersion, "version", false, "show version")
	flag.BoolVar(&c.Runtime.ShowHelp, "help", false, "show help")

	flag.StringVar(&c.Runtime.PushMsg, "push", "", "(instant mode) push clipboard data.")
	flag.StringVar(&c.Runtime.Upload, "upload", "", "(instant mode) upload what ever file you want.")
	flag.StringVar(&c.Runtime.Download, "download", "",
		"(instant mode) specify the file name you want to download."+
			"You can specify the file name to download the latest same name file."+
			"You can also specify file id by @id to download the file you want. e.g. -download @123")
	flag.StringVar(&c.Runtime.Test, "test", "", "componments test, [ctf] `c`: allow all cors request. `t`: disable token check `f`: disable frontend... multi-char is allowed. e.g. -test ct")
	flag.BoolVar(&c.Runtime.Pull, "pull", false, "(instant mode) pull clipboard data.")
	flag.BoolVar(&c.Runtime.Latest, "latest", false, "(instant mode) download latest file.")
	flag.StringVar(&c.Runtime.UploadFileLifetimeStr, "lifetime", "", "(instant mode) specify the file lifetime. (support unit: s, m, h, d)")
	flag.Parse()
}
