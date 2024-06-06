package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const ErrCodeAccessDenied = 5

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

		if strings.Contains(errString, "Clipboard is empty") {
			return "", ErrEmptyClipboard
		} else if errCode == ErrCodeAccessDenied {
			return "", ErrLockedClipboard
		}
		// else
		return "", errors.New(errString)
	}
	outStr := stdOut.String()
	outStr = strings.ReplaceAll(outStr, "\r\n", "\n")
	return outStr, nil
}

func NewWinClip() *WinClip {
	return &WinClip{}
}
