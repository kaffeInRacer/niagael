package utils

import (
	"regexp"
	"strings"
)

var (
	nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9\s-]`)
	multiDashRegex   = regexp.MustCompile(`-{2,}`)
)

func GenerateSlug(s string) string {
	slug := strings.ToLower(s)
	slug = strings.TrimSpace(slug)
	slug = nonAlphaNumRegex.ReplaceAllString(slug, "")
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = multiDashRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
