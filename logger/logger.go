package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kataras/golog"
)

func init() {
	golog.Levels[golog.FatalLevel].Title = "[FATAL]"
	golog.Levels[golog.ErrorLevel].Title = "[ERROR]"
	golog.Levels[golog.DebugLevel].Title = "[DEBUG]"
}

// Logger represents logger
type Logger interface {
	Error(v ...any)
	Errorf(format string, args ...any)

	Warn(v ...any)
	Warnf(format string, args ...any)

	Info(v ...any)
	Infof(format string, args ...any)

	Debug(v ...any)
	Debugf(format string, args ...any)
}

// Log represents Log implementation using golog.Logger
type Log struct {
	outputFile *os.File

	*golog.Logger
}

const logTimeFormat = "02.01.2006 15:04:05 MST"

// NewLog creates new Log from golog.Logger
func NewLog(log *golog.Logger) *Log {
	log.TimeFormat = logTimeFormat

	return &Log{
		Logger: log,
	}
}

const logFilePerm = 0o600

// SetOutputFile sets output file for logger
func (l *Log) SetOutputFile(filename string) error {
	file, err := os.OpenFile(filepath.Clean(filename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerm)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	l.outputFile = file
	l.SetOutput(l.outputFile)

	return nil
}

// Close closes logger if needed
func (l *Log) Close() error {
	if l.outputFile != nil {
		return l.outputFile.Close()
	}

	return nil
}
