package model

import (
	"os"

	"github.com/sirupsen/logrus"
)

var baseLogger = logrus.New()

func InitLogger(c *UContext) {
	baseLogger.SetFormatter(&logrus.TextFormatter{})
	if c.Runtime.LogPath != "" {
		file, err := os.OpenFile(c.Runtime.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			baseLogger.SetOutput(file)
		} else {
			baseLogger.SetOutput(os.Stdout)
			baseLogger.Warnf("Failed to log to file, using default stdout: %v", err)
		}
	}
	switch c.Runtime.LogLevel {
	case "debug":
		baseLogger.SetLevel(logrus.DebugLevel)
	case "trace":
		baseLogger.SetLevel(logrus.TraceLevel)
	default:
		baseLogger.SetLevel(logrus.InfoLevel)
	}

}

func NewModuleLogger(module_name string) *logrus.Entry {
	return baseLogger.WithField("module", module_name)
}
