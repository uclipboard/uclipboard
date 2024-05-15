package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
	"github.com/gin-gonic/gin"
)

func HandlerPush(conf *model.Conf) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		clipboardData := model.NewClipoardWithDefault()
		if err := ctx.BindJSON(&clipboardData); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server interal Error"})
			return
		}
		if err := core.AddClipboardRecord(clipboardData); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server interal Error"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}
func HandlerPull(conf *model.Conf) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		var clipboardArr []model.Clipboard
		if err := core.GetLatestClipboardRecord(&clipboardArr, conf.Server.PullHistorySize); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
		}
		ctx.JSON(http.StatusOK, clipboardArr)
	}
}

func HandlerUpload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerUpload")
	return func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server interal Error:" + err.Error()})
			return
		}
		fileMetadata := model.NewFileMetadataWithDefault()
		fileMetadata.FileName = file.Filename
		fileMetadata.TmpPath = fmt.Sprintf("%s_%s", strconv.FormatInt(fileMetadata.CreatedTs, 10), file.Filename)
		fileMetadata.ExpireTs = conf.Server.DefaultFileLife + fileMetadata.CreatedTs
		logger.Tracef("fileMetadata: %v", fileMetadata)

		newClipboardRecord := model.NewClipoardWithDefault()
		// TODO: support hostname
		// ctx.FormValue("hostname")

		newClipboardRecord.ContentType = file.Header.Get("Content-Type")
		logger.Tracef("newClipboardRecord.ContentType(raw): %v", newClipboardRecord.ContentType)
		if newClipboardRecord.ContentType == "" {
			newClipboardRecord.ContentType = "application/octet-stream"
		}
		// save file to tmp directory and get the path to save in db
		filePath := filepath.Join(conf.Server.TmpPath, fileMetadata.TmpPath)
		newClipboardRecord.Content = filePath
		logger.Tracef("newClipboardRecord: %v", newClipboardRecord)

		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
			return
		}
		// FIXME: I don't know, maybe I should use transaction
		// When one of the following operations fails, the saved file should be deleted
		// And both of the tables are not synchronized
		if err := core.AddClipboardRecord(newClipboardRecord); err != nil {
			logger.Tracef("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
			return
		}
		if err = core.AddFileMetadataRecord(fileMetadata); err != nil {
			logger.Tracef("AddFileMetadataRecord error: %v", err)

			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
	}
}
func HandlerHistory(ctx *gin.Context) {
	// TODO:Designed for WebUI, pulling all data from server
}
