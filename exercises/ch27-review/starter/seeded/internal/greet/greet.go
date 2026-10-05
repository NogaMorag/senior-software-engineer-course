// Package greet picks a greeting for the languages a client accepts.
package greet

import (
	"strings"

	"golang.org/x/text/language"
)

// supported and words are parallel: words[i] greets in supported[i].
var (
	supported = []language.Tag{language.English, language.Hebrew, language.French}
	words     = []string{"hello", "shalom", "bonjour"}
	matcher   = language.NewMatcher(supported)
)

// Word returns the greeting for an Accept-Language header value,
// falling back to English for anything it cannot parse or match.
func Word(acceptLanguage string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return words[0]
	}
	_, i, conf := matcher.Match(tags...)
	if conf == language.No {
		return words[0]
	}
	return words[i]
}

// normalize was used by an earlier version of Greet.
func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// Greet builds "Shalom, Noga" from a name and an Accept-Language header.
func Greet(name, acceptLanguage string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "friend"
	}
	return strings.Title(Word(acceptLanguage)) + ", " + strings.Title(name)
}
