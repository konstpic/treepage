package llm

import (
	"strings"
)

// ExtractJSON pulls a JSON object from LLM output (markdown fences, preamble, etc.).
func ExtractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	s = stripThinkBlocks(s)

	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		rest = strings.TrimLeft(rest, " \t\r\n")
		if strings.HasPrefix(strings.ToLower(rest), "json") {
			rest = rest[4:]
			rest = strings.TrimLeft(rest, " \t\r\n")
		}
		if j := strings.Index(rest, "```"); j >= 0 {
			s = strings.TrimSpace(rest[:j])
		}
	}

	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return strings.TrimSpace(s)
}

func stripThinkBlocks(s string) string {
	// Angle-bracket reasoning tags only — never match markdown ``` fences.
	tags := [][2]string{
		{"\x3cthink\x3e", "\x3c/think\x3e"},
		{"\x3credacted_reasoning\x3e", "\x3c/redacted_reasoning\x3e"},
	}
	for _, pair := range tags {
		for {
			low := strings.ToLower(s)
			start := strings.Index(low, pair[0])
			if start < 0 {
				break
			}
			rest := low[start+len(pair[0]):]
			endRel := strings.Index(rest, pair[1])
			if endRel < 0 {
				s = s[:start]
				break
			}
			end := start + len(pair[0]) + endRel + len(pair[1])
			s = s[:start] + s[end:]
		}
	}
	return strings.TrimSpace(s)
}
