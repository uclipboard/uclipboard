package adapter

type PBMClipboard struct {
}

func (pbm *PBMClipboard) Copy(s string) error {
	return defaultCopy("pbman copy")(s)
}
func (pbm *PBMClipboard) Paste() (string, error) {
	return defaultPaste("pbman paste")()
}
func NewPBMClipboard() *PBMClipboard {
	return &PBMClipboard{}
}
