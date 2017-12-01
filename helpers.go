package qbin

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseExpiration creates a time.Time object from an expiration string, taking the units m, h, d, w into account.
func ParseExpiration(expiration string) (time.Time, error) {
	expiration = strings.TrimSpace(expiration)
	var multiplier int64

	if strings.HasSuffix(expiration, "h") {
		expiration = strings.TrimSuffix(expiration, "h")
		multiplier = 60
	} else if strings.HasSuffix(expiration, "d") {
		expiration = strings.TrimSuffix(expiration, "d")
		multiplier = 60 * 24
	} else if strings.HasSuffix(expiration, "w") {
		expiration = strings.TrimSuffix(expiration, "w")
		multiplier = 60 * 24 * 7
	} else {
		expiration = strings.TrimSuffix(expiration, "m")
		multiplier = 1
	}

	value, err := strconv.ParseInt(expiration, 10, 0)
	if err != nil {
		return time.Time{}, err
	}

	expirationTime := time.Now()
	expirationTime.Add(time.Duration(multiplier*value) * time.Minute)

	return expirationTime, nil
}

// escapeHTML removes all special HTML characters (namely, &<>") in a string and replaces them with their entities (e.g. &amp;).
func escapeHTML(content string) string {
	content = strings.Replace(content, "&", "&amp;", -1)
	content = strings.Replace(content, "<", "&lt;", -1)
	content = strings.Replace(content, ">", "&gt;", -1)
	content = strings.Replace(content, "\"", "&quot;", -1)
	return content
}

// This does not match all HTML tags, but those created by Prism.js are fine for us.
var htmlTags = regexp.MustCompile(`<[^>]+>`)

// stripHTML strips all HTML tags and replaces the entities from escapeHTML backwards.
func stripHTML(content string) string {
	content = htmlTags.ReplaceAllString(content, "")
	content = strings.Replace(content, "&quot;", "\"", -1)
	content = strings.Replace(content, "&gt;", ">", -1)
	content = strings.Replace(content, "&lt;", "<", -1)
	content = strings.Replace(content, "&amp;", "&", -1)
	return content
}
