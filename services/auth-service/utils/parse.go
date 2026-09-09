package utils

import (
	"strconv"
	"strings"
)

// ParsePositiveInt parses raw as a positive integer, falling back when empty
// or invalid.
func ParsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

// BearerToken extracts the bearer token from an Authorization header value.
func BearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
