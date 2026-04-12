package logger

import (
	"os"

	"github.com/ix-pay/ixpay-pro/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 定义日志接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	// With 返回一个带有上下文的新日志记录器
	With(fields ...interface{}) Logger
	// Sync 刷新所有缓冲区
	Sync() error
}

// LoggerType 日志类型
type LoggerType string

const (
	// DefaultLogger 默认应用日志
	DefaultLogger LoggerType = "default"
	// ErrorLogger 错误日志
	ErrorLogger LoggerType = "error"
	// TaskLogger 任务日志
	TaskLogger LoggerType = "task"
	// RequestLogger 请求日志
	RequestLogger LoggerType = "request"
	// AuditLogger 审计日志
	AuditLogger LoggerType = "audit"
)

// zapLogger 实现 Logger 接口
type zapLogger struct {
	logger      *zap.Logger
	errorLogger *zap.Logger // 错误日志专用记录器
}

// MultiLogger 多类型日志管理器
type MultiLogger struct {
	defaultLogger Logger
	errorLogger   Logger
	taskLogger    Logger
	requestLogger Logger
	auditLogger   Logger
}

// 全局日志管理器实例
var globalMultiLogger *MultiLogger

// SetGlobalMultiLogger 设置全局日志管理器
func SetGlobalMultiLogger(ml *MultiLogger) {
	globalMultiLogger = ml
}

// GetGlobalLogger 获取全局日志记录器
func GetGlobalLogger(loggerType LoggerType) Logger {
	if globalMultiLogger == nil {
		// 如果未初始化，返回默认日志记录器
		return nil
	}
	return globalMultiLogger.GetLogger(loggerType)
}

// GetLogger 获取指定类型的日志记录器
func (ml *MultiLogger) GetLogger(loggerType LoggerType) Logger {
	switch loggerType {
	case ErrorLogger:
		return ml.errorLogger
	case TaskLogger:
		return ml.taskLogger
	case RequestLogger:
		return ml.requestLogger
	case AuditLogger:
		return ml.auditLogger
	default:
		return ml.defaultLogger
	}
}

// SetupLogger 创建新的日志记录器（默认类型，带错误日志注入）
func SetupLogger(cfg *config.Config) Logger {
	defaultLogger := setupLoggerWithType(cfg, DefaultLogger)
	errorLogger := setupLoggerWithType(cfg, ErrorLogger)

	// 将 errorLogger 注入到 defaultLogger 中，使得调用 Error 时同时写入两个文件
	if defaultZapLogger, ok := defaultLogger.(*zapLogger); ok {
		if errorZapLogger, ok := errorLogger.(*zapLogger); ok {
			defaultZapLogger.errorLogger = errorZapLogger.logger
		}
	}

	return defaultLogger
}

// SetupMultiLogger 创建多类型日志记录器
func SetupMultiLogger(cfg *config.Config) *MultiLogger {
	return &MultiLogger{
		defaultLogger: setupLoggerWithType(cfg, DefaultLogger),
		errorLogger:   setupLoggerWithType(cfg, ErrorLogger),
		taskLogger:    setupLoggerWithType(cfg, TaskLogger),
		requestLogger: setupLoggerWithType(cfg, RequestLogger),
		auditLogger:   setupLoggerWithType(cfg, AuditLogger),
	}
}

// setupLoggerWithType 根据日志类型创建日志记录器
func setupLoggerWithType(cfg *config.Config, loggerType LoggerType) Logger {
	// 设置日志级别
	var level zapcore.Level
	switch cfg.Logging.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	// 错误日志始终记录 ERROR 及以上级别
	if loggerType == ErrorLogger {
		level = zap.ErrorLevel
	}

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Logging.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 根据日志类型选择文件路径
	var filePath string
	switch loggerType {
	case DefaultLogger:
		filePath = cfg.Logging.FilePath
	case ErrorLogger:
		filePath = cfg.Logging.ErrorFile
	case TaskLogger:
		filePath = cfg.Logging.TaskFile
	case RequestLogger:
		filePath = cfg.Logging.RequestFile
	case AuditLogger:
		filePath = cfg.Logging.AuditFile
	}

	// 创建文件写入器
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   true,
	})

	// 创建核心（默认日志同时输出到控制台和文件，其他类型只输出到文件）
	var core zapcore.Core
	if loggerType == DefaultLogger {
		core = zapcore.NewCore(
			encoder,
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter), zapcore.Lock(os.Stdout)),
			level,
		)
	} else {
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
	}

	// 创建日志记录器
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return &zapLogger{
		logger: logger,
	}
}

// Debug 记录调试信息
func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, parseFields(fields)...)
}

// Info 记录信息
func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, parseFields(fields)...)
}

// Warn 记录警告信息
func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, parseFields(fields)...)
}

// Error 记录错误信息
func (l *zapLogger) Error(msg string, fields ...interface{}) {
	if l.errorLogger != nil {
		l.errorLogger.Error(msg, parseFields(fields)...)
	} else {
		l.logger.Error(msg, parseFields(fields)...)
	}
}

// Fatal 记录致命错误信息并退出
func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	if l.errorLogger != nil {
		l.errorLogger.Fatal(msg, parseFields(fields)...)
	} else {
		l.logger.Fatal(msg, parseFields(fields)...)
	}
}

// With 返回一个带有上下文的新日志记录器
func (l *zapLogger) With(fields ...interface{}) Logger {
	return &zapLogger{
		logger: l.logger.With(parseFields(fields)...),
	}
}

// Sync 刷新所有缓冲区
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// parseFields 解析日志字段
func parseFields(fields []interface{}) []zap.Field {
	var zapFields []zap.Field

	for _, field := range fields {
		// 如果是zap.Field类型，直接添加
		if f, ok := field.(zap.Field); ok {
			zapFields = append(zapFields, f)
			continue
		}
	}

	// 处理键值对形式的字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}

		key, ok := fields[i].(string)
		if !ok {
			continue
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
