package observability

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	mu     sync.Mutex
	level  Level
	out    io.Writer
	prefix string
}

var defaultLogger = NewLogger(LevelInfo, os.Stdout, "[SynX]")

func NewLogger(level Level, out io.Writer, prefix string) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{
		level:  level,
		out:    out,
		prefix: prefix,
	}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) logf(level Level, format string, args ...any) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("%s %s [%s] %s\n", timestamp, l.prefix, level.String(), msg)
	_, _ = l.out.Write([]byte(line))
}

func (l *Logger) Debug(format string, args ...any) { l.logf(LevelDebug, format, args...) }
func (l *Logger) Info(format string, args ...any)  { l.logf(LevelInfo, format, args...) }
func (l *Logger) Warn(format string, args ...any)  { l.logf(LevelWarn, format, args...) }
func (l *Logger) Error(format string, args ...any) { l.logf(LevelError, format, args...) }

// Package-level helpers
func Debug(format string, args ...any) { defaultLogger.Debug(format, args...) }
func Info(format string, args ...any)  { defaultLogger.Info(format, args...) }
func Warn(format string, args ...any)  { defaultLogger.Warn(format, args...) }
func Error(format string, args ...any) { defaultLogger.Error(format, args...) }

func init() {
	log.SetFlags(0)
}
