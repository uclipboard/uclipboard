package client

import (
	"fmt"
	"time"

	"github.com/dangjinghao/uclipboard/model"
)

func mainLoop(cfg *model.Conf, adapter model.ClipboardCmdAdapter) {
	for {
		s, _ := adapter.Paste()
		fmt.Printf("\r%s", s)
		time.Sleep(time.Duration(cfg.Client_Interval) * time.Millisecond)
	}

}
