package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

var logLevelNames = []string{
	"TRACE",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

var logLevelColors = []string{
	"\033[34m", // Blue
	"\033[36m", // Cyan
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[31m", // Red
	"\033[35m", // Magenta
}

var resetColor = "\033[0m"

type Logger struct {
	level        LogLevel
	consoleLevel LogLevel
	logger       *log.Logger
	console      *log.Logger
	useColor     bool
	file         *os.File
	logPath      string
	currentDate  string
	module       string
	mu           sync.Mutex
	done         chan struct{}
}

func NewLogger(level LogLevel, useColor bool, module string) *Logger {
	return &Logger{
		level:        level,
		consoleLevel: level,
		logger:       log.New(os.Stdout, "", 0),
		console:      log.New(os.Stdout, "", 0),
		useColor:     useColor,
		module:       module,
	}
}

func NewLoggerWithFile(level LogLevel, useColor bool, logPath string, module string) (*Logger, error) {
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	ext := filepath.Ext(logPath)
	base := strings.TrimSuffix(logPath, ext)
	today := time.Now().Format("2006-01-02")
	todayPath := base + "-" + today + ext

	f, err := os.OpenFile(todayPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 文件只记录 INFO 及以上级别
	fileLevel := INFO
	if level < INFO {
		fileLevel = INFO
	}

	l := &Logger{
		level:        fileLevel, // 文件输出级别
		consoleLevel: level,     // 控制台输出级别
		logger:       log.New(f, "", 0),
		console:      log.New(os.Stdout, "", 0),
		useColor:     useColor,
		file:         f,
		logPath:      logPath,
		currentDate:  today,
		module:       module,
		done:         make(chan struct{}),
	}

	go l.rotateChecker()
	return l, nil
}

func (l *Logger) rotateChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.checkAndRotate()
		case <-l.done:
			return
		}
	}
}

func (l *Logger) checkAndRotate() {
	today := time.Now().Format("2006-01-02")
	if today == l.currentDate {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if today == l.currentDate {
		return
	}

	if l.file != nil {
		l.file.Close()
	}

	ext := filepath.Ext(l.logPath)
	base := strings.TrimSuffix(l.logPath, ext)
	todayPath := base + "-" + today + ext

	f, err := os.OpenFile(todayPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "切换日志文件失败: %v\n", err)
		return
	}

	l.file = f
	l.currentDate = today
	l.logger = log.New(f, "", 0)
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// 检查是否需要输出到控制台
	if level <= l.consoleLevel {
		l.logToConsole(level, format, args...)
	}

	// 检查是否需要输出到文件
	if level <= l.level {
		l.logToFile(level, format, args...)
	}
}

func (l *Logger) logToConsole(level LogLevel, format string, args ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")
	levelName := logLevelNames[level]

	if l.useColor {
		levelName = fmt.Sprintf("%s%s%s", logLevelColors[level], levelName, resetColor)
	}

	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	if l.module != "" {
		l.console.Printf("[%s] [%s] [%s] %s", now, levelName, l.module, msg)
	} else {
		l.console.Printf("[%s] [%s] %s", now, levelName, msg)
	}
	l.mu.Unlock()
}

func (l *Logger) logToFile(level LogLevel, format string, args ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")
	levelName := logLevelNames[level]

	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	if l.module != "" {
		l.logger.Printf("[%s] [%s] [%s] %s", now, levelName, l.module, msg)
	} else {
		l.logger.Printf("[%s] [%s] %s", now, levelName, msg)
	}
	l.mu.Unlock()
}

func (l *Logger) Trace(format string, args ...interface{}) {
	l.log(TRACE, format, args...)
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
	l.Close()
	os.Exit(1)
}

func (l *Logger) Close() {
	close(l.done)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
}
