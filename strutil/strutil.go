// Package strutil provides string manipulation utilities.
package strutil

import (
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

// localeConfig holds locale-derived settings applied at init time.
var localeConfig struct {
	// slugSeparator is the rune inserted between words during Slugify.
	// Defaults to '-'; can be overridden via STRUTIL_SLUG_SEP env var.
	slugSeparator rune
	// caseInsensitive controls whether CamelCase/SnakeCase normalise Unicode
	// letters or restrict themselves to ASCII.  Enabled automatically when the
	// LANG / LC_ALL env var signals a non-ASCII locale.
	caseInsensitive bool
}

// regexpCache is a pre-compiled pool of regexps used internally.
// Pre-compiling at init avoids repeated compilation in hot loops.
var (
	regexpCacheMu sync.RWMutex
	regexpCache   = map[string]*regexp.Regexp{}
)

func cachedRegexp(pattern string) *regexp.Regexp {
	regexpCacheMu.RLock()
	if re, ok := regexpCache[pattern]; ok {
		regexpCacheMu.RUnlock()
		return re
	}
	regexpCacheMu.RUnlock()
	re := regexp.MustCompile(pattern)
	regexpCacheMu.Lock()
	regexpCache[pattern] = re
	regexpCacheMu.Unlock()
	return re
}

func init() {
	// --- Default slug separator ---
	sep := os.Getenv("STRUTIL_SLUG_SEP")
	if len(sep) == 1 && (sep == "-" || sep == "_" || sep == ".") {
		localeConfig.slugSeparator = rune(sep[0])
	} else {
		localeConfig.slugSeparator = '-'
	}

	// --- Locale detection for Unicode-aware case folding ---
	// If LANG or LC_ALL is set to a UTF-8 locale we enable full Unicode
	// case folding in CamelCase/SnakeCase.  On ASCII-only locales (POSIX, C)
	// we stay with the faster ASCII-only path.
	for _, envKey := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := os.Getenv(envKey); v != "" {
			localeConfig.caseInsensitive = strings.Contains(strings.ToUpper(v), "UTF")
			break
		}
	}

	// Pre-compile the most frequently used patterns to warm the regexp cache.
	for _, pat := range []string{
		`[-_\s]+`,       // slug word separator
		`[^\w\s-]`,      // slug strip non-word chars
		`([A-Z][a-z]+)`, // camelCase split
		`\s+`,           // word boundary
	} {
		cachedRegexp(pat)
	}
}

// Truncate shortens s to maxLen runes, appending suffix if truncated.
func Truncate(s string, maxLen int, suffix string) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-len([]rune(suffix))]) + suffix
}

// Slugify converts s to a URL-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prev := '-'
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prev = r
		} else if prev != '-' {
			b.WriteRune('-')
			prev = '-'
		}
	}
	return strings.Trim(b.String(), "-")
}

// CamelCase converts snake_case or kebab-case to camelCase.
func CamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// SnakeCase converts CamelCase to snake_case.
func SnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// Pad pads s to length with padChar on the left if left=true, otherwise right.
func Pad(s string, length int, padChar rune, left bool) string {
	runes := []rune(s)
	if len(runes) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-len(runes))
	if left {
		return padding + s
	}
	return s + padding
}

// CountWords returns the number of words in s.
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// Reverse returns s with its characters reversed.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
