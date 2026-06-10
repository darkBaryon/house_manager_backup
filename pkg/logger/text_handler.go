package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	colorReset   = "\033[0m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorCyan    = "\033[36m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorBold    = "\033[1m"
)

type prettyTextHandler struct {
	opts   slog.HandlerOptions
	writer io.Writer
	color  bool

	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

func newPrettyTextHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	handlerOpts := slog.HandlerOptions{}
	if opts != nil {
		handlerOpts = *opts
	}

	return &prettyTextHandler{
		opts:   handlerOpts,
		writer: w,
		color:  supportsColor(w),
		mu:     &sync.Mutex{},
	}
}

func (h *prettyTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle 的展示层约定（仅影响本地 text 输出，JSON/生产输出不受影响）：
//   - 进程级常量字段（service/env）不渲染——本地恒定，无信息量；
//   - 请求上下文字段（request_id、principal.*）压缩成调暗的尾巴，让业务字段突出；
//   - 超长十六进制 ID（如 Mongo ObjectID）截断为「前4…后4」显示，
//     需要完整 ID 时设置环境变量 LOG_FULL_IDS=1 或切 format=json。
func (h *prettyTextHandler) Handle(_ context.Context, rec slog.Record) error {
	var buf bytes.Buffer

	buf.WriteString(h.formatTime(rec.Time))
	buf.WriteByte(' ')
	buf.WriteString(h.formatLevel(rec.Level))

	if h.opts.AddSource && rec.PC != 0 {
		if src := formatSource(rec.PC, h.color); src != "" {
			buf.WriteByte(' ')
			buf.WriteString(src)
		}
	}

	buf.WriteByte(' ')
	buf.WriteString(h.formatMessage(rec.Message))

	pairs := make([]attrPair, 0, rec.NumAttrs()+len(h.attrs))
	for _, attr := range h.attrs {
		pairs = collectAttr(pairs, h.groups, attr)
	}
	rec.Attrs(func(attr slog.Attr) bool {
		pairs = collectAttr(pairs, h.groups, attr)
		return true
	})

	business, contextTail := h.renderAttrs(pairs)
	if len(business) > 0 {
		buf.WriteByte(' ')
		buf.WriteString(strings.Join(business, " "))
	}
	if contextTail != "" {
		buf.WriteByte(' ')
		buf.WriteString(contextTail)
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.writer.Write(buf.Bytes())
	return err
}

type attrPair struct {
	key   string
	value slog.Value
}

// renderAttrs 把扁平化的键值对分成「业务字段」与「请求上下文尾巴」两段。
func (h *prettyTextHandler) renderAttrs(pairs []attrPair) (business []string, contextTail string) {
	var requestID string
	principal := map[string]string{}

	for _, p := range pairs {
		switch {
		case p.key == "service" || p.key == "env":
			continue // 进程级常量，本地无信息量
		case p.key == "request_id":
			requestID = p.value.Resolve().String()
		case strings.HasPrefix(p.key, "principal."):
			principal[strings.TrimPrefix(p.key, "principal.")] = p.value.Resolve().String()
		default:
			business = append(business, h.renderBusinessAttr(p))
		}
	}

	tail := make([]string, 0, 2)
	if requestID != "" {
		if len(requestID) > 8 {
			requestID = requestID[:8]
		}
		tail = append(tail, "req="+requestID)
	}
	if len(principal) > 0 {
		parts := make([]string, 0, 4)
		for _, k := range []string{"terminal", "type", "id", "phone"} {
			if v := principal[k]; v != "" {
				parts = append(parts, shortHexID(v))
			}
		}
		tail = append(tail, "who="+strings.Join(parts, "/"))
	}
	if len(tail) == 0 {
		return business, ""
	}

	joined := strings.Join(tail, " ")
	if h.color {
		return business, colorMagenta + joined + colorReset
	}
	return business, joined
}

func (h *prettyTextHandler) renderBusinessAttr(p attrPair) string {
	value := formatValue(p.value)
	if !h.color {
		return p.key + "=" + value
	}
	if p.key == "error" {
		return colorRed + p.key + "=" + value + colorReset
	}
	// 键名青色、值默认色，业务数据在视觉上浮起（不用 dim，浅色背景下不可见）
	return colorCyan + p.key + "=" + colorReset + value
}

func (h *prettyTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := h.clone()
	next.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return next
}

func (h *prettyTextHandler) WithGroup(name string) slog.Handler {
	next := h.clone()
	next.groups = append(append([]string{}, h.groups...), name)
	return next
}

func (h *prettyTextHandler) clone() *prettyTextHandler {
	return &prettyTextHandler{
		opts:   h.opts,
		writer: h.writer,
		color:  h.color,
		mu:     h.mu,
		attrs:  h.attrs,
		groups: h.groups,
	}
}

func supportsColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if term := os.Getenv("TERM"); term == "" || term == "dumb" {
		return false
	}

	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func (h *prettyTextHandler) formatTime(t time.Time) string {
	// 不上色：dim/灰色在浅色或图片背景的终端里不可见
	return t.Format("15:04:05.000")
}

func (h *prettyTextHandler) formatLevel(level slog.Level) string {
	label := "INFO"
	color := colorGreen

	switch {
	case level <= slog.LevelDebug:
		label = "DEBUG"
		color = colorCyan
	case level < slog.LevelWarn:
		label = "INFO"
		color = colorGreen
	case level < slog.LevelError:
		label = "WARN"
		color = colorYellow
	default:
		label = "ERROR"
		color = colorRed
	}

	if h.color {
		return color + label + colorReset
	}
	return label
}

func (h *prettyTextHandler) formatMessage(msg string) string {
	if h.color {
		return colorBold + msg + colorReset
	}
	return msg
}

func formatSource(pc uintptr, color bool) string {
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	if frame.File == "" {
		return ""
	}

	source := fmt.Sprintf("%s:%d", shortSourcePath(frame.File), frame.Line)
	if color {
		return colorBlue + source + colorReset
	}
	return source
}

func shortSourcePath(path string) string {
	cleaned := filepath.ToSlash(path)
	parts := strings.Split(cleaned, "/")
	if len(parts) <= 3 {
		return cleaned
	}
	return strings.Join(parts[len(parts)-3:], "/")
}

func collectAttr(dst []attrPair, groups []string, attr slog.Attr) []attrPair {
	attr.Value = attr.Value.Resolve()
	if attr.Key == "" && attr.Value.Kind() == slog.KindAny && attr.Value.Any() == nil {
		return dst
	}

	if attr.Value.Kind() == slog.KindGroup {
		nextGroups := groups
		if attr.Key != "" {
			nextGroups = append(append([]string{}, groups...), attr.Key)
		}
		for _, groupAttr := range attr.Value.Group() {
			dst = collectAttr(dst, nextGroups, groupAttr)
		}
		return dst
	}

	key := attr.Key
	if len(groups) > 0 {
		key = strings.Join(append(append([]string{}, groups...), key), ".")
	}

	return append(dst, attrPair{key: key, value: attr.Value})
}

// shortHexID 把超长十六进制串（Mongo ObjectID 等）截断为「前4…后4」。
// 设置 LOG_FULL_IDS=1 时关闭截断（需要完整 ID 复制粘贴去查库的场景）。
func shortHexID(s string) string {
	if os.Getenv("LOG_FULL_IDS") != "" {
		return s
	}
	if len(s) < 20 {
		return s
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '-') {
			return s
		}
	}
	return s[:4] + "…" + s[len(s)-4:]
}

func formatValue(value slog.Value) string {
	value = value.Resolve()

	switch value.Kind() {
	case slog.KindString:
		text := value.String()
		if text == "" {
			return `""`
		}
		if strings.ContainsAny(text, " \t\n\"") {
			return strconv.Quote(text)
		}
		return shortHexID(text)
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.FormatBool(value.Bool())
	case slog.KindDuration:
		return formatDuration(value.Duration())
	case slog.KindTime:
		return value.Time().Format(time.RFC3339Nano)
	case slog.KindAny:
		return marshalAny(value.Any())
	default:
		return fmt.Sprint(value.Any())
	}
}

// formatDuration 按量级取整到两位小数（3.353826167s → 3.35s），只影响 text 显示。
func formatDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return d.Round(10 * time.Millisecond).String()
	case d >= time.Millisecond:
		return d.Round(10 * time.Microsecond).String()
	case d >= time.Microsecond:
		return d.Round(10 * time.Nanosecond).String()
	default:
		return d.String()
	}
}

func marshalAny(v any) string {
	if v == nil {
		return "null"
	}
	if err, ok := v.(error); ok {
		return strconv.Quote(err.Error())
	}
	if stringer, ok := v.(fmt.Stringer); ok {
		return strconv.Quote(stringer.String())
	}

	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}
