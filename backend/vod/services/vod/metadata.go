package vod

import (
	"strings"
	"unicode/utf8"

	"sen1or/letslive/vod/domains"
)

// normalizeTitle trims the title; a whitespace-only title counts as empty.
func normalizeTitle(title string) string {
	return strings.TrimSpace(title)
}

// uploadTitle returns the title an upload is stored with. A blank title falls
// back to the file name, cut to the maximum length because the user never
// typed it; a typed title that is too long is rejected instead.
func uploadTitle(title, filename string) (string, bool) {
	title = normalizeTitle(title)
	if title == "" {
		title = truncateRunes(strings.TrimSpace(filename), domains.VODTitleMaxLength)
	}
	count := utf8.RuneCountInString(title)
	return title, count >= domains.VODTitleMinLength && count <= domains.VODTitleMaxLength
}

func validDescription(description string) bool {
	return utf8.RuneCountInString(description) <= domains.VODDescriptionMaxLength
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}
