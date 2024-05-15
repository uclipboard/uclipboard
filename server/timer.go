package server

import (
	"time"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/dangjinghao/uclipboard/server/core"
)

func TimerGC(conf *model.Conf) {
	logger := model.NewModuleLogger("timer")
	logger.Debugf("TimerGC started, interval=%ds", conf.Server.TimerInterval)
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
		if len(expiredFiles) == 0 {
			logger.Debugf("No expired files")
		}
		// Delete expired files
		for _, file := range expiredFiles {
			logger.Debugf("Expired file: Name=%s, CreatedAt=%d", file.FileName, file.CreatedTs)
			err := core.DelTmpFile(conf, &file)
			if err != nil {
				logger.Warnf("Delete expired file failed: %v", err)
			}
			err = core.DelFileMetadataRecordById(&file)
			if err != nil {
				logger.Warnf("Delete expired file metadata record failed: %v", err)
			}

		}

		time.Sleep(interval)
	}
}
