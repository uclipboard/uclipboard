package adapter

import (
	"github.com/uclipboard/uclipboard/model"
)

type PBMClipboard struct {
}

func (PBM *PBMClipboard) Copy(s string) error {
	return defaultCopy("pbman copy")(s)
}
func (PBM *PBMClipboard) Paste() (string, error) {
	return defaultPaste("pbman paste")()
}

func (PBM *PBMClipboard) Watch(onChange func(string)) error {
	// pbman watch will print the content to stdout when it changed
	// we need to use a pipe to execute the command when the clipboard changed
	return defaultWatch("pbman watch %s", onChange)
}

func init() {
	RegisterFactory("pbm", func(_ *model.UContext) ClipboardCmdAdapter {
		return &PBMClipboard{}
	})
}
