package server

import (
	"fmt"
	"net/http"
	"path"
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
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
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
		if file.Filename == "" {
			// It would be happened on my client
			file.Filename = model.RandString(8)
			logger.Warnf("Filename is empty, generate a random filename:%v", file.Filename)
		}

		fileMetadata.FileName = file.Filename
		fileMetadata.TmpPath = fmt.Sprintf("%s_%s", strconv.FormatInt(fileMetadata.CreatedTs, 10), file.Filename)
		fileMetadata.ExpireTs = conf.Server.DefaultFileLife + fileMetadata.CreatedTs
		logger.Tracef("fileMetadata: %v", fileMetadata)

		newClipboardRecord := model.NewClipoardWithDefault()
		hostname := ctx.Request.Header.Get("hostname")
		if hostname == "" {
			hostname = "unknown"
		}
		logger.Tracef("uploader's hostname: %s", hostname)
		newClipboardRecord.Hostname = hostname
		newClipboardRecord.ContentType = "binary"

		// save file to tmp directory and get the path to save in db
		filePath := filepath.Join(conf.Server.TmpPath, fileMetadata.TmpPath)
		newClipboardRecord.Content = fileMetadata.FileName
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
			return
		}
		logger.Tracef("newClipboardRecord: %v", newClipboardRecord)
		// FIXME: I don't know, maybe I should use transaction
		// When one of the following operations fails, the saved file should be deleted
		// And at that time, both of the tables are not synchronized
		if err := core.AddClipboardRecord(newClipboardRecord); err != nil {
			logger.Tracef("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
			return
		}
		fileId, err := core.AddFileMetadataRecord(fileMetadata)
		if err != nil {
			logger.Tracef("AddFileMetadataRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
			return
		}
		logger.Debugf("the new fileId: %v", fileId)
		ctx.JSON(http.StatusOK, gin.H{"message": "ok", "fileId": fileId, "filename": fileMetadata.FileName})
	}
}

func HandlerDownload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerDownload")

	return func(ctx *gin.Context) {
		logger.Tracef("download request: %v", ctx.Param("filename"))
		fileName := ctx.Param("filename")[1:]
		metadata := model.NewFileMetadataWithDefault()
		if fileName == "" {
			if err := core.GetFileMetadataLatestRecord(metadata); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
				return
			}
			logger.Tracef("Get the latest file metadata record: %v", metadata)
		} else {
			if fileName[0] == '@' {
				idInt, err := strconv.Atoi(fileName[1:])
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
				}
				metadata.Id = int64(idInt)
				logger.Tracef("download by id: %v", metadata.Id)
			} else {
				metadata.FileName = fileName
				logger.Tracef("download by name: %s", metadata.FileName)

			}

			err := core.GetFileMetadataRecordByOrName(metadata)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal Error:" + err.Error()})
				return
			}
		}
		fullPath := path.Join(conf.Server.TmpPath, metadata.TmpPath)
		logger.Tracef("filePath: %s", fullPath)
		// set file name in header
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, metadata.FileName))
		// FIXME: What will be happened if send this file when the file is expired and deleted by the timerGC?
		ctx.File(fullPath)
	}

}

func HandlerHistory(ctx *gin.Context) {
	// TODO:Designed for WebUI, pulling all data from server
}
