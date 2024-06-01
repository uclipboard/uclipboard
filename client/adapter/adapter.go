package adapter

import "errors"

type ClipboardCmdAdapter interface {
	Copy(s string) error
	Paste() (string, error)
}

var (
	ErrEmptyClipboard  = errors.New("adapter errors, perhaps system clipboard is empty")
	ErrLockedClipboard = errors.New("adapter errors, perhaps system clipboard is locked so that cannot access it")
)
