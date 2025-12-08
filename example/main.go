package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/henderjon/verticalslog"
)

func main() {
	// Create a new handler with custom options
	// AddSourceShort shows only the filename (e.g., "main.go:23")
	// AddSource would show the full path (e.g., "/Users/jon/code/example/main.go:23")
	handler := verticalslog.NewHandler(os.Stdout, &verticalslog.Options{
		LeftWidth:      15,
		RightWidth:     50,
		Level:          slog.LevelInfo,
		AddSourceShort: true,
	})

	logger := slog.New(handler)

	// Simple log message
	logger.Info("Application started successfully")

	// Log with various attributes
	logger.Info("User logged in",
		slog.String("username", "john.doe"),
		slog.String("email", "john.doe@example.com"),
		slog.Int("user_id", 12345),
		slog.Duration("session_timeout", 30*time.Minute),
	)

	// Log with a long message that will wrap
	logger.Warn("This is a very long warning message that will demonstrate the text wrapping functionality of the two-column handler when the text exceeds the maximum width")

	// Log with multi-line attribute value
	logger.Error("Database query failed",
		slog.String("query", "SELECT * FROM users\nWHERE status = 'active'\nAND created_at > '2024-01-01'"),
		slog.String("error", "connection timeout"),
		slog.Int("retry_count", 3),
	)

	// Log with a very long attribute value
	logger.Info("Configuration loaded",
		slog.String("path", "/etc/myapp/config.yaml"),
		slog.String("description", "This is a very long configuration description that should wrap nicely within the right column without breaking the column barrier and maintaining proper indentation throughout the entire text"),
		slog.Bool("valid", true),
	)

	// Log with nested information
	logger.Info("Request processed",
		slog.String("method", "POST"),
		slog.String("path", "/api/v1/users"),
		slog.Int("status_code", 201),
		slog.Duration("duration", 145*time.Millisecond),
		slog.String("user_agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"),
	)

	// Log with error details
	logger.Error("Failed to process payment",
		slog.String("transaction_id", "txn_1234567890"),
		slog.Float64("amount", 149.99),
		slog.String("currency", "USD"),
		slog.String("error_message", "Payment gateway returned an error:\nInsufficient funds in account.\nPlease ensure the account has adequate balance."),
		slog.String("customer_id", "cust_abcdef"),
	)
}
