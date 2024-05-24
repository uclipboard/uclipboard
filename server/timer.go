package server

import (
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
)

func TimerGC(conf *model.Conf) {
	logger := model.NewModuleLogger("timer")
	logger.Infof("TimerGC started, interval: %ds", conf.Server.TimerInterval)
	interval := time.Duration(conf.Server.TimerInterval) * time.Second
	for {
		// Get the current time
		now := time.Now().Unix()
		// Query the database for expired files
		expiredFiles, err := core.QueryExpiredFiles(conf, now)
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
			}

			err := core.DelTmpFile(conf, &file)
			if err != nil {
				logger.Warnf("Delete expired file failed: %v", err)
			}
		}
		if len(expiredFiles) == 0 {
			logger.Debugf("No expired files")
		} else {
			logger.Debugf("Expired files count: %d", len(expiredFiles))
		}
		// TODO: clean clipboard data
		// I think delete clipboard records is not a good idea
		time.Sleep(interval)
	}
}
