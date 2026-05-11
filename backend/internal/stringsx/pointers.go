package stringsx

import "strings"

func TrimmedPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
