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
		logger.Panic("error unknown clipboard adapter")
	}
	client := newUClipboardHttpClient()
	mainLoop(c, clipboardAdapter, client)
}

func Instant(c *model.Conf) {
	client := newUClipboardHttpClient()
	logger := model.NewModuleLogger("instant")
	argMsg := c.Flags.Msg
	// priority: binary file > pull data > argument message > stdin

	if c.Flags.Upload != "" {
		fmt.Print(uploadFile(c.Flags.Upload, client, c, logger))

	} else if c.Flags.Latest || c.Flags.Download != "" {
		var fileName string
		if c.Flags.Latest {
			fileName = ""
		} else {
			fileName = c.Flags.Download
		}
		downloadFile(fileName, client, c, logger)
		// panic when download file failed
		fmt.Printf("Download file success: %s\n", fileName)

	} else if c.Flags.Pull {
		var clipboardArr []model.Clipboard
		resp, err := pullStringData(client, c, logger)
		if err != nil {
			logger.Panicf("PullStringData error:%s", err.Error())
		}
		logger.Tracef("resp:%s", resp)
		if err = json.Unmarshal(resp, &clipboardArr); err != nil {
			logger.Panicf("cannot parse response body: %s", err.Error())
		}
		if clipboardArr[0].ContentType == "binary" {
			fmt.Print(model.UrlDownloadApi(c, clipboardArr[0].Content))
		} else {
			fmt.Print(clipboardArr[0].Content)
		}

	} else if argMsg == "" {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Panicf("Read data from stdin error: %s", err.Error())
		}

		if len(in) != 0 {
			uploadStringData(string(in), client, c, logger)
		} else {
			logger.Warn("nothing readed")
			os.Exit(1)
		}
	} else if argMsg != "" {
		uploadStringData(argMsg, client, c, logger)
	}

}
