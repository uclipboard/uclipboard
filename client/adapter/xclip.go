package adapter

import (
	"bytes"
	"os/exec"
	"strings"
)

type XClipClipboard struct {
	selection string
}

func (XC *XClipClipboard) Copy(s string) error {
	copyCmd := exec.Command("xclip", "-selection", XC.selection)
	copyCmd.Stdin = bytes.NewBufferString(s)

	err := copyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (XC *XClipClipboard) Paste() (string, error) {
	pasteCmd := exec.Command("xclip", "-selection", XC.selection, "-o")
	var out bytes.Buffer
	pasteCmd.Stdout = &out
	stdErr := bytes.NewBuffer(nil)
	pasteCmd.Stderr = stdErr

	// If system clipboard is empty, xclip will return exit code 1 with `Error: target STRING not available` in stdout
	err := pasteCmd.Run()
	if err != nil {
		if strings.Contains(stdErr.String(), "target STRING not available") {
			return "", ErrEmptyClipboard
		} else {
			return "", err
		}
	}
	outputStr := out.String()
	return outputStr, nil
}

func NewXClip(selection string) *XClipClipboard {
	return &XClipClipboard{selection: selection}
}
