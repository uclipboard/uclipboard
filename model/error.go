package model

import (
	"errors"
)

// TODO: custom errors
var (
	ErrUnimplement = errors.New("unimplement")
	ErrChanClosed  = errors.New("this channel should not be closed")
)

var (
	ErrClipboardPushFailed = errors.New("push clipboard failed")
)
