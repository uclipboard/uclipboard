package server

import (
	"time"

	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/server/core"
)

func TimerGC(uctx *model.UContext) {
	logger := model.NewModuleLogger("timer")
	logger.Infof("TimerGC started, interval: %ds", uctx.Server.TimerInterval)
	interval := time.Duration(uctx.Server.TimerInterval) * time.Second
	for {
		time.Sleep(interval)

		// Get the current time
		now := time.Now().UnixMilli()
		// Query the database for expired files
		expiredFiles, err := core.QueryExpiredFiles(uctx, now)
		if err != nil {
			logger.Warnf("Query expired files failed: %v", err)
			continue
		}

		// Delete expired files
		for _, file := range expiredFiles {
			logger.Debugf("Expired file: Name=%s, CreatedAt=%d", file.FileName, file.CreatedTs)

			err = core.DelFileMetadataRecordById(&file)
			if err != nil {
				logger.Warnf("Delete expired file metadata record failed: %v", err)
				continue
			}

			err := core.DelTmpFile(uctx, &file)
			if err != nil {
				logger.Warnf("Delete expired file failed: %v", err)
				continue
			}
		}
		
		if len(expiredFiles) == 0 {
			logger.Debugf("No expired files")
		} else {
			logger.Debugf("Expired files count: %d", len(expiredFiles))
		}

		// delete outdated clipboard records
		if uctx.Server.Store.MaxClipboardRecordNumber == 0 {
			logger.Debugf("MaxClipboardRecordNumber is 0, skip delete outdated clipboard records")
		} else {
			err = core.DeleteOutdatedClipboard(uctx)
			if err != nil {
				logger.Warnf("Delete outdated clipboard records failed: %v", err)
			}
		}

	}
}
