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
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
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

	attrs := make([]string, 0, rec.NumAttrs()+len(h.attrs))
	for _, attr := range h.attrs {
		attrs = appendAttr(attrs, h.groups, attr)
	}
	rec.Attrs(func(attr slog.Attr) bool {
		attrs = appendAttr(attrs, h.groups, attr)
		return true
	})

	if len(attrs) > 0 {
		buf.WriteByte(' ')
		buf.WriteString(strings.Join(attrs, " "))
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.writer.Write(buf.Bytes())
	return err
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
	value := t.Format("15:04:05.000")
	if h.color {
		return colorDim + value + colorReset
	}
	return value
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

func appendAttr(dst []string, groups []string, attr slog.Attr) []string {
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
			dst = appendAttr(dst, nextGroups, groupAttr)
		}
		return dst
	}

	key := attr.Key
	if len(groups) > 0 {
		key = strings.Join(append(append([]string{}, groups...), key), ".")
	}

	dst = append(dst, key+"="+formatValue(attr.Value))
	return dst
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
		return text
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.FormatBool(value.Bool())
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339Nano)
	case slog.KindAny:
		return marshalAny(value.Any())
	default:
		return fmt.Sprint(value.Any())
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
