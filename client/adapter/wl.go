package adapter

import (
	"bytes"

	"github.com/uclipboard/uclipboard/model"
)

type WlClipboard struct {
}

func (WL *WlClipboard) Copy(s string) error {
	return defaultCopy("wl-copy")(s)
}

func (WL *WlClipboard) Paste() (string, error) {
	// first, check type
	checkCmd := neighborExec("wl-paste --list-types")
	// read stdout and check whether it contains "x-kde-lockscreen", that means the clipboard is locked
	checkTypeOut := bytes.NewBuffer(nil)
	checkCmd.Stdout = checkTypeOut

	if err := checkCmd.Run(); err != nil {
		return "", err
	}

	// In gnome(/debian), locking screen will clear the clipboard, but the clipboard type is not empty.
	// TODO:what will be happen in other DE or WM?
	// locked screen in kde will print "x-kde-lockscreen" in stderr
	if bytes.Contains(checkTypeOut.Bytes(), []byte("x-kde-lockscreen")) {
		return "", ErrLockedClipboard
	}

	pasteCmd := neighborExec("wl-paste -n")
	out := bytes.NewBuffer(nil)
	pasteCmd.Stdout = out

	if err := pasteCmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (WL *WlClipboard) Watch(onChange func(string)) error {
	firstTime := true
	logger := model.NewModuleLogger("WlWatch")
	return defaultWatch("wl-paste -n --watch %s", func(s string) {
		if firstTime {
			// when the wl-paste -n start, it will print the current clipboard content
			// so we need to ignore the first time
			firstTime = false
			logger.Debugf("ignore first time %s", s)
			return
		}
		onChange(s)
	})
}

func init() {
	RegisterFactory("wl", func(_ *model.UContext) ClipboardCmdAdapter {
		return &WlClipboard{}
	})
}
