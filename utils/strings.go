package utils

import "strings"

func SanitizeMessage(msg string) string {
	return strings.TrimSpace(strings.ToLower(msg))
}
