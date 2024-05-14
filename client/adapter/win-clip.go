package adapter

import (
	"bytes"
	"os/exec"
)

type WinClip struct {
}

func (WC *WinClip) Copy(s string) error {
	copyCmd := exec.Command("./win-clip.exe", "copy", "-u")
	copyCmd.Stdin = bytes.NewBufferString(s)

	err := copyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (WC *WinClip) Paste() (string, error) {
	pasteCmd := exec.Command("./win-clip.exe", "paste", "-u")
	var out bytes.Buffer
	pasteCmd.Stdout = &out

	err := pasteCmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func NewWinClip() *WinClip {
	return &WinClip{}
}
