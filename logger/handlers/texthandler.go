package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vnafikov/foundation/logger/parameters"
)

const (
	initBufferSize = 1024 // 1 KiB.

	// See [infrastructure/http/server/router/middleware.traceStart].
	httpStartMessagePartCount = 3
	// See [infrastructure/http/server/router/middleware.traceEnd].
	httpEndMessagePartCount = 4

	// See [infrastructure/grpc/server/interceptors.traceStart].
	grpcStartMessagePartCount = 2
	// See [infrastructure/grpc/server/interceptors.traceEnd].
	grpcEndMessagePartCount = 3

	groupIndent = "  "

	timeLayout = "15:04:05.000"

	levelWarn  = "⚠︎"
	levelError = "♨︎"

	returnArrow = "↪"
)

type TextHandler struct {
	opts              slog.HandlerOptions
	preformattedAttrs string
	// openedGroupCount количество групп в preformattedAttrs.
	openedGroupCount int
	groups           []string
	w                io.Writer
	mu               *sync.Mutex
}

func (h *TextHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	var sb strings.Builder
	sb.Grow(initBufferSize)
	h.writeMessage(&sb, &r)
	sb.WriteString(h.attrsColor(r.Level))
	h.writeAttrs(&sb, &r)
	h.writeSource(&sb, &r)
	h.writeTimeAndLevel(&sb, &r)
	sb.WriteString(resetColor)

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := io.WriteString(h.w, sb.String())
	return err
}

func (h *TextHandler) writeMessage(sb *strings.Builder, r *slog.Record) {
	msg := strings.TrimSpace(r.Message)
	if h.opts.ReplaceAttr != nil {
		msg = h.opts.ReplaceAttr(h.groups, slog.String(slog.MessageKey, msg)).Value.String()
	}
	if msg == "" {
		return
	}

	sb.WriteString(h.messageColor(r.Level))
	switch {
	case strings.HasPrefix(msg, parameters.DebugErrorMessagePrefix):
		msg = h.formatDebugErrorMessage(msg)
	case strings.HasPrefix(msg, parameters.HTTPStartMessagePrefix):
		msg = h.formatHTTPStartMessage(msg)
	case strings.HasPrefix(msg, parameters.HTTPEndMessagePrefix):
		msg = h.formatHTTPEndMessage(msg)
	case strings.HasPrefix(msg, parameters.GRPCStartMessagePrefix):
		msg = h.formatGRPCStartMessage(msg)
	case strings.HasPrefix(msg, parameters.GRPCEndMessagePrefix):
		msg = h.formatGRPCEndMessage(msg)
	case strings.HasPrefix(msg, parameters.PGSQLMessagePrefix):
		msg = h.formatPGSQLMessage(msg)
	}
	sb.WriteString(msg + "\n")
}

func (*TextHandler) messageColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return cyan
	case slog.LevelWarn:
		return brightOrange
	case slog.LevelError:
		return brightRed
	}
	return ""
}

func (*TextHandler) formatDebugErrorMessage(msg string) string {
	return parameters.DebugErrorMessagePrefix + yellow + msg[len(parameters.DebugErrorMessagePrefix):]
}

func (h *TextHandler) formatHTTPStartMessage(msg string) string {
	parts := strings.Split(msg[len(parameters.HTTPStartMessagePrefix):], " ")
	if len(parts) != httpStartMessagePartCount {
		return msg
	}

	method := parts[1]
	uri := parts[2]
	return parameters.HTTPStartMessagePrefix + " " +
		h.httpMethodColor(method) + method + " " +
		brightBlue + uri
}

func (*TextHandler) httpMethodColor(method string) string {
	switch method {
	case http.MethodGet:
		return green
	case http.MethodPost, http.MethodPut:
		return brightOrange
	case http.MethodPatch:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodHead, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return brightBlue
	}
	return resetColor
}

func (h *TextHandler) formatHTTPEndMessage(msg string) string {
	parts := strings.Split(msg[len(parameters.HTTPEndMessagePrefix):], " ")
	if len(parts) != httpEndMessagePartCount {
		return msg
	}

	method := parts[1]
	statusCode := parts[2]
	uri := parts[3]
	return parameters.HTTPEndMessagePrefix + " " +
		h.httpMethodColor(method) + method + " " +
		h.httpStatusColor(statusCode) + statusCode + " " +
		brightBlue + uri
}

func (*TextHandler) httpStatusColor(statusCode string) string {
	switch statusCode[0] {
	case '1':
		return brightBlue
	case '2':
		return green
	case '3':
		return yellow
	case '4':
		return brightOrange
	case '5':
		return red
	}
	return resetColor
}

func (*TextHandler) formatGRPCStartMessage(msg string) string {
	parts := strings.Split(msg[len(parameters.GRPCStartMessagePrefix):], " ")
	if len(parts) != grpcStartMessagePartCount {
		return msg
	}

	fullMethod := parts[1]
	return parameters.GRPCStartMessagePrefix + " " +
		brightBlue + fullMethod
}

func (h *TextHandler) formatGRPCEndMessage(msg string) string {
	parts := strings.Split(msg[len(parameters.GRPCEndMessagePrefix):], " ")
	if len(parts) != grpcEndMessagePartCount {
		return msg
	}

	statusCode := parts[1]
	fullMethod := parts[2]
	return parameters.GRPCEndMessagePrefix + " " +
		h.grpcStatusColor(statusCode) + statusCode + " " +
		brightBlue + fullMethod
}

func (*TextHandler) grpcStatusColor(statusCode string) string {
	switch statusCode {
	case "OK":
		return green
	case "Canceled", "Aborted":
		return yellow
	case "InvalidArgument",
		"NotFound",
		"AlreadyExists",
		"PermissionDenied",
		"ResourceExhausted",
		"FailedPrecondition",
		"OutOfRange",
		"Unauthenticated":
		return brightOrange
	case "Unknown",
		"DeadlineExceeded",
		"Unimplemented",
		"Internal",
		"Unavailable",
		"DataLoss":
		return red
	}
	return resetColor
}

func (*TextHandler) formatPGSQLMessage(msg string) string {
	return parameters.PGSQLMessagePrefix + blue + msg[len(parameters.PGSQLMessagePrefix):]
}

func (*TextHandler) attrsColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return cyan
	case slog.LevelInfo, parameters.LevelImportant:
		return springGreen
	case slog.LevelWarn:
		return tangerine
	case slog.LevelError:
		return red
	}
	return resetColor
}

func (h *TextHandler) writeAttrs(sb *strings.Builder, r *slog.Record) {
	if h.preformattedAttrs != "" {
		sb.WriteString(h.preformattedAttrs)
	}
	if r.NumAttrs() == 0 {
		return
	}

	indent := h.writeGroupKeys(sb)
	r.Attrs(func(attr slog.Attr) bool {
		h.writeAttr(sb, h.groups, attr, indent)
		return true
	})
}

func (h *TextHandler) writeGroupKeys(sb *strings.Builder) (indent string) {
	indent = strings.Repeat(groupIndent, h.openedGroupCount)
	for _, group := range h.groups[h.openedGroupCount:] {
		indent = h.writeGroupKey(sb, group, indent)
	}
	return indent
}

func (*TextHandler) writeGroupKey(sb *strings.Builder, group, indent string) (newIndent string) {
	sb.WriteString(indent + group + ":\n")
	return indent + groupIndent
}

func (h *TextHandler) writeAttr(sb *strings.Builder, groups []string, attr slog.Attr, indent string) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		h.writeGroup(sb, groups, attr, indent)
		return
	}

	if h.opts.ReplaceAttr != nil {
		attr = h.opts.ReplaceAttr(groups, attr)
		if attr.Equal(slog.Attr{}) {
			return
		}
	}
	if attr.Key == parameters.StackKey {
		if sources, ok := attr.Value.Any().([]slog.Source); ok {
			h.writeStack(sb, attr, sources, indent)
			return
		}
	}

	q := strconv.Quote(attr.Value.String())
	s := indent + attr.Key + ": " + q[1:len(q)-1]
	sb.WriteString(s + "\n")
}

func (h *TextHandler) writeGroup(sb *strings.Builder, groups []string, attr slog.Attr, indent string) {
	if attr.Key != "" {
		groups = append(groups, attr.Key)
		indent = h.writeGroupKey(sb, attr.Key, indent)
	}
	for _, groupAttr := range attr.Value.Group() {
		h.writeAttr(sb, groups, groupAttr, indent)
	}
}

func (h *TextHandler) writeStack(sb *strings.Builder, attr slog.Attr, sources []slog.Source, indent string) {
	indent = h.writeGroupKey(sb, attr.Key, indent)
	for _, source := range sources {
		src := indent + source.Function + ": " + source.File + ":" + strconv.Itoa(source.Line)
		sb.WriteString(src + "\n")
	}
}

func (h *TextHandler) writeSource(sb *strings.Builder, r *slog.Record) {
	if !h.opts.AddSource {
		return
	}

	frames := runtime.CallersFrames([]uintptr{r.PC})
	frame, _ := frames.Next()
	var src string
	if h.opts.ReplaceAttr != nil {
		source := &slog.Source{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		}
		value := h.opts.ReplaceAttr(h.groups, slog.Any(slog.SourceKey, source)).Value
		source, ok := value.Any().(*slog.Source)
		if ok && source.Function != "" {
			src = source.Function + ": " + source.File + ":" + strconv.Itoa(source.Line)
		} else {
			src = value.String()
		}
	} else if frame.Function != "" {
		src = frame.Function + ": " + frame.File + ":" + strconv.Itoa(frame.Line)
	}
	if src == "" {
		return
	}

	sb.WriteString(src + "\n")
}

func (h *TextHandler) writeTimeAndLevel(sb *strings.Builder, r *slog.Record) {
	t := h.formatTime(r)
	level := h.formatLevel(r)
	if t == "" && level == "" {
		return
	}

	var s string
	if t != "" && level != "" {
		s = t + " " + level
	} else {
		s = t + level
	}
	s = "  " + h.timeAndLevelColor(r.Level) + returnArrow + s
	sb.WriteString(s + "\n")
}

func (h *TextHandler) formatTime(r *slog.Record) string {
	if r.Time.IsZero() {
		return ""
	}

	t := r.Time
	if h.opts.ReplaceAttr != nil {
		value := h.opts.ReplaceAttr(h.groups, slog.Time(slog.TimeKey, r.Time)).Value
		var ok bool
		if t, ok = value.Any().(time.Time); !ok {
			return value.String()
		}
	}
	return t.Format(timeLayout)
}

func (h *TextHandler) formatLevel(r *slog.Record) string {
	if r.Level < slog.LevelWarn || r.Level == parameters.LevelImportant {
		return ""
	}
	if h.opts.ReplaceAttr != nil {
		return h.opts.ReplaceAttr(h.groups, slog.Any(slog.LevelKey, r.Level)).Value.String()
	}

	switch r.Level {
	case slog.LevelWarn:
		return levelWarn
	case slog.LevelError:
		return levelError
	}
	return r.Level.String()
}

func (*TextHandler) timeAndLevelColor(level slog.Level) string {
	if level == slog.LevelDebug {
		return magenta
	}
	return ""
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := h.clone()
	newHandler.appendPreformattedAttrs(attrs)
	return newHandler
}

func (h *TextHandler) clone() *TextHandler {
	return &TextHandler{
		opts:              h.opts,
		preformattedAttrs: h.preformattedAttrs,
		openedGroupCount:  h.openedGroupCount,
		groups:            h.groups,
		w:                 h.w,
		mu:                h.mu,
	}
}

func (h *TextHandler) appendPreformattedAttrs(attrs []slog.Attr) {
	var sb strings.Builder
	sb.Grow(initBufferSize)
	indent := h.writeGroupKeys(&sb)
	h.openedGroupCount += len(h.groups[h.openedGroupCount:])
	for _, attr := range attrs {
		h.writeAttr(&sb, h.groups, attr, indent)
	}
	h.preformattedAttrs += sb.String()
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	newHandler := h.clone()
	newHandler.groups = append(newHandler.groups, name)
	return newHandler
}

func NewTextHandler(w io.Writer, opts *slog.HandlerOptions) *TextHandler {
	if opts == nil {
		opts = new(slog.HandlerOptions)
	}
	return &TextHandler{
		opts: *opts,
		w:    w,
		mu:   new(sync.Mutex),
	}
}
