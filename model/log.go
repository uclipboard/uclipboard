package model

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func InitLogger(c *UContext) {
	logger.SetFormatter(&logrus.TextFormatter{})
	if c.Runtime.LogPath != "" {
		file, err := os.OpenFile(c.Runtime.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		} else {
			logger.SetOutput(os.Stdout)
			logger.Warnf("Failed to log to file, using default stdout: %v", err)
		}
	}
	switch c.Runtime.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

}

func NewModuleLogger(module_name string) *logrus.Entry {
	return logger.WithField("module", module_name)
}
