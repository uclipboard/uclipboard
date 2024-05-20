package server

import (
	"database/sql"
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
	logger := model.NewModuleLogger("HandlerPush")

	return func(ctx *gin.Context) {
		clipboardData := model.NewClipoardWithDefault()
		if err := ctx.BindJSON(&clipboardData); err != nil {
			logger.Tracef("BindJSON error: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Request is invalid."})
			return
		}
		if err := core.AddClipboardRecord(clipboardData); err != nil {
			logger.Tracef("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "add clipboard record error: " + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}
func HandlerPull(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerPull")

	return func(ctx *gin.Context) {
		var clipboardArr []model.Clipboard
		if err := core.GetLatestClipboardRecord(&clipboardArr, conf.Server.PullHistorySize); err != nil {
			logger.Tracef("GetLatestClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "get latest clipboard record error: " + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, clipboardArr)
	}
}

func HandlerUpload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerUpload")
	return func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Cannot read file from request:" + err.Error()})
			return
		}
		fileMetadata := model.NewFileMetadataWithDefault()
		if file.Filename == "" {
			// It would not be happened on uclipboard client
			file.Filename = model.RandString(8)
			logger.Warnf("Filename is empty, generate a random filename:%v", file.Filename)
		}

		fileMetadata.FileName = file.Filename
		fileMetadata.TmpPath = fmt.Sprintf("%s_%s", strconv.FormatInt(fileMetadata.CreatedTs, 10), file.Filename)
		fileMetadata.ExpireTs = conf.Server.DefaultFileLife + fileMetadata.CreatedTs
		logger.Tracef("Upload file metadata is: %v", fileMetadata)

		newClipboardRecord := model.NewClipoardWithDefault()
		hostname := ctx.Request.Header.Get("hostname")
		if hostname == "" {
			hostname = "unknown"
		}
		logger.Tracef("uploader's hostname is: %s", hostname)
		newClipboardRecord.Hostname = hostname
		newClipboardRecord.ContentType = "binary"

		// save file to tmp directory and get the path to save in db
		filePath := filepath.Join(conf.Server.TmpPath, fileMetadata.TmpPath)
		newClipboardRecord.Content = fileMetadata.FileName
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			logger.Tracef("SaveUploadedFile error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server save internal error:" + err.Error()})
			return
		}
		logger.Tracef("Upload binary file clipboard record: %v", newClipboardRecord)
		// FIXME: I don't know, maybe I should use transaction
		// When one of the following operations fails, the saved file should be deleted
		// And at that time, both of the tables are not synchronized
		if err := core.AddClipboardRecord(newClipboardRecord); err != nil {
			logger.Tracef("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "add clipboard record error:" + err.Error()})
			return
		}
		fileId, err := core.AddFileMetadataRecord(fileMetadata)
		if err != nil {
			logger.Tracef("AddFileMetadataRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "add file metadata record error:" + err.Error()})
			return
		}
		logger.Debugf("The new file id is %v", fileId)
		ctx.JSON(http.StatusOK, gin.H{"message": "ok", "file_id": fileId, "filename": fileMetadata.FileName, "life_time": conf.Server.DefaultFileLife})
	}
}

func HandlerDownload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerDownload")

	return func(ctx *gin.Context) {
		logger.Tracef("request download raw filename paramater: %v", ctx.Param("filename"))

		fileName := ctx.Param("filename")[1:] // skip '/' in '/xxx'
		metadata := model.NewFileMetadataWithDefault()
		if fileName == "" {
			if err := core.GetFileMetadataLatestRecord(metadata); err != nil {
				logger.Tracef("GetFileMetadataLatestRecord error: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "get latest file metadata error:" + err.Error()})
				return
			}
			logger.Debugf("Get the latest file metadata record: %v", metadata)
		} else {
			if fileName[0] == '@' {
				idInt, err := strconv.Atoi(fileName[1:])
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server internal error:" + err.Error()})
				}
				metadata.Id = int64(idInt)
				logger.Debugf("download by id: %v", metadata.Id)
			} else {
				metadata.FileName = fileName
				logger.Debugf("download by name: %s", metadata.FileName)

			}

			err := core.GetFileMetadataRecordByOrName(metadata)
			if err != nil {
				if err == sql.ErrNoRows {
					ctx.JSON(http.StatusNotFound, gin.H{"message": "the file may be expired or does not exists."})
					return
				}
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "get file metadata record error:" + err.Error()})
				return
			}
		}
		fullPath := path.Join(conf.Server.TmpPath, metadata.TmpPath)
		logger.Debugf("Required file full path: %s", fullPath)
		// set file name in header
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, metadata.FileName))
		// FIXME: What will be happened if send this file when the file is expired and deleted by the timerGC?
		ctx.File(fullPath)
	}

}

func HandlerHistory(c *model.Conf) func(ctx *gin.Context) {
	// TODO:Designed for WebUI, pulling all data from server
	return func(ctx *gin.Context) {

	}
}

func HandlerPublicShare(c *model.Conf) func(ctx *gin.Context) {
	// TODO share binary file to public
	return func(ctx *gin.Context) {
	}
}
