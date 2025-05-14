package adapter

import (
	"bytes"
	"strings"

	"github.com/uclipboard/uclipboard/model"
)

type XClipClipboard struct {
	selection string
}

func (XC *XClipClipboard) Copy(s string) error {
	return defaultCopy("xclip -selection " + XC.selection)(s)
}

func (XC *XClipClipboard) Paste() (string, error) {
	pasteCmd := neighborExec("xclip -o -selection " + XC.selection)
	out := bytes.NewBuffer(nil)
	pasteCmd.Stdout = out
	stdErr := bytes.NewBuffer(nil)
	pasteCmd.Stderr = stdErr

	// If system clipboard is empty, xclip will return exit code 1 with `Error: target STRING not available` in stdout
	if err := pasteCmd.Run(); err != nil {
		if strings.Contains(stdErr.String(), "target STRING not available") {
			return "", ErrEmptyClipboard
		} else {
			return "", err
		}
	}
	return out.String(), nil
}

func (XC *XClipClipboard) Watch(_ func(string)) error {
	// xclip doesn't support watch
	return ErrUnsupportedWatchMode
}

func init() {
	RegisterFactory("xc", func(u *model.UContext) ClipboardCmdAdapter {
		return &XClipClipboard{selection: u.Client.Adapter.XSelection}
	})
}
