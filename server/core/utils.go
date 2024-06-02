package core

import (
	"strconv"
	"strings"
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
