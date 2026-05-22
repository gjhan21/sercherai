package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RandomHex generates a random hex string of length n*2.
func RandomHex(n int) string {
	if n <= 0 {
		n = 4
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

// SanitizeUploadFileName removes path separators and returns the base file name.
func SanitizeUploadFileName(fileName string) string {
	name := strings.TrimSpace(filepath.Base(fileName))
	if name == "" || name == "." || name == ".." {
		return ""
	}
	return name
}

// NormalizeAdminDateTime converts various date string formats to RFC3339.
func NormalizeAdminDateTime(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("is required")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			if ts, err := time.Parse(layout, trimmed); err == nil {
				return ts.Format(time.RFC3339), nil
			}
			continue
		}
		if ts, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return ts.Format(time.RFC3339), nil
		}
	}
	return "", errors.New("format invalid, expected RFC3339 or YYYY-MM-DD")
}

// ParseConfigBool converts string values like "true", "1", "yes" to bool.
func ParseConfigBool(raw string, fallback bool) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return fallback
	}
	switch text {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

// ParseConfigInt converts string values to int.
func ParseConfigInt(raw string, fallback int) int {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fallback
	}
	return value
}

// ResolveUploadExt returns the file extension based on file name or mime type.
func ResolveUploadExt(fileName string, mimeType string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	if ext != "" && len(ext) <= 10 {
		return ext
	}
	if exts, err := mime.ExtensionsByType(mimeType); err == nil && len(exts) > 0 {
		return strings.ToLower(exts[0])
	}
	return ".bin"
}

// JoinObjectPath joins path parts with forward slashes, removing leading/trailing slashes from parts.
func JoinObjectPath(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		p = strings.Trim(p, "/")
		if p == "" {
			continue
		}
		normalized = append(normalized, p)
	}
	return strings.Join(normalized, "/")
}

// ParsePage extracts page and page_size from gin context with defaults.
func ParsePage(c *gin.Context) (int, int) {
	page := ParseIntOrDefault(c.Query("page"), 1)
	pageSize := ParseIntOrDefault(c.Query("page_size"), 20)
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

// ParseIntOrDefault parses a string to int with a fallback value.
func ParseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
// UniqueNonEmptyStrings returns a slice of unique non-empty strings from the input.
func UniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		current := strings.TrimSpace(value)
		if current == "" {
			continue
		}
		if _, exists := seen[current]; exists {
			continue
		}
		seen[current] = struct{}{}
		result = append(result, current)
	}
	return result
}

// UniqueUpperStrings returns a slice of unique uppercase non-empty strings from the input.
func UniqueUpperStrings(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		current := strings.ToUpper(strings.TrimSpace(value))
		if current == "" {
			continue
		}
		if _, exists := seen[current]; exists {
			continue
		}
		seen[current] = struct{}{}
		result = append(result, current)
	}
	return result
}
