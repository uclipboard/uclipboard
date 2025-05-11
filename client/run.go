package client

import (
	"fmt"
	"io"
	"os"

	"github.com/uclipboard/uclipboard/client/adapter"
	"github.com/uclipboard/uclipboard/model"
)

func Run(c *model.UContext) {
	var clipboardAdapter adapter.ClipboardCmdAdapter
	logger := model.NewModuleLogger("client")
	switch c.Client.Adapter.Type {
	case "wl":
		clipboardAdapter = adapter.NewWl()
	case "xc":
		clipboardAdapter = adapter.NewXClip(c.Client.Adapter.XSelection)
	case "wc":
		clipboardAdapter = adapter.NewWinClip()
	case "pbm":
		clipboardAdapter = adapter.NewPBMClipboard()
	default:
		// MacOS(pbcopy/paste)
		logger.Fatal("error unknown clipboard adapter")
	}
	client := NewUClipboardHttpClient(c)
	mainLoop(c, clipboardAdapter, client)
}

func Instant(c *model.UContext) {
	client := NewUClipboardHttpClient(c)
	logger := model.NewModuleLogger("instant")
	argMsg := c.Runtime.PushMsg
	// priority: binary file > pull data > argument message > stdin

	if c.Runtime.Upload != "" {
		logger.Debugf("upload binary file: %s", c.Runtime.Upload)

		UploadFile(c.Runtime.Upload, client, c, logger)

	} else if c.Runtime.Latest || c.Runtime.Download != "" {
		logger.Debugf("download binary file. c.Flags.Latest:%t, c.Flags.Download:%s", c.Runtime.Latest, c.Runtime.Download)

		var fileName string
		if c.Runtime.Latest {
			fileName = ""
		} else {
			fileName = c.Runtime.Download
		}
		DownloadFile(fileName, client, c, logger)

	} else if c.Runtime.Pull {
		logger.Debugf("pull clipboard from server")

		respBody, err := SendPullReq(client, c)
		if err != nil {
			logger.Fatalf("cannot pull data from server: %s", err.Error())
		}
		logger.Tracef("respBody: %s", respBody)
		clipboardArr, err := ParsePullData(respBody)
		if err != nil {
			logger.Fatalf("parse pull data error: %v", err)
		}
		newContent := DetectAndConcatFileUrl(c, &clipboardArr[0])
		logger.Tracef("newContent: %s", newContent)
		fmt.Println(newContent)

	} else if argMsg == "" {
		logger.Debug("read data from stdin because there is no argument message")
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Fatalf("Read data from stdin error: %s", err.Error())
		}

		if len(in) == 0 {
			logger.Fatal("nothing readed")
		} else if len(in) > c.ContentLengthLimit {
			logger.Fatalf("stdin data size is too large, skip push")
		}

		if _, err := SendPushReq(string(in), client, c); err != nil {
			logger.Fatalf("send push request error: %v", err)
		}

	} else if argMsg != "" {
		if len(argMsg) > c.ContentLengthLimit {
			logger.Fatalf("argument message size is too large, skip push")
		}
		logger.Debugf("upload argument message: %s", argMsg)
		if _, err := SendPushReq(argMsg, client, c); err != nil {
			logger.Fatalf("send push request error:%v", err)
		}
	}

}
