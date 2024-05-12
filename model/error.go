package model

import (
	"errors"
)

// TODO: custom errors
var (
	ErrUnimplement = errors.New("unimplement")
	ErrOpenFailed  = errors.New("open file failed")
	ErrSetDefault  = errors.New("canot set the default value")
)

var (
	ErrClipboardPushFailed = errors.New("push clipboard failed")
)
