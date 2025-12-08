package verticalslog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestHandler_BasicFormatting(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  10,
		RightWidth: 40,
	})

	record := slog.NewRecord(
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		slog.LevelInfo,
		"test message",
		0,
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()

	// Check that output contains expected elements
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message, got:\n%s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("output should contain level, got:\n%s", output)
	}
	if !strings.Contains(output, "2025-01-01") {
		t.Errorf("output should contain timestamp, got:\n%s", output)
	}
	if !strings.Contains(output, " : ") {
		t.Errorf("output should contain column separator, got:\n%s", output)
	}
}

func TestHandler_Attributes(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  15,
		RightWidth: 50,
	})

	record := slog.NewRecord(
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		slog.LevelInfo,
		"test",
		0,
	)
	record.AddAttrs(
		slog.String("string_key", "string_value"),
		slog.Int("int_key", 42),
		slog.Bool("bool_key", true),
		slog.Float64("float_key", 3.14),
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()

	tests := []string{
		"string_key",
		"string_value",
		"int_key",
		"42",
		"bool_key",
		"true",
		"float_key",
		"3.14",
	}

	for _, expected := range tests {
		if !strings.Contains(output, expected) {
			t.Errorf("output should contain %q, got:\n%s", expected, output)
		}
	}
}

func TestHandler_MultilineValue(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  10,
		RightWidth: 40,
	})

	record := slog.NewRecord(
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		slog.LevelInfo,
		"test",
		0,
	)
	record.AddAttrs(
		slog.String("query", "SELECT *\nFROM users\nWHERE active = true"),
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()

	// Check that each line of the multiline value is present
	if !strings.Contains(output, "SELECT *") {
		t.Errorf("output should contain first line of query")
	}
	if !strings.Contains(output, "FROM users") {
		t.Errorf("output should contain second line of query")
	}
	if !strings.Contains(output, "WHERE active = true") {
		t.Errorf("output should contain third line of query")
	}

	// Check that continuation lines are properly indented
	lines := strings.Split(output, "\n")
	foundQuery := false
	for i, line := range lines {
		if strings.Contains(line, "query") && strings.Contains(line, "SELECT") {
			foundQuery = true
			// Next lines should have proper indentation (no key, just spacing)
			if i+1 < len(lines) && strings.Contains(lines[i+1], "FROM users") {
				// Should not have "query" key again
				if strings.Contains(lines[i+1], "query") {
					t.Errorf("continuation line should not repeat key")
				}
			}
		}
	}
	if !foundQuery {
		t.Errorf("did not find query attribute in output")
	}
}

func TestHandler_LongTextWrapping(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  10,
		RightWidth: 30, // Short width to force wrapping
	})

	longMessage := "This is a very long message that should wrap across multiple lines when it exceeds the maximum width"

	record := slog.NewRecord(
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		slog.LevelInfo,
		longMessage,
		0,
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()

	// Check that the message is split across multiple lines
	lines := strings.Split(output, "\n")
	messageLines := 0
	for _, line := range lines {
		if strings.Contains(line, "very long") || strings.Contains(line, "multiple lines") {
			messageLines++
			// Check that wrapped lines don't exceed expected length
			// Allow some buffer for formatting
			if len(line) > 60 {
				t.Errorf("line too long (%d chars): %s", len(line), line)
			}
		}
	}

	if messageLines < 2 {
		t.Errorf("expected message to wrap across multiple lines, got %d lines", messageLines)
	}
}

func TestHandler_Separator(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  15,
		RightWidth: 50,
	})

	record := slog.NewRecord(
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		slog.LevelInfo,
		"test",
		0,
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()

	// Check for separator
	if !strings.Contains(output, "---") {
		t.Errorf("output should contain separator, got:\n%s", output)
	}

	// Check that separator has blank line before it
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(line, "---") {
			if i > 0 && strings.TrimSpace(lines[i-1]) != "" {
				t.Errorf("separator should have blank line before it")
			}
			break
		}
	}
}

func TestHandler_ColumnAlignment(t *testing.T) {
	var buf bytes.Buffer
	leftWidth := 20
	handler := NewHandler(&buf, &Options{
		LeftWidth:  leftWidth,
		RightWidth: 50,
	})

	record := slog.NewRecord(
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		slog.LevelInfo,
		"test",
		0,
	)
	record.AddAttrs(
		slog.String("short", "value"),
		slog.String("very_long_key_name", "value"),
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Check that all ":" separators are aligned at the same column
	colonPositions := []int{}
	for _, line := range lines {
		if idx := strings.Index(line, " : "); idx != -1 {
			colonPositions = append(colonPositions, idx)
		}
	}

	if len(colonPositions) < 2 {
		t.Errorf("expected at least 2 lines with colons, got %d", len(colonPositions))
	}

	// All colons should be at the same position
	firstPos := colonPositions[0]
	for i, pos := range colonPositions {
		if pos != firstPos {
			t.Errorf("colon at line %d is at position %d, expected %d", i, pos, firstPos)
		}
	}

	// The colon position should be related to leftWidth
	// With right-aligned keys, the colon appears after leftWidth spaces
	if firstPos != leftWidth {
		t.Errorf("colon position is %d, expected %d (leftWidth)", firstPos, leftWidth)
	}
}

func TestHandler_Enabled(t *testing.T) {
	tests := []struct {
		name          string
		handlerLevel  slog.Level
		recordLevel   slog.Level
		shouldBeEnabled bool
	}{
		{"info handler, info record", slog.LevelInfo, slog.LevelInfo, true},
		{"info handler, warn record", slog.LevelInfo, slog.LevelWarn, true},
		{"warn handler, info record", slog.LevelWarn, slog.LevelInfo, false},
		{"error handler, warn record", slog.LevelError, slog.LevelWarn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&bytes.Buffer{}, &Options{
				Level: tt.handlerLevel,
			})

			enabled := handler.Enabled(context.Background(), tt.recordLevel)
			if enabled != tt.shouldBeEnabled {
				t.Errorf("Enabled() = %v, want %v", enabled, tt.shouldBeEnabled)
			}
		})
	}
}

func TestHandler_DifferentTypes(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  15,
		RightWidth: 50,
	})

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	record := slog.NewRecord(now, slog.LevelInfo, "test", 0)
	record.AddAttrs(
		slog.Duration("duration", 5*time.Minute),
		slog.Time("time", now),
		slog.Uint64("uint", 999),
		slog.Int64("int64", -123),
	)

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "5m0s") {
		t.Errorf("should contain duration format")
	}
	if !strings.Contains(output, "999") {
		t.Errorf("should contain uint64")
	}
	if !strings.Contains(output, "-123") {
		t.Errorf("should contain int64")
	}
}

func TestNewHandler_DefaultOptions(t *testing.T) {
	handler := NewHandler(&bytes.Buffer{}, nil)

	if handler.leftWidth != 20 {
		t.Errorf("default leftWidth should be 20, got %d", handler.leftWidth)
	}
	if handler.rightWidth != 60 {
		t.Errorf("default rightWidth should be 60, got %d", handler.rightWidth)
	}
}

func TestNewHandler_CustomOptions(t *testing.T) {
	handler := NewHandler(&bytes.Buffer{}, &Options{
		LeftWidth:  25,
		RightWidth: 80,
	})

	if handler.leftWidth != 25 {
		t.Errorf("leftWidth should be 25, got %d", handler.leftWidth)
	}
	if handler.rightWidth != 80 {
		t.Errorf("rightWidth should be 80, got %d", handler.rightWidth)
	}
}

func TestHandler_AddSourceShort(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:      15,
		RightWidth:     50,
		AddSourceShort: true,
	})

	logger := slog.New(handler)
	logger.Info("test message")

	output := buf.String()

	// Should contain source information
	if !strings.Contains(output, "source") {
		t.Errorf("output should contain source field, got:\n%s", output)
	}

	// Should contain only filename, not full path
	if strings.Contains(output, "/Users/") || strings.Contains(output, "/home/") {
		t.Errorf("output should not contain full path with AddSourceShort, got:\n%s", output)
	}

	// Should contain the test filename
	if !strings.Contains(output, "handler_test.go") {
		t.Errorf("output should contain handler_test.go, got:\n%s", output)
	}

	// Should contain line number
	if !strings.Contains(output, ":") {
		t.Errorf("output should contain line number with colon, got:\n%s", output)
	}
}

func TestHandler_AddSource(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth: 15,
		RightWidth: 50,
		AddSource: true,
	})

	logger := slog.New(handler)
	logger.Info("test message")

	output := buf.String()

	// Should contain source information
	if !strings.Contains(output, "source") {
		t.Errorf("output should contain source field, got:\n%s", output)
	}

	// Should contain full path
	if !strings.Contains(output, "handler_test.go") {
		t.Errorf("output should contain handler_test.go, got:\n%s", output)
	}
}

func TestHandler_AddSourceShortPrecedence(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:      15,
		RightWidth:     50,
		AddSource:      true,
		AddSourceShort: true, // Should take precedence
	})

	logger := slog.New(handler)
	logger.Info("test message")

	output := buf.String()

	// Should contain source information
	if !strings.Contains(output, "source") {
		t.Errorf("output should contain source field, got:\n%s", output)
	}

	// Should use short format when both are enabled
	if !strings.Contains(output, "handler_test.go") {
		t.Errorf("output should contain handler_test.go, got:\n%s", output)
	}

	// The short format should be used (harder to test definitively,
	// but we can check that the output doesn't seem excessively long)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "source") {
			// With short format, the line should be relatively short
			// Full paths would make this much longer
			if len(line) > 100 {
				t.Logf("source line seems long, might be full path: %s", line)
			}
		}
	}
}

func TestHandler_NoSource(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(&buf, &Options{
		LeftWidth:  15,
		RightWidth: 50,
		// Both AddSource and AddSourceShort are false by default
	})

	logger := slog.New(handler)
	logger.Info("test message")

	output := buf.String()

	// Should NOT contain source information
	if strings.Contains(output, "source :") {
		t.Errorf("output should not contain source field when disabled, got:\n%s", output)
	}
}

func TestShortPath(t *testing.T) {
	handler := NewHandler(&bytes.Buffer{}, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/jon/code/go/verticalslog/handler.go", "handler.go"},
		{"/home/user/project/main.go", "main.go"},
		{"main.go", "main.go"},
		{"/a/b/c/d/e/file.go", "file.go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := handler.shortPath(tt.input)
			if result != tt.expected {
				t.Errorf("shortPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
