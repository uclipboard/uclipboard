package model

type ClipboardCmdAdapter interface {
	Copy(s string) error
	Paste() (string, error)
}
