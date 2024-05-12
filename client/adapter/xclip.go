package adapter

import (
	"bytes"
	"os/exec"
)

type XClipClipboard struct {
}

func (XC *XClipClipboard) Copy(s string) error {
	copyCmd := exec.Command("xclip", "-selection", "clipboard")
	copyCmd.Stdin = bytes.NewBufferString(s)

	err := copyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (XC *XClipClipboard) Paste() (string, error) {
	pasteCmd := exec.Command("xclip", "-o")
	var out bytes.Buffer
	pasteCmd.Stdout = &out

	err := pasteCmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func NewXClip() *XClipClipboard {
	return &XClipClipboard{}
}
