package fts

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	wordRe = regexp.MustCompile(`[\p{L}\p{N}]+`)
	stopWords = map[string]struct{}{
		// Russian
		"как": {}, "что": {}, "где": {}, "когда": {}, "почему": {}, "зачем": {}, "какой": {}, "какая": {}, "какие": {}, "какое": {},
		"это": {}, "этот": {}, "эта": {}, "эти": {}, "тот": {}, "та": {}, "те": {},
		"ли": {}, "же": {}, "бы": {}, "не": {}, "ни": {}, "из": {}, "на": {}, "по": {}, "за": {}, "от": {}, "до": {}, "при": {},
		"в": {}, "с": {}, "у": {}, "о": {}, "об": {}, "и": {}, "или": {}, "а": {}, "но": {}, "да": {}, "нет": {}, "мне": {}, "мой": {}, "моя": {}, "мои": {},
		"можно": {}, "нужно": {}, "надо": {}, "есть": {}, "был": {}, "была": {}, "были": {},
		"кто": {}, "чем": {},
		// English
		"how": {}, "what": {}, "where": {}, "when": {}, "why": {}, "who": {}, "which": {},
		"the": {}, "a": {}, "an": {}, "is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "been": {},
		"to": {}, "do": {}, "does": {}, "did": {}, "can": {}, "could": {}, "should": {}, "would": {},
		"i": {}, "me": {}, "my": {}, "we": {}, "you": {}, "your": {}, "it": {}, "its": {}, "this": {}, "that": {},
		"for": {}, "of": {}, "in": {}, "on": {}, "at": {}, "with": {}, "from": {}, "or": {}, "and": {},
	}
)

// ExtractKeywords pulls meaningful tokens from a natural-language question.
func ExtractKeywords(question string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, w := range wordRe.FindAllString(strings.ToLower(question), -1) {
		if len([]rune(w)) < 3 {
			continue
		}
		if _, stop := stopWords[w]; stop {
			continue
		}
		if _, dup := seen[w]; dup {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
	}
	return out
}

// KeywordPrefix returns a prefix for fuzzy ILIKE matching (деплоить → депл, установить → уста).
func KeywordPrefix(keyword string) string {
	runes := []rune(keyword)
	n := len(runes)
	if n <= 4 {
		return keyword
	}
	if n > 7 {
		n = 5
	} else {
		n = 4
	}
	return string(runes[:n])
}

// HasCyrillic reports whether the text contains Cyrillic letters.
func HasCyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}
