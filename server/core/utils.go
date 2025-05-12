package core

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/uclipboard/uclipboard/model"
)

// lifetime: s unit
func ConvertLifetime(lifetime string, defaultLifetime int64) (int64, error) {
	var lifetimeSecs int64
	if lifetime != "" {
		lifetimeInt, err := strconv.ParseInt(lifetime, 10, 64)
		if err != nil {
			return 0, err
		}
		lifetimeSecs = lifetimeInt
	} else {
		lifetimeSecs = defaultLifetime
	}
	// hardcode: the maximum lifetime is 90 days
	if lifetimeSecs > 60*60*24*90 {
		lifetimeSecs = 60 * 60 * 24 * 90
	}
	return lifetimeSecs, nil
}

func ExtractFileId(s string, startChar string) int64 {
	if !strings.Contains(s, startChar) {
		return 0
	}
	idStr := s[strings.Index(s, startChar)+1:]
	var idInt64 int64
	idInt64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return idInt64
}

func ServerInternalErrorLogEcho(ctx *gin.Context, logger *logrus.Entry, msg string, args ...any) {
	logger.Errorf(msg, args...)
	ctx.JSON(http.StatusInternalServerError, model.NewDefaultServeRes(msg, nil))
}

func AddClipboardRecordAndNotify(uctx *model.UContext, clipboardData *model.Clipboard) error {
	logger := model.NewModuleLogger("AddClipboardRecordAndNotify")
	logger.Tracef("before insert: %v", clipboardData)
	if err := AddClipboardRecord(clipboardData); err != nil {
		return err
	}
	logger.Tracef("after insert: %v", clipboardData)
	uctx.Runtime.ClipboardUpdateNotify.Push(clipboardData)
	return nil
}
