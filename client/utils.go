package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dangjinghao/uclipboard/model"
)

var ErrUnexpectRespStatus = errors.New("this response status code isn't ok")

// return the marshaled request body and the raw clipboard struct
func GenClipboardReqBody(c string) ([]byte, *model.Clipboard, error) {
	reqData := model.NewFullClipoard(c)
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, nil, err
	}
	return reqBody, reqData, nil
}

func SendPushReq(s string, client *http.Client, c *model.Conf) error {
	reqBody, _, err := GenClipboardReqBody(s)
	if err != nil {
		return err
	}
	resp, err := client.Post(model.UrlPushApi(c),
		"application/json", bytes.NewReader(reqBody))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrUnexpectRespStatus
	}

	return nil
}

func SendPullReq(client *http.Client, c *model.Conf) ([]byte, error) {
	pullApi := model.UrlPullApi(c)
	resp, err := client.Get(pullApi)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnexpectRespStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return body, nil
}

func NewUClipboardHttpClient() *http.Client {
	return &http.Client{}
}

// if this clipboard content is a binary file, return the download url
func DeteckAndConcatFileUrl(conf *model.Conf, clipboard *model.Clipboard) string {
	content := clipboard.Content
	if clipboard.ContentType == "binary" {
		content = model.UrlDownloadApi(conf, content)
	}
	return content
}

func ExtractErrorMsg(body []byte) (string, error) {
	var respModel model.ServerResponse
	json.Unmarshal(body, &respModel)
	return respModel.Msg, nil
}

func ParsePullData(body []byte) (remoteClipboards []model.Clipboard, err error) {
	var bodyJson model.ServerResponse
	if err = json.Unmarshal(body, &bodyJson); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyJson.Data, &remoteClipboards); err != nil {
		return nil, err
	}

	return
}

func ParseUploadInfomation(body []byte) (info map[string]interface{}, err error) {
	var bodyJson model.ServerResponse
	if err = json.Unmarshal(body, &bodyJson); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyJson.Data, &info); err != nil {
		return nil, err
	}

	return

}
