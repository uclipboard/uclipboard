package adapter

import (
	"bytes"
	"os/exec"

	"github.com/uclipboard/uclipboard/model"
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
	// first, check type
	checkCmd := exec.Command("wl-paste", "--list-types")
	// read stdout and check whether it contains "x-kde-lockscreen", that means the clipboard is locked
	var checkTypeOut bytes.Buffer
	checkCmd.Stdout = &checkTypeOut

	err := checkCmd.Run()
	if err != nil {
		return "", err
	}

	// In gnome(/debian), locking screen will clear the clipboard, but the clipboard type is not empty.
	// TODO:what will be happen in other DE or WM?
	// locked screen in kde will print "x-kde-lockscreen" in stderr
	if bytes.Contains(checkTypeOut.Bytes(), []byte("x-kde-lockscreen")) {
		return "", ErrLockedClipboard
	}

	pasteCmd := exec.Command("wl-paste", "-n")
	var out bytes.Buffer
	pasteCmd.Stdout = &out

	err = pasteCmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func init() {
	RegisterFactory("wl", func(_ *model.UContext) ClipboardCmdAdapter {
		return &WlClipboard{}
	})
}
