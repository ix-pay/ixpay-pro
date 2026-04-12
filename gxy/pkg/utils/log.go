package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var logLevelNames = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

var logLevelColors = []string{
	"\033[36m", // Cyan
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[31m", // Red
	"\033[35m", // Magenta
}

var resetColor = "\033[0m"

type Logger struct {
	level    LogLevel
	logger   *log.Logger
	useColor bool
}

func NewLogger(level LogLevel, useColor bool) *Logger {
	return &Logger{
		level:    level,
		logger:   log.New(os.Stdout, "", 0),
		useColor: useColor,
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	levelName := logLevelNames[level]

	if l.useColor {
		levelName = fmt.Sprintf("%s%s%s", logLevelColors[level], levelName, resetColor)
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", now, levelName, msg)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}
