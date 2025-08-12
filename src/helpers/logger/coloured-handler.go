package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

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

func newColoredHandler(opts *slog.HandlerOptions) *ColoredHandler {
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

func (h *ColoredHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.debugEnabled && (record.Level == slog.LevelDebug || record.Level == LevelTODO) {
		return nil
	}

	var color string
	var levelStr string

	switch record.Level {
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
		levelStr = record.Level.String()
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

	record.Attrs(func(a slog.Attr) bool {
		if attrStr.Len() > 0 {
			attrStr.WriteString(" ")
		}
		attrStr.WriteString(fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})

	timestamp := record.Time.Format("2006-01-02 15:04:05")
	message := record.Message

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
