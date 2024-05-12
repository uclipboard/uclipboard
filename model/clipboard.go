package model

type Clipboard struct {
	Id          int64  `json:"-" db:"id"`
	Ts          int64  `json:"ts" db:"ts"` //ms timestamp
	Content     string `json:"content" db:"content"`
	Hostname    string `json:"hostname" db:"hostname"`         // sender
	ContentType string `json:"content_type" db:"content_type"` //
}

func NewClipoardWithDefault() *Clipboard {
	// I don't know why it doesn't support default value
	return &Clipboard{Hostname: "unknown", ContentType: "text"}
}
