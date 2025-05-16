package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/uclipboard/uclipboard/model"
)

const (
	ErrCodeAccessDenied             = 5
	ErrCodeClipboardEmpty           = -2
	ErrCodeClipboardDataTypeUnknown = -3
)

type WinClip struct {
}

func (WC *WinClip) Copy(s string) error {
	s = strings.ReplaceAll(s, "\n", "\r\n")
	return defaultCopy("win-clip.exe copy -u")(s)
}

func parseStdErr(stdErrStr string) (int, string) {
	errCode := 0
	fmt.Sscanf(stdErrStr, "[%d]", &errCode)
	errString := stdErrStr[strings.Index(stdErrStr, "]")+1:]
	return errCode, errString
}

func (WC *WinClip) Paste() (string, error) {
	pasteCmd := neighborExec("win-clip.exe paste -u")
	stdOut := bytes.NewBuffer(nil)
	pasteCmd.Stdout = stdOut
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

func (WC *WinClip) Watch(f func(string)) error {
	newlineReplaceWrapper := func(s string) {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		f(s)
	}
	return defaultWatch("win-clip.exe paste -u -w %s", newlineReplaceWrapper)
}

func init() {
	RegisterFactory("wc", func(_ *model.UContext) ClipboardCmdAdapter {
		return &WinClip{}
	})
}
