package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LogEntry represents a single log line from a container.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Container string    `json:"container"`
	Stream    string    `json:"stream"` // "stdout", "stderr", "orchestrator"
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Logger provides thread-safe centralized logging for all containers.
// It writes to a single log file on disk and also to stdout.
type Logger struct {
	mu      sync.Mutex
	file    *os.File
	entries []LogEntry
}

// NewLogger creates a Logger that writes to the specified file path.
func NewLogger(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}
	return &Logger{
		file: f,
	}, nil
}

// Log writes a structured log entry for a container.
func (l *Logger) Log(container, level, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Container: container,
		Stream:    "orchestrator",
		Level:     level,
		Message:   strings.TrimSpace(message),
	}
	l.writeEntry(entry)
}

// writeEntry formats and writes a log entry to file and stdout.
func (l *Logger) writeEntry(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	line := fmt.Sprintf("[%s] [%s] [%s/%s] %s\n",
		entry.Timestamp.Format("2006-01-02 15:04:05.000"),
		strings.ToUpper(entry.Level),
		entry.Container,
		entry.Stream,
		entry.Message,
	)

	// Write to file.
	if l.file != nil {
		_, _ = l.file.WriteString(line)
	}

	// Write to stdout.
	fmt.Print(line)

	// Store in memory for API retrieval.
	l.entries = append(l.entries, entry)
}

// Writer returns an io.Writer that captures output from a container's
// stdout or stderr stream. Each line is logged as a structured entry.
func (l *Logger) Writer(container, stream string) io.Writer {
	return &logWriter{
		logger:    l,
		container: container,
		stream:    stream,
	}
}

// logWriter implements io.Writer and routes output through the centralized logger.
type logWriter struct {
	logger    *Logger
	container string
	stream    string
}

// Write implements io.Writer. It splits data by newlines and logs each line.
func (w *logWriter) Write(p []byte) (int, error) {
	data := string(p)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		entry := LogEntry{
			Timestamp: time.Now(),
			Container: w.container,
			Stream:    w.stream,
			Level:     w.streamLevel(),
			Message:   trimmed,
		}
		w.logger.writeEntry(entry)
	}
	return len(p), nil
}

// streamLevel returns a log level based on the stream type.
func (w *logWriter) streamLevel() string {
	if w.stream == "stderr" {
		return "error"
	}
	return "info"
}

// GetEntries returns a copy of all stored log entries.
func (l *Logger) GetEntries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Close flushes and closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
