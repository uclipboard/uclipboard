package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/uclipboard/uclipboard/model"
)

var ErrUnexpectRespStatus = errors.New("this response status code isn't ok")

// return the marshaled request body and the raw clipboard struct
func GenClipboardReqBody(c string) ([]byte, *model.Clipboard, error) {
	reqData := model.NewFullClipoard(c)

	if reqData.Hostname == "unknown" {
		reqData.Hostname = "uclipboard_client"
	}

	reqByte, err := json.Marshal(reqData)
	if err != nil {
		return nil, nil, err
	}
	return reqByte, reqData, nil
}

func SendPushReq(s string, client *http.Client, c *model.UContext) (*model.Clipboard, error) {
	reqBody, clipboardInstance, err := GenClipboardReqBody(s)
	if err != nil {
		return nil, err
	}
	resp, err := client.Post(model.UrlPushApi(c),
		"application/json", bytes.NewReader(reqBody))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnexpectRespStatus
	}

	return clipboardInstance, nil
}

func SendPullReq(client *http.Client, c *model.UContext) ([]byte, error) {
	pullApi := model.UrlPullApi(c)
	resp, err := client.Get(pullApi)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = ErrUnexpectRespStatus
	}

	return body, err
}

func SendWebSocketPush(s string, wso *model.WsObject) (*model.Clipboard, error) {
	_, clipboardInstance, err := GenClipboardReqBody(s)
	if err != nil {
		return nil, err
	}
	wsMsg := model.WSRequestPushMessage{
		Type:      model.WSMsgTypePush,
		Clipboard: *clipboardInstance,
	}
	if err := wso.WriteJSON(wsMsg); err != nil {
		return nil, err
	}
	return clipboardInstance, nil
}

func SendWebSocketPull(wso *model.WsObject) error {
	wsMsg := model.WSBaseMessage{
		Type: model.WSMsgTypePull,
	}
	return wso.WriteJSON(wsMsg)
}

func CreateWsConn(c *model.UContext) (*model.WsObject, error) {
	wsApi := model.UrlWsApi(c)
	conn, _, err := websocket.DefaultDialer.Dial(wsApi, nil)
	if err != nil {
		return nil, err
	}
	return model.NewWsObject(conn, websocket.DefaultDialer, wsApi), nil
}

func NewUClipboardHttpClient(c *model.UContext) *http.Client {
	if c.Runtime.Mode == "instant" {
		return &http.Client{Timeout: time.Duration(c.Client.Connect.UploadTimeout) * time.Second}
	}
	// else
	client := &http.Client{Timeout: time.Duration(c.Client.Connect.Timeout) * time.Millisecond}
	return client
}

// if this clipboard content is a binary file, return the download url,
// else return the raw content itself
func DetectAndConcatFileUrl(conf *model.UContext, clipboard *model.Clipboard) string {
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

func ParseUploadInfomation(body []byte) (*model.FileMetadataResponse, error) {
	var bodyJson model.ServerResponse
	if err := json.Unmarshal(body, &bodyJson); err != nil {
		return nil, err
	}
	info := model.FileMetadataResponse{}
	if err := json.Unmarshal(bodyJson.Data, &info); err != nil {
		return nil, err
	}

	return &info, nil

}
