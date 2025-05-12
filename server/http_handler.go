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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/server/core"
)

func HandlerPush(uctx *model.UContext) gin.HandlerFunc {
	logger := model.NewModuleLogger("HandlerPush")

	return func(ctx *gin.Context) {
		clipboardData := model.NewClipboardWithDefault()
		if err := ctx.BindJSON(&clipboardData); err != nil {
			logger.Debugf("BindJSON error: %v", err)
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("request is invalid.", nil))
			return
		}
		// we should not trust the client's timestamp
		clipboardData.Ts = time.Now().UnixMilli()
		if len(clipboardData.Content) > uctx.ContentLengthLimit {
			ctx.JSON(http.StatusRequestEntityTooLarge, model.NewDefaultServeRes(fmt.Sprintf("clipboard is too large[limit: %dB]", uctx.ContentLengthLimit), nil))
			return
		}
		if clipboardData.Content == "" {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("content is empty", nil))
			return
		}
		if err := core.AddClipboardRecordAndNotify(uctx, clipboardData); err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "AddClipboardRecordAndNotify error: %v", err)
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", nil))
	}
}
func HandlerPull(conf *model.UContext) gin.HandlerFunc {
	logger := model.NewModuleLogger("HandlerPull")

	return func(ctx *gin.Context) {
		clipboardArr, err := core.QueryLatestClipboardRecord(conf.Server.Api.PullSize)
		if err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "GetLatestClipboardRecord error: %v", err)
			return
		}
		clipboardsBytes, err := json.Marshal(clipboardArr)
		if err != nil {
			logger.Debugf("Marshal clipboardArr error: %v", err)
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("marshal clipboardArr error", nil))
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", clipboardsBytes))
	}
}

func HandlerUpload(uctx *model.UContext) gin.HandlerFunc {
	logger := model.NewModuleLogger("HandlerUpload")
	return func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("can't read file from request", nil))
			return
		}
		lifetime := ctx.Query("lifetime")
		lifetimeSecs, err := core.ConvertLifetime(lifetime, uctx.Server.Store.DefaultFileLife)
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
		// prepare to upload file
		fileMetadata := model.NewFileMetadataWithDefault()
		fileMetadata.FileName = file.Filename
		fileMetadata.TmpPath = fmt.Sprintf("%s_%s", strconv.FormatInt(fileMetadata.CreatedTs, 10), file.Filename)
		fileMetadata.ExpireTs = lifetimeSecs*1000 + fileMetadata.CreatedTs
		logger.Debugf("Upload file metadata is: %v", fileMetadata)

		// save file to tmp directory and get the path to save in db
		filePath := filepath.Join(uctx.Server.Store.TmpPath, fileMetadata.TmpPath)
		logger.Debugf("Save file to: %s", filePath)

		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "SaveUploadedFile error: %v", err)
			return
		}

		// FIXME: I don't know, maybe I should use transaction
		// When one of the following operations fails, the saved file should be deleted
		// And at that time, both of the tables are not synchronized

		// save file metadata to db
		fileId, err := core.AddFileMetadataRecord(fileMetadata)
		if err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "AddFileMetadataRecord error: %v", err)
			return
		}
		logger.Debugf("The new file id is %v", fileId)

		// save clipboard record to db
		newClipboardRecord := model.NewClipboardWithDefault()
		newClipboardRecord.Content = fmt.Sprintf("%s@%d", fileMetadata.FileName, fileId) // add fileid to clipboard record
		newClipboardRecord.Hostname = hostname
		newClipboardRecord.ContentType = "binary"
		logger.Tracef("Upload binary file clipboard record: %v", newClipboardRecord)
		if err := core.AddClipboardRecordAndNotify(uctx, newClipboardRecord); err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "AddClipboardRecordAndNotify error: %v", err)
			return
		}
		fmr := model.FileMetadataResponse{
			Id:       fileId,
			Name:     fileMetadata.FileName,
			LifeTime: lifetimeSecs,
		}
		responseData, err := json.Marshal(fmr)
		if err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "Marshal response data error: %v", err)
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", responseData))
	}
}

func HandlerDownload(conf *model.UContext) gin.HandlerFunc {
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
				core.ServerInternalErrorLogEcho(ctx, logger, "GetFileMetadataLatestRecord error: %v", err)
				return
			}

			logger.Debugf("Get the latest file metadata record: %v", metadata)
		} else {
			if strings.Contains(fileName, "@") {
				logger.Trace("download file by id")
				// get the id number starts after @
				id := core.ExtractFileId(fileName, "@")
				if id == 0 {
					core.ServerInternalErrorLogEcho(ctx, logger, "Wrong file id format")
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
				core.ServerInternalErrorLogEcho(ctx, logger, "GetFileMetadataRecordByIdOrName error: %v", err)
				return
			}
		}

		fullPath := path.Join(conf.Server.Store.TmpPath, metadata.TmpPath)
		logger.Debugf("Required file full path: %s", fullPath)
		// set file name in header
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, metadata.FileName))
		ctx.File(fullPath)
	}
}

func HandlerHistory(c *model.UContext) gin.HandlerFunc {
	logger := model.NewModuleLogger("HandlerHistory")
	return func(ctx *gin.Context) {
		logger.Trace("into HeadlerHistory")
		page := ctx.Query("page")
		if page == "" {
			page = "1"
		}
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			logger.Errorf("Atoi error: %v", err)
			ctx.JSON(http.StatusBadRequest, model.NewDefaultServeRes("invalid page number", nil))
			return
		}
		logger.Debugf("Request clipboard history page: %v", pageInt)
		clipboardsCount, err := core.CountClipboardHistory(c)
		if err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "CountClipboardHistory error: %v", err)
			return
		}

		historyPageCount := math.Ceil(float64(clipboardsCount) / float64(c.Server.Api.HistoryPageSize))
		clipboards, err := core.QueryClipboardHistory(c, pageInt)
		if err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "QueryClipboardHistory error: %v", err)
			return
		}
		historyResponse := model.HistoryResponse{
			History: clipboards,
			Pages:   int64(historyPageCount),
			Total:   int64(clipboardsCount),
		}
		clipboardsBytes, err := json.Marshal(&historyResponse)
		if err != nil {
			core.ServerInternalErrorLogEcho(ctx, logger, "Marshal clipboards error: %v", err)
			return
		}

		ctx.JSON(http.StatusOK, model.NewDefaultServeRes("", clipboardsBytes))
	}
}
