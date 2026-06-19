package service

import (
	"testing"
)

func TestLineDiff(t *testing.T) {
	a := []string{"line1", "line2", "line3"}
	b := []string{"line1", "line2 changed", "line3", "line4"}
	lines := lineDiff(a, b)
	var adds, removes int
	for _, l := range lines {
		switch l.Type {
		case "add":
			adds++
		case "remove":
			removes++
		}
	}
	if adds == 0 || removes == 0 {
		t.Fatalf("expected both adds and removes, got adds=%d removes=%d", adds, removes)
	}
}
