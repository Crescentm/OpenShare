package stringsx

import "strings"

func TrimmedPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func Deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func Ptr(value string) *string {
	copied := value
	return &copied
}
