package utils

import "time"

func GetCurrentTimeFormatted() string {
	return time.Now().Format("15:04:05")
}
