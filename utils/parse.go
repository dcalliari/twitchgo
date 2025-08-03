package utils

import "strings"

func ParseCommand(input string) (cmd string, args []string) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", nil
	}
	return strings.ToLower(fields[0]), fields[1:]
}
