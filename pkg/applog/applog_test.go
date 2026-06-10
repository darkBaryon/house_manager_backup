package applog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"
)

// decodeLine 解析单行 JSON 日志，断言用。
func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected one log line, got empty output")
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("invalid json log line %q: %v", line, err)
	}
	return entry
}

func validPrincipal() session.Principal {
	return session.Principal{
		Terminal:      session.TerminalPublish,
		PrincipalType: session.PrincipalTypeLandlord,
		PrincipalID:   "landlord-1",
		Phone:         "13812345678",
	}
}

func TestMaskPhone(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"13812345678", "138****5678"},
		{"  13812345678  ", "138****5678"}, // TrimSpace 语义（采纳自 middleware 版本）
		{"123456", "123456"},               // 过短不脱敏
		{"", ""},
	}
	for _, c := range cases {
		if got := MaskPhone(c.in); got != c.want {
			t.Errorf("MaskPhone(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPhoneLogValue(t *testing.T) {
	if got := Phone("13812345678").LogValue().String(); got != "138****5678" {
		t.Errorf("Phone.LogValue() = %q, want %q", got, "138****5678")
	}
}

func TestPrincipalGroup(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	logger.Info("evt", PrincipalGroup(validPrincipal()))

	entry := decodeLine(t, &buf)
	principal, ok := entry["principal"].(map[string]any)
	if !ok {
		t.Fatalf("principal group missing or not an object: %v", entry)
	}
	if principal["terminal"] != "publish" || principal["type"] != "landlord" || principal["id"] != "landlord-1" {
		t.Errorf("unexpected principal fields: %v", principal)
	}
	if principal["phone"] != "138****5678" {
		t.Errorf("phone not masked: %v", principal["phone"])
	}
}

func TestPrincipalGroupWithoutPhone(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	p := validPrincipal()
	p.Phone = ""
	logger.Info("evt", PrincipalGroup(p))

	entry := decodeLine(t, &buf)
	principal := entry["principal"].(map[string]any)
	if _, exists := principal["phone"]; exists {
		t.Errorf("empty phone should be omitted: %v", principal)
	}
}

func TestContextHandlerInjection(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(ContextHandler(slog.NewJSONHandler(&buf, nil)))

	ctx := requestlog.ContextWithRequestID(context.Background(), "req-123")
	ctx = session.ContextWithPrincipal(ctx, validPrincipal())
	logger.InfoContext(ctx, "evt", "room_id", "r1")

	entry := decodeLine(t, &buf)
	if entry["request_id"] != "req-123" {
		t.Errorf("request_id not injected: %v", entry)
	}
	principal, ok := entry["principal"].(map[string]any)
	if !ok {
		t.Fatalf("principal group not injected: %v", entry)
	}
	if principal["id"] != "landlord-1" || principal["phone"] != "138****5678" {
		t.Errorf("unexpected principal: %v", principal)
	}
	if entry["room_id"] != "r1" {
		t.Errorf("business attr lost: %v", entry)
	}
}

func TestContextHandlerZeroInjection(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(ContextHandler(slog.NewJSONHandler(&buf, nil)))
	logger.InfoContext(context.Background(), "evt")

	entry := decodeLine(t, &buf)
	if _, exists := entry["request_id"]; exists {
		t.Errorf("request_id should not be injected: %v", entry)
	}
	if _, exists := entry["principal"]; exists {
		t.Errorf("principal should not be injected: %v", entry)
	}
}

// TestContextHandlerWithAttrsChain 验证 WithAttrs 透传后注入仍生效
// （pkg/logger.New 末尾的 logger.With(service/env) 正是这条链路）。
func TestContextHandlerWithAttrsChain(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(ContextHandler(slog.NewJSONHandler(&buf, nil))).With("service", "hm")

	ctx := requestlog.ContextWithRequestID(context.Background(), "req-456")
	logger.InfoContext(ctx, "evt")

	entry := decodeLine(t, &buf)
	if entry["service"] != "hm" {
		t.Errorf("base attr lost after WithAttrs: %v", entry)
	}
	if entry["request_id"] != "req-456" {
		t.Errorf("injection lost after WithAttrs: %v", entry)
	}
}

func setDefault(t *testing.T, opts *slog.HandlerOptions) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, opts)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestResult(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantLevel string
		wantMsg   string
		wantError bool
	}{
		{"success", nil, "INFO", "op.success", false},
		{"business failure -> warn", errcode.New(40001, "bad param"), "WARN", "op.failed", true},
		{"wrapped business failure -> warn", fmt.Errorf("wrap: %w", errcode.New(40400, "not found")), "WARN", "op.failed", true},
		{"server failure -> error", errcode.New(50001, "boom"), "ERROR", "op.failed", true},
		{"plain error -> error", errors.New("plain"), "ERROR", "op.failed", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := setDefault(t, nil)
			Result(context.Background(), "op.success", "op.failed", c.err, "k", "v")

			entry := decodeLine(t, buf)
			if entry["level"] != c.wantLevel {
				t.Errorf("level = %v, want %v", entry["level"], c.wantLevel)
			}
			if entry["msg"] != c.wantMsg {
				t.Errorf("msg = %v, want %v", entry["msg"], c.wantMsg)
			}
			if entry["k"] != "v" {
				t.Errorf("business attr lost: %v", entry)
			}
			if _, exists := entry["error"]; exists != c.wantError {
				t.Errorf("error attr presence = %v, want %v (entry: %v)", exists, c.wantError, entry)
			}
		})
	}
}

// TestResultSourcePointsToCaller 验证 caller skip：source 指向业务调用点而非 applog 内部。
func TestResultSourcePointsToCaller(t *testing.T) {
	buf := setDefault(t, &slog.HandlerOptions{AddSource: true})
	Result(context.Background(), "op.success", "op.failed", nil)

	entry := decodeLine(t, buf)
	source, ok := entry["source"].(map[string]any)
	if !ok {
		t.Fatalf("source missing: %v", entry)
	}
	file, _ := source["file"].(string)
	if !strings.HasSuffix(file, "applog_test.go") {
		t.Errorf("source.file = %q, want suffix applog_test.go (caller skip broken)", file)
	}
}

// TestResultDisabledShortCircuit 验证 Enabled 短路：级别被过滤时零输出。
func TestResultDisabledShortCircuit(t *testing.T) {
	buf := setDefault(t, &slog.HandlerOptions{Level: slog.LevelError})
	Result(context.Background(), "op.success", "op.failed", nil) // Info 被过滤

	if buf.Len() != 0 {
		t.Errorf("expected no output for filtered level, got %q", buf.String())
	}
}
