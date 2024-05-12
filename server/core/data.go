package core

import (
	"fmt"

	"github.com/dangjinghao/uclipboard/model"
)

// ServerClipboard maintained by a coroutine named `ClipboardManager`
var serverData struct {
	actionChannel chan *model.ActionChannelItem
	clipboard     []model.ClipboardItem
}

// The Data returned  to gin request goroutine by ClipboardManager
var returnChannel chan *model.ReturnChannelItem

func pushClipboard(i *model.ClipboardItem) {
	serverData.clipboard = append(serverData.clipboard, *i)
}

func pullClipboard(n int) []model.ClipboardItem {
	// FIXME: empty in the initial
	return serverData.clipboard[len(serverData.clipboard)-n:]
}
func managerPushClipboard(act *model.ActionChannelItem) {
	ret := model.NewReturnChannelItem()
	ret.Ctx = act.Ctx
	pushClipboard(&act.Clipboard)
	returnChannel <- ret

}

func mamangerPullClipboard(act *model.ActionChannelItem) {
	ret := model.NewReturnChannelItem()
	ret.Ctx = act.Ctx
	ret.Clipboard = pullClipboard(1)[0]
	returnChannel <- ret
}

func clipboardManager() {
	for {
		act, open := <-serverData.actionChannel
		if !open {
			panic(model.ErrChanClosed)
		}
		switch act.Act {
		case model.ActCmdPushClipboard:
			managerPushClipboard(act)

		case model.ActCmdPullClipboard:
			mamangerPullClipboard(act)
		case model.ActCmdRemoveClipboard:
			fmt.Println("act remove!")
		}
	}
}

func PushActionChan(d *model.ActionChannelItem) {
	serverData.actionChannel <- d
}

func PullFromReturnChan() *model.ReturnChannelItem {
	d, ok := <-returnChannel
	if !ok {
		panic(model.ErrChanClosed)
	}
	return d
}

func init() {
	returnChannel = make(chan *model.ReturnChannelItem)
	serverData.actionChannel = make(chan *model.ActionChannelItem)
	serverData.clipboard = make([]model.ClipboardItem, 0)
	go clipboardManager()
}
