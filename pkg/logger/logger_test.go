package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    slog.Level
		wantErr bool
	}{
		{name: "empty defaults to info", input: "", want: slog.LevelInfo},
		{name: "debug", input: "debug", want: slog.LevelDebug},
		{name: "info", input: "info", want: slog.LevelInfo},
		{name: "warn", input: "warn", want: slog.LevelWarn},
		{name: "warning", input: "warning", want: slog.LevelWarn},
		{name: "error", input: "error", want: slog.LevelError},
		{name: "invalid", input: "xxx", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
	}{
		{name: "empty defaults to text", input: "", want: FormatText},
		{name: "text", input: "text", want: FormatText},
		{name: "json", input: "json", want: FormatJSON},
		{name: "invalid", input: "xml", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew_JSONOutput(t *testing.T) {
	var buf bytes.Buffer

	l, err := New(Config{
		Level:   "info",
		Format:  "json",
		Service: "user-api",
		Env:     "test",
	}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.Info("hello", "user_id", 123)

	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatalf("expected log output, got empty string")
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("failed to parse json log: %v, raw=%q", err, out)
	}

	if m["msg"] != "hello" {
		t.Fatalf("msg = %v, want %q", m["msg"], "hello")
	}
	if m["level"] != "INFO" {
		t.Fatalf("level = %v, want %q", m["level"], "INFO")
	}
	if m["service"] != "user-api" {
		t.Fatalf("service = %v, want %q", m["service"], "user-api")
	}
	if m["env"] != "test" {
		t.Fatalf("env = %v, want %q", m["env"], "test")
	}

	// JSON 数字会解成 float64
	if got := m["user_id"]; got != float64(123) {
		t.Fatalf("user_id = %v, want %v", got, 123)
	}
}

func TestNew_TextOutput(t *testing.T) {
	var buf bytes.Buffer

	l, err := New(Config{
		Level:   "debug",
		Format:  "text",
		Service: "user-api",
		Env:     "test",
	}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.Debug("hello", "user_id", 123)

	out := buf.String()
	if out == "" {
		t.Fatalf("expected log output, got empty string")
	}

	mustContain := []string{
		"DEBUG",
		"hello",
		"service=user-api",
		"env=test",
		"user_id=123",
	}

	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Fatalf("output %q does not contain %q", out, s)
		}
	}
}

func TestNew_TextOutput_WithSource(t *testing.T) {
	var buf bytes.Buffer

	l, err := New(Config{
		Level:     "info",
		Format:    "text",
		AddSource: true,
	}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.Info("with source")

	out := buf.String()
	if out == "" {
		t.Fatalf("expected log output, got empty string")
	}

	if !strings.Contains(out, "pkg/logger/logger_test.go:") {
		t.Fatalf("output %q does not contain source location", out)
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	var buf bytes.Buffer

	_, err := New(Config{
		Level:  "bad-level",
		Format: "json",
	}, &buf)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer

	_, err := New(Config{
		Level:  "info",
		Format: "bad-format",
	}, &buf)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestInit_SetsDefaultLogger(t *testing.T) {
	var buf bytes.Buffer

	err := Init(Config{
		Level:   "info",
		Format:  "json",
		Service: "user-api",
		Env:     "test",
	}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 注意：这会使用全局默认 logger，所以不要并行跑这个测试
	slog.Info("hello from default")

	out := buf.String()
	if !strings.Contains(out, "hello from default") {
		t.Fatalf("expected default logger output, got %q", out)
	}
}
