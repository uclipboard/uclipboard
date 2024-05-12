package adapter

import (
	"bytes"
	"os/exec"
)

type WlClipboard struct {
}

func (WL *WlClipboard) Copy(s string) error {
	copyCmd := exec.Command("wl-copy")
	copyCmd.Stdin = bytes.NewBufferString(s)

	err := copyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (WL *WlClipboard) Paste() (string, error) {
	pasteCmd := exec.Command("wl-paste", "-n")
	var out bytes.Buffer
	pasteCmd.Stdout = &out

	err := pasteCmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func NewWl() *WlClipboard {
	return &WlClipboard{}
}
