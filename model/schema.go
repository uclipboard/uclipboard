package model

import (
	"encoding/json"
	"os"
	"reflect"
	"time"
)

type Clipboard struct {
	Id          int64  `json:"id" db:"id"`
	Ts          int64  `json:"ts" db:"ts"` //ms timestamp
	Content     string `json:"content" db:"content"`
	Hostname    string `json:"hostname" db:"hostname"`         // sender
	ContentType string `json:"content_type" db:"content_type"` //
}

func NewClipoardWithDefault() *Clipboard {
	// I don't know why it doesn't support default value
	return &Clipboard{Hostname: "unknown", ContentType: "text", Ts: time.Now().UnixMilli()}
}

// It generates the hostname so take care of where it is called
func NewFullClipoard(c string) *Clipboard {
	data := NewClipoardWithDefault()
	data.Content = c
	hostname, err := os.Hostname()
	if err != nil {
		logger.Warnf("Can't get hostname:%v", err)
	} else {
		data.Hostname = hostname
	}
	return data
}

func CmpClipboard(a *Clipboard, b *Clipboard) bool {
	// ignore id comparison
	prevId := a.Id
	a.Id = b.Id
	result := reflect.DeepEqual(a, b)
	a.Id = prevId
	return result
}
func IndexClipboardArray(arr []Clipboard, item *Clipboard) int {
	for index, arrItem := range arr {
		if CmpClipboard(&arrItem, item) {
			return index
		}
	}
	return -1
}

type FileMetadata struct {
	Id        int64  `json:"-" db:"id"`
	CreatedTs int64  `json:"created_ts" db:"created_ts"` //ms timestamp
	ExpireTs  int64  `json:"expire_ts" db:"expire_ts"`   //ms timestamp
	FileName  string `json:"file_name" db:"file_name"`
	TmpPath   string `json:"tmp_path" db:"tmp_path"` //relative path based on the tmpPath in conf
}

func NewFileMetadataWithDefault() *FileMetadata {
	return &FileMetadata{
		CreatedTs: time.Now().UnixMilli(),
	}
}

type ServerResponse struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func NewDefaultServeRes(msg string, data []byte) *ServerResponse {
	if msg == "" {
		msg = "ok"
	}
	return &ServerResponse{
		Msg:  msg,
		Data: data,
	}
}
