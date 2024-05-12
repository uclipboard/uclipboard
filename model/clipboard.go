package model

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

type ClipboardItem struct {
	No           uint64 //should be empty when push clipboard item
	Ts           uint64 `json:"ts"` //ms timestamp
	Content      string `json:"content"`
	FromHostname string `json:"hostname"` // sender
}

// Thoese action command is used to ask ClipboardManager to do sth
const (
	ActCmdPushClipboard int = iota
	ActCmdPushManyClipboards
	ActCmdRemoveClipboard
	ActCmdPullClipboard
)

// Used to commucation between gin serve goroutine and ClipboardManager goroutine
type ActionChannelItem struct {
	Act       int
	Clipboard ClipboardItem
	Ctx       *gin.Context
}

type ReturnChannelItem struct {
	Err       error
	Clipboard ClipboardItem
	Ctx       *gin.Context
}

func (cbi *ClipboardItem) Cmp(another *ClipboardItem) bool {
	return reflect.DeepEqual(cbi, another)
}

func NewClipoardItem() *ClipboardItem {
	return &ClipboardItem{}
}

func NewClipboardAction() *ActionChannelItem {

	return &ActionChannelItem{}
}

func NewReturnChannelItem() *ReturnChannelItem {
	return &ReturnChannelItem{}
}
