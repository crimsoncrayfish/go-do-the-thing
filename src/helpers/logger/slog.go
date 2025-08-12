package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

const (
	DEV  = "development"
	PROD = "production"
)

const LevelTODO = slog.Level(-8) // Lower than Debug (-4)

const (
	colorRed    = "\033[0;31m"
	colorYellow = "\033[0;33m"
	colorPurple = "\033[0;35m"
	colorNone   = "\033[0m"
)

type ColoredHandler struct {
	opts         slog.HandlerOptions
	debugEnabled bool
	attrs        []slog.Attr
	groups       []string
}

func NewColoredHandler(opts *slog.HandlerOptions) *ColoredHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	env := os.Getenv("ENV")
	debugEnabled := env == DEV || env == ""

	return &ColoredHandler{
		opts:         *opts,
		debugEnabled: debugEnabled,
	}
}

func (h *ColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.debugEnabled && (r.Level == slog.LevelDebug || r.Level == LevelTODO) {
		return nil
	}

	var color string
	var levelStr string

	switch r.Level {
	case slog.LevelError:
		color = colorRed
		levelStr = "ERROR"
	case slog.LevelWarn:
		color = colorYellow
		levelStr = "WARN"
	case slog.LevelInfo:
		levelStr = "INFO"
	case slog.LevelDebug:
		color = colorYellow
		levelStr = "DEBUG"
	case LevelTODO:
		color = colorPurple
		levelStr = "TODO"
	default:
		levelStr = r.Level.String()
	}

	var callerInfo string
	if r.Level >= slog.LevelError {
		if pc := r.PC; pc != 0 {
			fs := runtime.CallersFrames([]uintptr{pc})
			f, _ := fs.Next()
			if f.File != "" {
				// Get short filename
				parts := strings.Split(f.File, "/")
				shortFile := parts[len(parts)-1]
				callerInfo = fmt.Sprintf(" %s:%d", shortFile, f.Line)
			}
		}
	}

	var source string
	var otherAttrs []slog.Attr
	for _, attr := range h.attrs {
		if attr.Key == "source" {
			source = attr.Value.String()
		} else {
			otherAttrs = append(otherAttrs, attr)
		}
	}
	var attrStr strings.Builder
	for _, attr := range otherAttrs {
		if attrStr.Len() > 0 {
			attrStr.WriteString(" ")
		}
		attrStr.WriteString(fmt.Sprintf("%s=%v", attr.Key, attr.Value))
	}

	r.Attrs(func(a slog.Attr) bool {
		if attrStr.Len() > 0 {
			attrStr.WriteString(" ")
		}
		attrStr.WriteString(fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})

	timestamp := r.Time.Format("2006-01-02 15:04:05")
	message := r.Message
	if callerInfo != "" {
		message += callerInfo
	}

	// TIMESTAMP - LEVEL - SOURCE - MESSAGE - ATTRIBUTES
	var logLine string
	if source != "" {
		if attrStr.Len() > 0 {
			logLine = fmt.Sprintf("%s - %s - %s - %s - %s", timestamp, levelStr, source, message, attrStr.String())
		} else {
			logLine = fmt.Sprintf("%s - %s - %s - %s", timestamp, levelStr, source, message)
		}
	} else {
		if attrStr.Len() > 0 {
			logLine = fmt.Sprintf("%s - %s - %s - %s", timestamp, levelStr, message, attrStr.String())
		} else {
			logLine = fmt.Sprintf("%s - %s - %s", timestamp, levelStr, message)
		}
	}

	if color != "" {
		fmt.Printf("%s%s%s\n", color, logLine, colorNone)
	} else {
		fmt.Printf("%s\n", logLine)
	}
	return nil
}

func (h *ColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColoredHandler{
		opts:         h.opts,
		debugEnabled: h.debugEnabled,
		attrs:        append(h.attrs, attrs...),
		groups:       h.groups,
	}
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return &ColoredHandler{
		opts:         h.opts,
		debugEnabled: h.debugEnabled,
		attrs:        h.attrs,
		groups:       append(h.groups, name),
	}
}

func SetupLogger() {
	env := os.Getenv("ENV")

	if env == PROD {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		slog.SetDefault(slog.New(handler))
	} else {
		handler := NewColoredHandler(&slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(handler))
	}
}

func TODO(msg string, args ...any) {
	slog.Log(context.Background(), LevelTODO, msg, args...)
}

func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	panic(1)
}
