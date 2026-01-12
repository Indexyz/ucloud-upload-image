package utils

import "strings"

func Yes(data string) bool {
	if len(data) == 0 {
		return false
	}

	if strings.EqualFold(data, "true") {
		return true
	}

	if strings.EqualFold(data, "yes") {
		return true
	}

	if strings.EqualFold(data, "on") {
		return true
	}

	if strings.EqualFold(data, "1") {
		return true
	}

	return false
}
