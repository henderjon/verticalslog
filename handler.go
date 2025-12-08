package verticalslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Handler implements slog.Handler with two-column vertical output.
// The left column contains attribute names (right-aligned) and the
// right column contains values (left-aligned) with proper wrapping.
type Handler struct {
	w          io.Writer
	mu         sync.Mutex
	leftWidth  int
	rightWidth int
	opts       Options
}

// Options configures the Handler behavior.
type Options struct {
	// LeftWidth is the width of the left column (default: 20)
	LeftWidth int

	// RightWidth is the maximum width of the right column before wrapping (default: 60)
	RightWidth int

	// Level is the minimum log level to output (default: Info)
	Level slog.Level

	// AddSource adds source file information with full path (default: false)
	AddSource bool

	// AddSourceShort adds source file information with only filename, not full path (default: false)
	// If both AddSource and AddSourceShort are true, AddSourceShort takes precedence
	AddSourceShort bool
}

// NewHandler creates a new two-column text handler.
func NewHandler(w io.Writer, opts *Options) *Handler {
	h := &Handler{
		w:          w,
		leftWidth:  20,
		rightWidth: 60,
	}

	if opts != nil {
		h.opts = *opts
		if opts.LeftWidth > 0 {
			h.leftWidth = opts.LeftWidth
		}
		if opts.RightWidth > 0 {
			h.rightWidth = opts.RightWidth
		}
	}

	return h
}

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level
}

// Handle handles the Record.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var buf strings.Builder

	// Write time
	h.writeRow(&buf, "time", r.Time.Format(time.RFC3339))

	// Write level
	h.writeRow(&buf, "level", r.Level.String())

	// Write message
	h.writeRow(&buf, "message", r.Message)

	// Write source if enabled
	if (h.opts.AddSource || h.opts.AddSourceShort) && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := fs.Next()
		sourceFile := frame.File
		if h.opts.AddSourceShort {
			sourceFile = h.shortPath(sourceFile)
		}
		h.writeRow(&buf, "source", fmt.Sprintf("%s:%d", sourceFile, frame.Line))
	}

	// Write attributes
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(&buf, a)
		return true
	})

	// Write separator
	h.writeSeparator(&buf)

	_, err := h.w.Write([]byte(buf.String()))
	return err
}

// WithAttrs returns a new Handler whose attributes consist of both h's attributes and attrs.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, this implementation doesn't support attribute groups
	// A full implementation would need to track attributes across handlers
	return h
}

// WithGroup returns a new Handler with the given group appended to the receiver's groups.
func (h *Handler) WithGroup(name string) slog.Handler {
	// For simplicity, this implementation doesn't support groups
	// A full implementation would need to track group names
	return h
}

// writeRow writes a single row with proper column formatting.
func (h *Handler) writeRow(buf *strings.Builder, key, value string) {
	// Handle multi-line values
	lines := strings.Split(value, "\n")

	for i, line := range lines {
		if i == 0 {
			// First line: include the key
			buf.WriteString(fmt.Sprintf("%*s : %s\n", h.leftWidth, key, h.wrapLine(line)))
		} else {
			// Subsequent lines: indent to align with right column
			wrapped := h.wrapLine(line)
			buf.WriteString(fmt.Sprintf("%*s   %s\n", h.leftWidth, "", wrapped))
		}
	}
}

// writeSeparator writes a separator line between records.
func (h *Handler) writeSeparator(buf *strings.Builder) {
	buf.WriteString(fmt.Sprintf("\n%*s ---\n\n", h.leftWidth, ""))
}

// shortPath extracts just the filename from a full path.
func (h *Handler) shortPath(fullPath string) string {
	return filepath.Base(fullPath)
}

// wrapLine wraps a single line if it exceeds rightWidth.
func (h *Handler) wrapLine(text string) string {
	if h.rightWidth <= 0 || len(text) <= h.rightWidth {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= h.rightWidth {
			currentLine += " " + word
		} else {
			result.WriteString(currentLine)
			result.WriteString("\n")
			result.WriteString(fmt.Sprintf("%*s   ", h.leftWidth, ""))
			currentLine = word
		}
	}

	result.WriteString(currentLine)
	return result.String()
}

// writeAttr writes an attribute, handling nested groups and different value types.
func (h *Handler) writeAttr(buf *strings.Builder, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}

	// Handle groups
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		for _, groupAttr := range attrs {
			h.writeAttr(buf, groupAttr)
		}
		return
	}

	// Format the value
	value := h.formatValue(a.Value)
	h.writeRow(buf, a.Key, value)
}

// formatValue formats a slog.Value into a string.
func (h *Handler) formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%g", v.Float64())
	case slog.KindBool:
		return fmt.Sprintf("%t", v.Bool())
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}
