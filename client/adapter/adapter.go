package adapter

import "errors"

type ClipboardCmdAdapter interface {
	Copy(s string) error
	Paste() (string, error)
}

var (
	ErrEmptyClipboard  = errors.New("perhaps system clipboard is empty")
	ErrLockedClipboard = errors.New("perhaps system clipboard is locked so that adapter can't access it")
)
