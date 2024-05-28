package adapter

import (
	"bytes"
	"os/exec"
	"strings"
)

type WinClip struct {
}

func (WC *WinClip) Copy(s string) error {
	copyCmd := exec.Command("win-clip.exe", "copy", "-u")
	s = strings.ReplaceAll(s, "\n", "\r\n")
	copyCmd.Stdin = bytes.NewBufferString(s)

	err := copyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (WC *WinClip) Paste() (string, error) {
	pasteCmd := exec.Command("win-clip.exe", "paste", "-u")
	var out bytes.Buffer
	pasteCmd.Stdout = &out
	stdErr := bytes.NewBuffer(nil)
	pasteCmd.Stderr = stdErr
	err := pasteCmd.Run()
	if err != nil {
		if strings.Contains(stdErr.String(), "no data") {
			return "", ErrEmptyClipboard
		}
		return "", err
	}

	outStr := out.String()
	outStr = strings.ReplaceAll(outStr, "\r\n", "\n")
	return outStr, nil
}

func NewWinClip() *WinClip {
	return &WinClip{}
}
