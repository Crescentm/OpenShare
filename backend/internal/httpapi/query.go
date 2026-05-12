package httpapi

import "strconv"

func ParseIntQuery(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}
