package client

import (
	"fmt"
	"io"
	"os"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func Run(c *model.Conf) {
	var clipboardAdapter model.ClipboardCmdAdapter

	switch c.Client_Adapter {
	case "wl":
		clipboardAdapter = adapter.NewWl()
	default:
		// X win MacOS(pbcopy/paste)
		panic(model.ErrUnimplement)
	}

	mainLoop(c, clipboardAdapter)
}

func Instant(c *model.Conf, argMsg string) {

	// TODO:Support binary file uploading
	// priority: argument message > stdin > download
	if argMsg == "" {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic("read data from stdin error:" + err.Error())
		}

		if len(in) != 0 {
			fmt.Println("upload stdin data:" + string(in))
		} else {
			fmt.Println("nothing")
			os.Exit(1)
		}
	} else if argMsg != "" {
		fmt.Println("upload argMsg data:" + string(argMsg))
	} else {
		fmt.Println("download data!")

	}

}
