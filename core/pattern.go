package core

import (
	"regexp"
	"strings"
)

func isWildcard(s string) bool {
	return strings.Contains(s, "*")
}

func patternToRegex(pattern string) string {
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, "\\*", "(.*)")
	return "^" + escaped + "$"
}
