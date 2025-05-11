package adapter

import "github.com/uclipboard/uclipboard/model"

type PBMClipboard struct {
}

func (pbm *PBMClipboard) Copy(s string) error {
	return defaultCopy("pbman copy")(s)
}
func (pbm *PBMClipboard) Paste() (string, error) {
	return defaultPaste("pbman paste")()
}

func init() {
	RegisterFactory("pbm", func(_ *model.UContext) ClipboardCmdAdapter {
		return &PBMClipboard{}
	})
}
