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
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("request is invalid.", gin.H{}))
			return
		}
		if err := core.AddClipboardRecord(clipboardData); err != nil {
			logger.Tracef("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("add clipboard record error", gin.H{}))
			return
		}
		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", gin.H{}))
	}
}
func HandlerPull(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerPull")

	return func(ctx *gin.Context) {
		clipboardArr, err := core.QueryLatestClipboardRecord(conf.Server.PullHistorySize)
		if err != nil {
			logger.Tracef("GetLatestClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("get latest clipboard record error", gin.H{}))
			return
		}
		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", clipboardArr))
	}
}

func HandlerUpload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerUpload")
	return func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("can't read file from request", gin.H{}))
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
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("save file error", gin.H{}))
			return
		}
		logger.Tracef("Upload binary file clipboard record: %v", newClipboardRecord)
		// FIXME: I don't know, maybe I should use transaction
		// When one of the following operations fails, the saved file should be deleted
		// And at that time, both of the tables are not synchronized
		if err := core.AddClipboardRecord(newClipboardRecord); err != nil {
			logger.Tracef("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("add clipboard record error", gin.H{}))
			return
		}
		fileId, err := core.AddFileMetadataRecord(fileMetadata)
		if err != nil {
			logger.Tracef("AddFileMetadataRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("add file metadata record error", gin.H{}))
			return
		}
		logger.Debugf("The new file id is %v", fileId)
		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("",
			gin.H{"file_id": fileId, "file_name": fileMetadata.FileName,
				"life_time": conf.Server.DefaultFileLife}))
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
				ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("get the latest file metadata record error", gin.H{}))
				return
			}
			logger.Debugf("Get the latest file metadata record: %v", metadata)
		} else {
			if fileName[0] == '@' {
				idInt, err := strconv.Atoi(fileName[1:])
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("invalid format of file id ", gin.H{}))
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
					ctx.JSON(http.StatusNotFound, model.NewDefaultServeRes("file not found", gin.H{}))
					return
				}
				ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("get file metadata record error: "+err.Error(), gin.H{}))
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
	logger := model.NewModuleLogger("HandlerHistory")
	return func(ctx *gin.Context) {
		page := ctx.Query("page")
		if page == "" {
			page = "1"
		}
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("invalid page number", gin.H{}))
			return
		}
		logger.Tracef("Request clipboard history page: %v", pageInt)
		clipboards, err := core.QueryClipboardHistory(c, pageInt)
		if err != nil {
			logger.Tracef("QueryClipboardHistory error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("query clipboard history error", gin.H{}))
			return
		}
		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", clipboards))
	}
}

func HandlerPublicShare(c *model.Conf) func(ctx *gin.Context) {
	// TODO share binary file to public
	return func(ctx *gin.Context) {
	}
}
