package core

import (
	"strconv"
	"strings"
)

// lifetime: ms	unit
func ConvertLifetime(lifetime string, defaultLifetime int64) (int64, error) {
	var lifetimeMS int64
	if lifetime != "" {
		lifetimeInt, err := strconv.ParseInt(lifetime, 10, 64)
		if err != nil {
			return 0, err
		}
		lifetimeMS = lifetimeInt
	} else {
		lifetimeMS = defaultLifetime
	}
	// hardcode: the maximum lifetime is 90 days
	if lifetimeMS > 1000*60*60*24*90 {
		lifetimeMS = 1000 * 60 * 60 * 24 * 90
	}
	return lifetimeMS, nil
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
