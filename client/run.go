package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func Run(c *model.Conf) {
	var clipboardAdapter model.ClipboardCmdAdapter
	logger := model.NewModuleLogger("client")
	switch c.Client.Adapter {
	case "wl":
		clipboardAdapter = adapter.NewWl()
	case "xc":
		clipboardAdapter = adapter.NewXClip()
	case "wc":
		clipboardAdapter = adapter.NewWinClip()
	default:
		// win MacOS(pbcopy/paste)
		logger.Panic("error unknown clipboard adapter")
	}
	client := &http.Client{}
	mainLoop(c, clipboardAdapter, client)
}

func Instant(c *model.Conf) {
	client := &http.Client{}
	logger := model.NewModuleLogger("instant")
	argMsg := c.Run.Msg
	// TODO:Support binary file uploading
	// priority: pull data > argument message > stdin
	if c.Run.Pull {
		var clipboardArr []model.Clipboard
		resp, err := PullStringData(client, c, logger)
		if err != nil {
			logger.Panicf("PullStringData error:%s", err.Error())
		}
		logger.Tracef("resp:%s", resp)
		if err = json.Unmarshal(resp, &clipboardArr); err != nil {
			logger.Panicf("cannot parse response body: %s", err.Error())
		}

		fmt.Print(clipboardArr[0].Content)

	} else if argMsg == "" {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Panicf("Read data from stdin error: %s", err.Error())
		}

		if len(in) != 0 {
			UploadStringData(string(in), client, c, logger)
		} else {
			logger.Warn("nothing readed")
			os.Exit(1)
		}
	} else if argMsg != "" {
		UploadStringData(argMsg, client, c, logger)
	}

}
