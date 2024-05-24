package adapter

import "errors"

type ClipboardCmdAdapter interface {
	Copy(s string) error
	Paste() (string, error)
}

var (
	ErrEmptyClipboard = errors.New("adapter exits with 1, perhaps system clipboard is empty")
)
