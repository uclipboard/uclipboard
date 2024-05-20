package client

import (
	"encoding/json"
	"fmt"
	"io"
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
		clipboardAdapter = adapter.NewXClip(c.Client.XSelection)
	case "wc":
		clipboardAdapter = adapter.NewWinClip()
	default:
		// win MacOS(pbcopy/paste)
		logger.Fatal("error unknown clipboard adapter")
	}
	client := newUClipboardHttpClient()
	mainLoop(c, clipboardAdapter, client)
}

func Instant(c *model.Conf) {
	client := newUClipboardHttpClient()
	logger := model.NewModuleLogger("instant")
	argMsg := c.Runtime.Msg
	// priority: binary file > pull data > argument message > stdin

	if c.Runtime.Upload != "" {
		logger.Tracef("upload binary file: %s", c.Runtime.Upload)

		uploadFile(c.Runtime.Upload, client, c, logger)

	} else if c.Runtime.Latest || c.Runtime.Download != "" {
		logger.Tracef("download binary file. c.Flags.Latest:%t, c.Flags.Download:%s", c.Runtime.Latest, c.Runtime.Download)

		var fileName string
		if c.Runtime.Latest {
			fileName = ""
		} else {
			fileName = c.Runtime.Download
		}
		downloadFile(fileName, client, c, logger)

	} else if c.Runtime.Pull {
		logger.Trace("pull clipboard from server")

		var clipboardArr []model.Clipboard
		resp, err := pullStringData(client, c, logger)
		if err != nil {
			logger.Fatalf("cannot pull data  from server: %s", err.Error())
		}
		logger.Tracef("resp: %s", resp)
		if err = json.Unmarshal(resp, &clipboardArr); err != nil {
			logger.Fatalf("cannot parse response body: %s", err.Error())
		}
		newContent := deteckAndconcatClipboardFileUrl(c, &clipboardArr[0])
		logger.Tracef("newContent: %s", newContent)
		fmt.Println(newContent)

	} else if argMsg == "" {
		logger.Trace("read data from stdin because there is no argument message")
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Fatalf("Read data from stdin error: %s", err.Error())
		}

		if len(in) != 0 {
			uploadStringData(string(in), client, c, logger)
		} else {
			logger.Fatal("nothing readed")
		}

	} else if argMsg != "" {
		logger.Tracef("upload argument message: %s", argMsg)
		uploadStringData(argMsg, client, c, logger)
	}

}
