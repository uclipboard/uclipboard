package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/server/core"
)

func HandlerPush(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerPush")

	return func(ctx *gin.Context) {
		clipboardData := model.NewClipoardWithDefault()
		if err := ctx.BindJSON(&clipboardData); err != nil {
			logger.Debugf("BindJSON error: %v", err)
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("request is invalid.", nil))
			return
		}

		if len(clipboardData.Content) > conf.MaxClipboardSize {
			ctx.JSON(http.StatusRequestEntityTooLarge, model.NewDefaultServeRes(fmt.Sprintf("clipboard is too large[limit: %dB]", conf.MaxClipboardSize), nil))
			return
		}
		if clipboardData.Content == "" {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("content is empty", nil))
			return
		}
		if err := core.AddClipboardRecord(clipboardData); err != nil {
			logger.Debugf("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("add clipboard record error", nil))
			return
		}
		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", nil))
	}
}
func HandlerPull(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerPull")

	return func(ctx *gin.Context) {
		clipboardArr, err := core.QueryLatestClipboardRecord(conf.Server.PullHistorySize)
		if err != nil {
			logger.Debugf("GetLatestClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("get latest clipboard record error", nil))
			return
		}
		clipboardsBytes, err := json.Marshal(clipboardArr)
		if err != nil {
			logger.Debugf("Marshal clipboardArr error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("marshal clipboardArr error", nil))
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", clipboardsBytes))
	}
}

func HandlerUpload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerUpload")
	return func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("can't read file from request", nil))
			return
		}
		lifetime := ctx.Query("lifetime")
		lifetimeSecs, err := core.ConvertLifetime(lifetime, conf.Server.DefaultFileLife)
		if err != nil {
			logger.Debugf("ConvertLifetime error: %v", err)
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("invalid lifetime", nil))
			return
		}
		logger.Tracef("lifetime: %vs", lifetimeSecs)

		hostname := ctx.Request.Header.Get("hostname")
		if hostname == "" {
			hostname = "unknown"
		}
		logger.Debugf("uploader's hostname is: %s", hostname)

		if file.Filename == "" {
			// It would not be happened on uclipboard client
			file.Filename = model.RandString(8)
			logger.Warnf("Filename is empty, generate a random filename:%v", file.Filename)
		}
		// prepare to download file
		fileMetadata := model.NewFileMetadataWithDefault()
		fileMetadata.FileName = file.Filename
		fileMetadata.TmpPath = fmt.Sprintf("%s_%s", strconv.FormatInt(fileMetadata.CreatedTs, 10), file.Filename)
		fileMetadata.ExpireTs = lifetimeSecs*1000 + fileMetadata.CreatedTs
		logger.Debugf("Upload file metadata is: %v", fileMetadata)

		// save file to tmp directory and get the path to save in db
		filePath := filepath.Join(conf.Server.TmpPath, fileMetadata.TmpPath)
		logger.Debugf("Save file to: %s", filePath)

		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			logger.Warnf("SaveUploadedFile error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("save file error", nil))
			return
		}

		// FIXME: I don't know, maybe I should use transaction
		// When one of the following operations fails, the saved file should be deleted
		// And at that time, both of the tables are not synchronized

		// save file metadata to db
		fileId, err := core.AddFileMetadataRecord(fileMetadata)
		if err != nil {
			logger.Tracef("AddFileMetadataRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("add file metadata record error", nil))
			return
		}
		logger.Debugf("The new file id is %v", fileId)

		// save clipboard record to db
		newClipboardRecord := model.NewClipoardWithDefault()
		newClipboardRecord.Content = fmt.Sprintf("%s@%d", fileMetadata.FileName, fileId) // add fileid to clipboard record
		newClipboardRecord.Hostname = hostname
		newClipboardRecord.ContentType = "binary"
		logger.Tracef("Upload binary file clipboard record: %v", newClipboardRecord)
		if err := core.AddClipboardRecord(newClipboardRecord); err != nil {
			logger.Debugf("AddClipboardRecord error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("add clipboard record error", nil))
			return
		}

		responseData, err := json.Marshal(gin.H{"file_id": fileId, "file_name": fileMetadata.FileName,
			"life_time": lifetimeSecs})
		if err != nil {
			logger.Debugf("Marshal response data error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("marshal response data error", nil))
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", responseData))
	}
}

func HandlerDownload(conf *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerDownload")

	return func(ctx *gin.Context) {
		logger.Trace("into HandlerDownload")
		logger.Tracef("request download raw filename paramater: %v", ctx.Param("filename"))

		fileName := ctx.Param("filename")[1:] // skip '/' in '/xxx'
		metadata := model.NewFileMetadataWithDefault()
		if fileName == "" {
			// download latest binary file
			if err := core.GetFileMetadataLatestRecord(metadata); err != nil {
				if err == sql.ErrNoRows {
					ctx.JSON(http.StatusNotFound, model.NewDefaultServeRes("there are no files in server.", nil))
					return
				}

				logger.Debugf("GetFileMetadataLatestRecord error: %v", err)
				ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("get the latest file metadata record error", nil))
				return
			}

			logger.Debugf("Get the latest file metadata record: %v", metadata)
		} else {
			if strings.Contains(fileName, "@") {
				logger.Trace("download file by id")
				// get the id number starts after @
				id := core.ExtractFileId(fileName, "@")
				if id == 0 {
					logger.Debugf("invalid format of file id: %v", id)
					ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("invalid format of file id.", nil))
					return
				}
				metadata.Id = id
				logger.Debugf("download by id: %v", metadata.Id)
			} else {
				// download by name
				metadata.FileName = fileName
				logger.Debugf("download by name: %s", metadata.FileName)
			}

			err := core.GetFileMetadataRecordByIdOrName(metadata)
			if err != nil {
				if err == sql.ErrNoRows {
					ctx.JSON(http.StatusNotFound, model.NewDefaultServeRes("specified file not found or has expired.", nil))
					return
				}
				ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("get file metadata record error: "+err.Error(), nil))
				return
			}

		}

		fullPath := path.Join(conf.Server.TmpPath, metadata.TmpPath)
		logger.Debugf("Required file full path: %s", fullPath)
		// set file name in header
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, metadata.FileName))
		ctx.File(fullPath)
	}
}

func HandlerHistory(c *model.Conf) func(ctx *gin.Context) {
	logger := model.NewModuleLogger("HandlerHistory")
	return func(ctx *gin.Context) {
		logger.Trace("into HeadlerHistory")
		page := ctx.Query("page")
		if page == "" {
			page = "1"
		}
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("invalid page number", nil))
			return
		}
		logger.Debugf("Request clipboard history page: %v", pageInt)
		clipboardsCount, err := core.CountClipboardHistory(c)
		historyPageCount := math.Ceil(float64(clipboardsCount) / float64(c.Server.ClipboardHistoryPageSize))

		if err != nil {
			logger.Debugf("CountClipboardHistory error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("count clipboard history error", nil))
			return
		}

		clipboards, err := core.QueryClipboardHistory(c, pageInt)
		if err != nil {
			logger.Debugf("QueryClipboardHistory error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("query clipboard history error", nil))
			return
		}
		logger.Tracef("clipboards: %v", clipboards)
		clipboardsBytes, err := json.Marshal(gin.H{"history": clipboards, "pages": historyPageCount})
		if err != nil {
			logger.Debugf("Marshal clipboards error: %v", err)
			ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes("marshal clipboards error", nil))
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", clipboardsBytes))
	}
}

func HandlerPublicShare(c *model.Conf) func(ctx *gin.Context) {
	// TODO share binary file to public
	return func(ctx *gin.Context) {
	}
}
