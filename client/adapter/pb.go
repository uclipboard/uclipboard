package adapter

import (
	"bytes"
	"os/exec"
	"strings"
)

type PBClipboard struct {
}

func (PB *PBClipboard) Copy(s string) error {
	copyCmd := exec.Command("pbcopy")
	copyCmd.Stdin = bytes.NewBufferString(s)

	err := copyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}
func (PB *PBClipboard) Paste() (string, error) {
	pasteCmd := exec.Command("pbpaste")
	var out bytes.Buffer
	pasteCmd.Stdout = &out
	stdErr := bytes.NewBuffer(nil)
	pasteCmd.Stderr = stdErr

	// If system clipboard is empty, pbpaste will return exit code 1 with `Error: target STRING not available` in stdout
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
func NewPBClipboard() *PBClipboard {
	return &PBClipboard{}
}
