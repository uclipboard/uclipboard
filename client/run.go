package client

import (
	"io"
	"net/http"
	"os"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func Run(c *model.Conf) {
	var clipboardAdapter model.ClipboardCmdAdapter
	logger = model.NewModuleLogger("client")
	switch c.Client.Adapter {
	case "wl":
		clipboardAdapter = adapter.NewWl()
	case "xc":
		clipboardAdapter = adapter.NewXClip()
	default:
		// win MacOS(pbcopy/paste)
		logger.Panic("error unknown clipboard adapter")
	}
	client := &http.Client{}
	mainLoop(c, clipboardAdapter, client, logger)
}

func Instant(c *model.Conf) {
	client := &http.Client{}

	argMsg := c.Run.Msg
	// TODO:Support binary file uploading
	// priority: argument message > stdin
	if argMsg == "" {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Panicf("Read data from stdin error: %s", err.Error())
		}

		if len(in) != 0 {
			UploadStringData(string(in), client, c)
		} else {
			logger.Warn("nothing readed")
			os.Exit(1)
		}
	} else if argMsg != "" {
		UploadStringData(argMsg, client, c)
	}

}
