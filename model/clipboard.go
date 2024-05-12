package model

import "reflect"

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

func CmpClipboard(a *Clipboard, b *Clipboard) bool {
	return reflect.DeepEqual(a, b)
}
func IndexClipboardArray(arr []Clipboard, item *Clipboard) int {
	for index, arrItem := range arr {
		if CmpClipboard(&arrItem, item) {
			return index
		}
	}
	return -1
}
