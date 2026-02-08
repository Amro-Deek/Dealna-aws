package middleware

import (
	"context"
	"log"
)

type StdLogger struct{}

func NewStdLogger() *StdLogger { return &StdLogger{} }

func (l *StdLogger) Debug(ctx context.Context, event string, fields map[string]any) {
	// keep debug simple; you can gate it with IsVerbose(ctx) if you want
	log.Printf("[DEBUG] %s %v", event, fields)
}

func (l *StdLogger) Info(ctx context.Context, event string, fields map[string]any) {
	log.Printf("[INFO] %s %v", event, fields)
}

func (l *StdLogger) Warn(ctx context.Context, event string, fields map[string]any) {
	log.Printf("[WARN] %s %v", event, fields)
}

func (l *StdLogger) Error(ctx context.Context, event string, fields map[string]any) {
	log.Printf("[ERROR] %s %v", event, fields)
}
