package utils

import (
	"log"
	"strconv"
)

func ParseToUint(val string, def ...uint64) uint64 {
	if val == "" {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	parsed, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		log.Fatal("Error parsing value:", err)
	}
	return parsed
}
