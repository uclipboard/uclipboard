package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dangjinghao/uclipboard/client/adapter"
	"github.com/dangjinghao/uclipboard/model"
)

func Run(c *model.Conf) {
	var clipboardAdapter model.ClipboardCmdAdapter

	switch c.Client.Adapter {
	case "wl":
		clipboardAdapter = adapter.NewWl()
	case "xc":
		clipboardAdapter = adapter.NewXClip()
	default:
		// win MacOS(pbcopy/paste)
		panic(model.ErrUnimplement)
	}
	client := &http.Client{}
	mainLoop(c, clipboardAdapter, client)
}

func Instant(c *model.Conf) {
	client := &http.Client{}

	argMsg := c.Run.Msg
	// TODO:Support binary file uploading
	// priority: argument message > stdin > download
	if argMsg == "" {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic("read data from stdin error:" + err.Error())
		}

		if len(in) != 0 {
			// upload stdin data
			reqBody, _ := GenClipboardReqBody(string(in))
			resp, err := client.Post(model.UrlPushApi(c),
				"application/json", bytes.NewReader(reqBody))

			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

		} else {
			fmt.Println("nothing readed")
			os.Exit(1)
		}
	} else if argMsg != "" {
		// uplad argument msg
		reqBody, _ := GenClipboardReqBody(argMsg)
		resp, err := client.Post(model.UrlPushApi(c),
			"application/json", bytes.NewReader(reqBody))

		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}

}
