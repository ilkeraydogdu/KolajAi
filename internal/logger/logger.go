package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// Global logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = New()
}

// New creates a new logger instance
func New() *Logger {
	level := INFO
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		switch strings.ToLower(env) {
		case "debug":
			level = DEBUG
		case "info":
			level = INFO
		case "warn", "warning":
			level = WARN
		case "error":
			level = ERROR
		case "fatal":
			level = FATAL
		}
	}

	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Debug logs debug messages
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logWithLevel("DEBUG", format, args...)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.logWithLevel("INFO", format, args...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.logWithLevel("WARN", format, args...)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.logWithLevel("ERROR", format, args...)
	}
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logWithLevel("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) logWithLevel(level, format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	fileShort := file[strings.LastIndex(file, "/")+1:]
	
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s:%d %s", level, fileShort, line, message)
}

// Global logging functions
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// ErrorWithStack logs error with stack trace
func ErrorWithStack(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	
	message := fmt.Sprintf(format, args...)
	defaultLogger.Error("%s: %v", message, err)
	
	// Add stack trace in debug mode
	if defaultLogger.level <= DEBUG {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		defaultLogger.Debug("Stack trace:\n%s", buf[:n])
	}
}