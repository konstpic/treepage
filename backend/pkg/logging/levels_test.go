package logging

import (
	"testing"

	"gorm.io/gorm/logger"
)

func TestGormLogLevel(t *testing.T) {
	tests := map[string]logger.LogLevel{
		"debug": logger.Info,
		"info":  logger.Warn,
		"warn":  logger.Error,
		"error": logger.Silent,
	}
	for level, want := range tests {
		if got := GormLogLevel(level); got != want {
			t.Fatalf("GormLogLevel(%q) = %v, want %v", level, got, want)
		}
	}
}

func TestShouldLogRequest(t *testing.T) {
	if !ShouldLogRequest("debug", 200) {
		t.Fatal("debug should log 200")
	}
	if ShouldLogRequest("info", 200) {
		t.Fatal("info should skip 200")
	}
	if !ShouldLogRequest("info", 404) {
		t.Fatal("info should log 404")
	}
	if ShouldLogRequest("warn", 404) {
		t.Fatal("warn should skip 404")
	}
	if !ShouldLogRequest("warn", 500) {
		t.Fatal("warn should log 500")
	}
}
