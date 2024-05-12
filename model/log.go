package model

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func LoggerInit(logInfo string) {
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(os.Stdout)
	switch logInfo {
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
