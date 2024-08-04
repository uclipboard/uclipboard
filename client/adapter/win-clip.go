package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	ErrCodeAccessDenied             = 5
	ErrCodeClipboardEmpty           = -2
	ErrCodeClipboardDataTypeUnknown = -3
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

func parseStdErr(stdErrStr string) (int, string) {
	var errCode int
	fmt.Sscanf(stdErrStr, "[%d]", &errCode)
	errString := stdErrStr[strings.Index(stdErrStr, "]")+1:]
	return errCode, errString
}

func (WC *WinClip) Paste() (string, error) {
	pasteCmd := exec.Command("win-clip.exe", "paste", "-u")
	var stdOut bytes.Buffer
	pasteCmd.Stdout = &stdOut
	stdErr := bytes.NewBuffer(nil)
	pasteCmd.Stderr = stdErr
	err := pasteCmd.Run()
	if err != nil {
		stdErrStr := stdErr.String()
		errCode, errString := parseStdErr(stdErrStr)
		switch errCode {
		case ErrCodeClipboardEmpty:
			return "", ErrEmptyClipboard
		case ErrCodeAccessDenied:
			return "", ErrLockedClipboard
		case ErrCodeClipboardDataTypeUnknown:
			return "", ErrClipboardDataTypeUnknown
		default:
			if errString != "" {
				return "", errors.New(errString)
			} else {
				return "", err
			}
		}
	}
	outStr := stdOut.String()
	outStr = strings.ReplaceAll(outStr, "\r\n", "\n")
	return outStr, nil
}

func NewWinClip() *WinClip {
	return &WinClip{}
}
