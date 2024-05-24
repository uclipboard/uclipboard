package client

import (
	"fmt"
	"io"
	"os"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func Run(c *model.Conf) {
	var clipboardAdapter adapter.ClipboardCmdAdapter
	logger := model.NewModuleLogger("client")
	switch c.Client.Adapter {
	case "wl":
		clipboardAdapter = adapter.NewWl()
	case "xc":
		clipboardAdapter = adapter.NewXClip(c.Client.XSelection)
	case "wc":
		clipboardAdapter = adapter.NewWinClip()
	default:
		// MacOS(pbcopy/paste)
		logger.Fatal("error unknown clipboard adapter")
	}
	client := NewUClipboardHttpClient()
	mainLoop(c, clipboardAdapter, client)
}

func Instant(c *model.Conf) {
	client := NewUClipboardHttpClient()
	logger := model.NewModuleLogger("instant")
	argMsg := c.Runtime.Msg
	// priority: binary file > pull data > argument message > stdin

	if c.Runtime.Upload != "" {
		logger.Tracef("upload binary file: %s", c.Runtime.Upload)

		UploadFile(c.Runtime.Upload, client, c, logger)

	} else if c.Runtime.Latest || c.Runtime.Download != "" {
		logger.Tracef("download binary file. c.Flags.Latest:%t, c.Flags.Download:%s", c.Runtime.Latest, c.Runtime.Download)

		var fileName string
		if c.Runtime.Latest {
			fileName = ""
		} else {
			fileName = c.Runtime.Download
		}
		DownloadFile(fileName, client, c, logger)

	} else if c.Runtime.Pull {
		logger.Trace("pull clipboard from server")

		respBody, err := SendPullReq(client, c)
		if err != nil {
			logger.Fatalf("cannot pull data from server: %s", err.Error())
		}
		logger.Tracef("respBody: %s", respBody)
		clipboardArr, err := ParsePullData(respBody)
		if err != nil {
			logger.Fatalf("parse pull data error: %v", err)
		}
		newContent := DeteckAndConcatFileUrl(c, &clipboardArr[0])
		logger.Tracef("newContent: %s", newContent)
		fmt.Println(newContent)

	} else if argMsg == "" {
		logger.Trace("read data from stdin because there is no argument message")
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Fatalf("Read data from stdin error: %s", err.Error())
		}

		if len(in) != 0 {
			if err := SendPushReq(string(in), client, c); err != nil {
				logger.Fatalf("send push request error: %v", err)
			}
		} else {
			logger.Fatal("nothing readed")
		}

	} else if argMsg != "" {
		logger.Tracef("upload argument message: %s", argMsg)
		if err := SendPushReq(argMsg, client, c); err != nil {
			logger.Fatalf("send push request error:%v", err)
		}
	}

}
